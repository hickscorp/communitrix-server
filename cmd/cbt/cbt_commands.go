package cbt

import "github.com/hickscorp/communitrix-server/logic"
import "github.com/hickscorp/communitrix-server/util"

func Wrap(sub interface{}) *Base {
	return &Base{Command: sub}
}

type Base struct{ Command interface{} }
type AddPlayer struct{ Player interface{} }
type RemovePlayer struct{ Player interface{} }

type Summarize struct {
	Ret chan util.MapHelper
}
type Prepare struct{}
type Start struct {
	Target *logic.Piece
	Units  logic.Units
	Pieces logic.Pieces
}

type StartNewTurn struct{}
type PlayTurn struct {
	Player      interface{}
	PieceIndex  int
	Translation *logic.Vector
	Rotation    *logic.Quaternion
}
type Vote struct {
	Player   interface{}
	PlayerID string
}
