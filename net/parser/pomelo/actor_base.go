package pomelo

import (
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
)

type ActorBase struct {
	cactor.Base
}

func (p *ActorBase) Response(session *cproto.Session, v interface{}) {
	cactor.Response(p, session.AgentPath, session.Sid, session.Mid, v)
}

func (p *ActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	cactor.ResponseCode(p, session.AgentPath, session.Sid, session.Mid, statusCode)
}

func (p *ActorBase) ResponseCodeAndMessage(session *cproto.Session, statusCode int32, message string) {
	cactor.ResponseCodeAndMessage(p, session.AgentPath, session.Sid, session.Mid, statusCode, message)
}

func (p *ActorBase) ResponseError(session *cproto.Session, err error) {
	cactor.ResponseError(p, session.AgentPath, session.Sid, session.Mid, err)
}

func (p *ActorBase) Push(session *cproto.Session, route string, v interface{}) {
	cactor.Push(p, session.AgentPath, session.Sid, route, v)
}

func (p *ActorBase) Kick(session *cproto.Session, reason interface{}, closed bool) {
	cactor.Kick(p, session.AgentPath, session.Sid, reason, closed)
}

func (p *ActorBase) Broadcast(agentPath string, uidList []int64, allUID bool, route string, v interface{}) {
	cactor.Broadcast(p, agentPath, uidList, allUID, route, v)
}
