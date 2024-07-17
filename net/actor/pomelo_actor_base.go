package cherryActor

import (
	cproto "github.com/cherry-game/cherry/net/proto"
)

type PomeloActorBase struct {
	Base
}

func (p *PomeloActorBase) Response(session *cproto.Session, v interface{}) {
	Response(p, session.AgentPath, session.Sid, session.Mid, v)
}

func (p *PomeloActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	ResponseCode(p, session.AgentPath, session.Sid, session.Mid, statusCode)
}

func (p *PomeloActorBase) ResponseCodeAndMessage(session *cproto.Session, statusCode int32, message string) {
	ResponseCodeAndMessage(p, session.AgentPath, session.Sid, session.Mid, statusCode, message)
}

func (p *PomeloActorBase) ResponseError(session *cproto.Session, err error) {
	ResponseError(p, session.AgentPath, session.Sid, session.Mid, err)
}

func (p *PomeloActorBase) Push(session *cproto.Session, route string, v interface{}) {
	Push(p, session.AgentPath, session.Sid, route, v)
}

func (p *PomeloActorBase) Kick(session *cproto.Session, reason interface{}, closed bool) {
	Kick(p, session.AgentPath, session.Sid, reason, closed)
}

func (p *PomeloActorBase) Broadcast(agentPath string, uidList []int64, allUID bool, route string, v interface{}) {
	Broadcast(p, agentPath, uidList, allUID, route, v)
}
