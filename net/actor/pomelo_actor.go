package cherryActor

import (
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"net"
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pomeloMessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	ppacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nuid"
	"go.uber.org/zap/zapcore"
)

type (
	pomeloActor struct {
		Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewPomeloAgentFunc
		onInitFunc     func()
	}

	OnNewPomeloAgentFunc func(newAgent *pomelo.Agent)
)

func NewPomeloActor(agentActorID string) *pomeloActor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &pomeloActor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
		onInitFunc:   nil,
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *pomeloActor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
	p.Remote().Register(PushFuncName, p.push)
	p.Remote().Register(KickFuncName, p.kick)
	p.Remote().Register(BroadcastName, p.broadcast)

	if p.onInitFunc != nil {
		p.onInitFunc()
	}
}

func (p *pomeloActor) SetOnInitFunc(fn func()) {
	p.onInitFunc = fn
}

func (p *pomeloActor) Load(app cfacade.IApplication) {
	if len(p.connectors) < 1 {
		panic("connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	pomelo.Cmd().Init(app)

	//  Create agent actor
	if _, err := app.ActorSystem().CreateActor(p.agentActorID, p); err != nil {
		clog.Panicf("Create agent actor fail. err = %+v", err)
	}

	for _, connector := range p.connectors {
		connector.OnConnect(p.defaultOnConnectFunc)
		go connector.Start() // start connector!
	}
}

func (p *pomeloActor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

func (p *pomeloActor) Connectors() []cfacade.IConnector {
	return p.connectors
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *pomeloActor) defaultOnConnectFunc(conn net.Conn) {
	session := &cproto.Session{
		Sid:       nuid.Next(),
		AgentPath: p.Path().String(),
		Data:      map[string]string{},
	}

	agent := pomelo.NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(&agent)
	}

	pomelo.BindSID(&agent)
	agent.Run()
}

func (*pomeloActor) SetDictionary(dict map[string]uint16) {
	pomeloMessage.SetDictionary(dict)
}

func (*pomeloActor) SetDataCompression(compression bool) {
	pomeloMessage.SetDataCompression(compression)
}

func (*pomeloActor) SetWriteBacklog(size int) {
	pomelo.Cmd().SetWriteBacklog(size)
}

func (*pomeloActor) SetHeartbeat(t time.Duration) {
	if t.Seconds() < 1 {
		t = 60 * time.Second
	}
	pomelo.Cmd().SetHeartbeat(t)
}

func (*pomeloActor) SetSysData(key string, value interface{}) {
	pomelo.Cmd().SetSysData(key, value)
}

func (p *pomeloActor) SetOnNewAgent(fn OnNewPomeloAgentFunc) {
	p.onNewAgentFunc = fn
}

func (*pomeloActor) SetOnDataRoute(fn pomelo.DataRouteFunc) {
	pomelo.Cmd().SetOnDataRoute(fn)
}

func (*pomeloActor) SetOnPacket(typ ppacket.Type, fn pomelo.PacketFunc) {
	pomelo.Cmd().SetOnPacket(typ, fn)
}

func (p *pomeloActor) response(rsp *cproto.PomeloResponse) {
	agent, found := pomelo.GetAgent(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	if ccode.IsOK(rsp.Code) {
		agent.ResponseMID(rsp.Mid, rsp.Data, false)
	} else {
		errRsp := &cproto.Response{
			Code:    rsp.Code,
			Message: rsp.Message,
		}
		agent.ResponseMID(rsp.Mid, errRsp, true)
	}
}

func (p *pomeloActor) push(rsp *cproto.PomeloPush) {
	agent, found := pomelo.GetAgent(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[push] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Push(rsp.Route, rsp.Data)
}

func (p *pomeloActor) kick(rsp *cproto.PomeloKick) {
	agent, found := pomelo.GetAgentWithUID(rsp.Uid)
	if !found {
		agent, found = pomelo.GetAgent(rsp.Sid)
	}

	if found {
		agent.Kick(rsp.Reason, rsp.Close)
	}
}

func (p *pomeloActor) broadcast(rsp *cproto.PomeloBroadcastPush) {
	if rsp.AllUID {
		pomelo.ForeachAgent(func(agent *pomelo.Agent) {
			if agent.IsBind() {
				agent.Push(rsp.Route, rsp.Data)
			}
		})
	} else {
		for _, uid := range rsp.UidList {
			if agent, found := pomelo.GetAgentWithUID(uid); found {
				agent.Push(rsp.Route, rsp.Data)
			}
		}
	}
}
