package cherryActor

import (
	cherryError "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

const (
	ResponseFuncName = "response"
	PushFuncName     = "push"
	KickFuncName     = "kick"
	BroadcastName    = "broadcast"
)

func Response(iActor cfacade.IActor, agentPath, sid string, mid uint32, v interface{}) {
	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Response] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Data: data,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

func ResponseCode(iActor cfacade.IActor, agentPath, sid string, mid uint32, statusCode int32) {
	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Code: statusCode,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

func ResponseCodeAndMessage(iActor cfacade.IActor, agentPath, sid string, mid uint32, statusCode int32, message string) {
	rsp := &cproto.PomeloResponse{
		Sid:     sid,
		Mid:     mid,
		Code:    statusCode,
		Message: message,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

func ResponseError(iActor cfacade.IActor, agentPath, sid string, mid uint32, err error) {
	baseError, ok := err.(*cherryError.BaseError)
	if !ok {
		clog.Errorf("[ResponseError] err type is not *cherryError.BaseError, err=%s", err.Error())
		return
	}
	ResponseCodeAndMessage(iActor, agentPath, sid, mid, baseError.ErrorCode(), baseError.Error())
}

func Push(iActor cfacade.IActor, agentPath, sid, route string, v interface{}) {
	if route == "" {
		clog.Warn("[Push] route value error.")
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Push] Marshal error. route =%s, v = %+v", route, v)
		return
	}

	rsp := &cproto.PomeloPush{
		Sid:   sid,
		Route: route,
		Data:  data,
	}

	iActor.Call(agentPath, PushFuncName, rsp)
}

func Kick(iActor cfacade.IActor, agentPath, sid string, reason interface{}, closed bool) {
	data, err := iActor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. reason = %+v", reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Sid:    sid,
		Reason: data,
		Close:  closed,
	}

	iActor.Call(agentPath, KickFuncName, rsp)
}

func Broadcast(iActor cfacade.IActor, agentPath string, uidList []int64, allUID bool, route string, v interface{}) {
	if !allUID && len(uidList) < 1 {
		clog.Warn("[Broadcast] uidList value error.")
		return
	}

	if route == "" {
		clog.Warn("[Broadcast] route value error.")
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloBroadcastPush{
		UidList: uidList,
		AllUID:  allUID,
		Route:   route,
		Data:    data,
	}

	iActor.Call(agentPath, BroadcastName, rsp)
}
