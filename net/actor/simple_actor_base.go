package cherryActor

import (
	cproto "github.com/cherry-game/cherry/net/proto"
)

type SimpleActorBase struct {
	Base
}

func (p *SimpleActorBase) Response(session *cproto.Session, mid uint32, v interface{}) {
	Response(p, session.AgentPath, session.Sid, mid, v)
}
