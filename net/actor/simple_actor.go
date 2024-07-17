package cherryActor

import (
	"encoding/binary"
	"github.com/cherry-game/cherry/net/parser/simple"
	"net"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nuid"
	"go.uber.org/zap/zapcore"
)

type (
	simpleActor struct {
		Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewSimpleAgentFunc
	}

	OnNewSimpleAgentFunc func(newAgent *simple.Agent)
)

func NewSimpleActor(agentActorID string) *simpleActor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &simpleActor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *simpleActor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
}

func (p *simpleActor) Load(app cfacade.IApplication) {
	if len(p.connectors) < 1 {
		panic("Connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	//  Create agent actor
	if _, err := app.ActorSystem().CreateActor(p.agentActorID, p); err != nil {
		clog.Panicf("Create agent actor fail. err = %+v", err)
	}

	for _, connector := range p.connectors {
		connector.OnConnect(p.defaultOnConnectFunc)
		go connector.Start() // start connector!
	}
}

func (p *simpleActor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

func (p *simpleActor) Connectors() []cfacade.IConnector {
	return p.connectors
}

func (p *simpleActor) AddNodeRoute(mid uint32, nodeRoute *simple.NodeRoute) {
	simple.AddNodeRoute(mid, nodeRoute)
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *simpleActor) defaultOnConnectFunc(conn net.Conn) {
	session := &cproto.Session{
		Sid:       nuid.Next(),
		AgentPath: p.Path().String(),
		Data:      map[string]string{},
	}

	agent := simple.NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(&agent)
	}

	simple.BindSID(&agent)
	agent.Run()
}

func (p *simpleActor) SetOnNewAgent(fn OnNewSimpleAgentFunc) {
	p.onNewAgentFunc = fn
}

func (p *simpleActor) SetHeartbeatTime(t time.Duration) {
	simple.SetHeartbeatTime(t)
}

func (p *simpleActor) SetWriteBacklog(backlog int) {
	simple.SetWriteBacklog(backlog)
}

func (p *simpleActor) SetEndian(e binary.ByteOrder) {
	simple.SetEndian(e)
}

func (*simpleActor) SetOnDataRoute(fn simple.DataRouteFunc) {
	if fn != nil {
		simple.OnDataRouteFunc = fn
	}
}

func (p *simpleActor) response(rsp *cproto.PomeloResponse) {
	agent, found := simple.GetAgent(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Response(rsp.Mid, rsp.Data)
}
