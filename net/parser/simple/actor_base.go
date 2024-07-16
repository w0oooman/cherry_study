package simple

import (
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
)

type ActorBase struct {
	cactor.Base
}

func (p *ActorBase) Response(session *cproto.Session, mid uint32, v interface{}) {
	cactor.Response(p, session.AgentPath, session.Sid, mid, v)
}
