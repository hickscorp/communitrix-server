package cbt

import "gogs.pierreqr.fr/doodloo/communitrix/i"
import "gogs.pierreqr.fr/doodloo/communitrix/logic"

func Wrap(sub interface{}) *Base {
	return &Base{Command: sub}
}

type Base struct{ Command interface{} }
type AddPlayer struct{ Player i.Player }
type RemovePlayer struct{ Player i.Player }

type Start struct{}

type PlayTurn struct {
	Player      i.Player
	UUID        string
	Rotation    *logic.Quaternion
	Translation *logic.Vector
}
