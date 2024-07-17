package cherryActor

import (
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"google.golang.org/protobuf/proto"
	"reflect"

	ccode "github.com/cherry-game/cherry/code"
	cerror "github.com/cherry-game/cherry/error"
	cherryError "github.com/cherry-game/cherry/error"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

func PCall(method *creflect.FuncInfo, args []reflect.Value) (rets interface{}, err error, ok bool) {
	r := method.Value.Call(args)
	// r can have 0 length in case of notify handlers
	// otherwise it will have 2 outputs: an interface and an error
	if len(r) == 2 {
		ok = true
		if v := r[1].Interface(); v != nil {
			err = v.(error)
		} else if !r[0].IsNil() {
			rets = r[0].Interface()
		} else {
			err = cherryError.ErrReplyShouldBeNotNull
		}
	}
	return
}

func InvokeLocalFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message, actor cfacade.IActor) {
	if app == nil {
		clog.Errorf("[InvokeLocalFunc] app is nil. [message = %+v]", m)
		return
	}

	EncodeLocalArgs(app, fi, m)

	values := make([]reflect.Value, 2)
	values[0] = reflect.ValueOf(m.Session) // session
	values[1] = reflect.ValueOf(m.Args)    // args
	resp, err, ok := PCall(fi, values)
	if !ok {
		clog.Debugf("[InvokeLocalFunc]. function: %s is not standardization, target=%s",
			m.FuncName, m.Target)
		return
	}

	s := m.Session
	if err == nil || reflect.ValueOf(err).IsNil() {
		Response(actor, s.AgentPath, s.Sid, s.Mid, resp)
	} else {
		ResponseError(actor, s.AgentPath, s.Sid, s.Mid, err)
		clog.Warnf("[InvokeLocalFunc] err:%s,target=%s,funcName=%s",
			err.Error(), m.Target, m.FuncName)
	}
}

func AgentInvokeLocalFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message, actor cfacade.IActor) {
	if app == nil {
		clog.Errorf("[AgentInvokeLocalFunc] app is nil. [message = %+v]", m)
		return
	}

	EncodeLocalArgs(app, fi, m)

	values := make([]reflect.Value, 2)
	values[0] = reflect.ValueOf(m.Session) // session
	values[1] = reflect.ValueOf(m.Args)    // args
	resp, err, ok := PCall(fi, values)
	if !ok {
		clog.Debugf("[AgentInvokeLocalFunc]. function: %s is not standardization, target=%s",
			m.FuncName, m.Target)
		return
	}

	s := m.Session
	agent, ok := pomelo.GetAgent(s.Sid)
	if !ok {
		clog.Debugf("[AgentInvokeLocalFunc]. agent is not found, sid=%s, uid=%d, target=%s",
			s.Sid, s.Uid, m.FuncName, m.Target)
		return
	}

	if err == nil || reflect.ValueOf(err).IsNil() {
		agent.Response(s, resp)
	} else {
		agent.ResponseError(s, err)
		clog.Warnf("[InvokeLocalFunc] err:%s,target=%s,funcName=%s",
			err.Error(), m.Target, m.FuncName)
	}
}

func InvokeRemoteFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message, actor cfacade.IActor) {
	if app == nil {
		clog.Errorf("[InvokeRemoteFunc] app is nil. [message = %+v]", m)
		return
	}

	EncodeRemoteArgs(app, fi, m)

	values := make([]reflect.Value, fi.InArgsLen)
	if fi.InArgsLen > 0 {
		values[0] = reflect.ValueOf(m.Args) // args
	}

	if m.IsCluster {
		cutils.Try(func() {
			rets := fi.Value.Call(values)
			rspCode, rspData := retValue(app.Serializer(), rets)

			retResponse(m.ClusterReply, &cproto.Response{
				Code: rspCode,
				Data: rspData,
			})

		}, func(errString string) {
			retResponse(m.ClusterReply, &cproto.Response{
				Code: ccode.RPCRemoteExecuteError,
			})
			clog.Errorf("[InvokeRemoteFunc] invoke error. [message = %+v, err = %s]", m, errString)
		})
	} else {
		cutils.Try(func() {
			if m.ChanResult == nil {
				fi.Value.Call(values)
			} else {
				rets := fi.Value.Call(values)
				rspCode, rspData := retValue(app.Serializer(), rets)
				m.ChanResult <- &cproto.Response{
					Code: rspCode,
					Data: rspData,
				}
			}
		}, func(errString string) {
			if m.ChanResult != nil {
				m.ChanResult <- nil
			}

			clog.Errorf("[remote] invoke error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
				m.Source,
				m.Target,
				m.FuncName,
				fi.InArgs,
				errString,
			)
		})
	}
}

func EncodeRemoteArgs(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) error {
	if m.IsCluster {
		if fi.InArgsLen == 0 {
			return nil
		}

		return EncodeArgs(app, fi, 0, m)
	}

	return nil
}

func EncodeLocalArgs(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) error {
	return EncodeArgs(app, fi, 1, m)
}

func EncodeArgs(app cfacade.IApplication, fi *creflect.FuncInfo, index int, m *cfacade.Message) error {
	argBytes, ok := m.Args.([]byte)
	if !ok {
		return cerror.Errorf("Encode args error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
	}

	argValue := reflect.New(fi.InArgs[index].Elem()).Interface()
	err := app.Serializer().Unmarshal(argBytes, argValue)
	if err != nil {
		return cerror.Errorf("Encode args unmarshal error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
	}

	m.Args = argValue

	return nil
}

func retValue(serializer cfacade.ISerializer, rets []reflect.Value) (int32, []byte) {
	var (
		retsLen = len(rets)
		rspCode = ccode.OK
		rspData []byte
	)

	if retsLen == 1 {
		if val := rets[0].Interface(); val != nil {
			if c, ok := val.(int32); ok {
				rspCode = c
			}
		}
	} else if retsLen == 2 {
		if !rets[0].IsNil() {
			data, err := serializer.Marshal(rets[0].Interface())
			if err != nil {
				rspCode = ccode.RPCRemoteExecuteError
				clog.Warn(err)
			} else {
				rspData = data
			}
		}

		if val := rets[1].Interface(); val != nil {
			if c, ok := val.(int32); ok {
				rspCode = c
			}
		}
	}

	return rspCode, rspData
}

func retResponse(reply cfacade.IRespond, rsp *cproto.Response) {
	if reply != nil {
		rspData, _ := proto.Marshal(rsp)
		err := reply.Respond(rspData)
		if err != nil {
			clog.Warn(err)
		}
	}
}
