package gameError

import cherryError "github.com/cherry-game/cherry/error"

var (
	ErrRoleNameIsNil               = cherryError.Error("role name is null")
	ErrRoleInitConfigIsNonExistent = cherryError.Error("role init config is non-existent")
)
