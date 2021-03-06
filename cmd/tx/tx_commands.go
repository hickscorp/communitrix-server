package tx

import (
	"github.com/hickscorp/communitrix-server/util"
	"reflect"
)

func Wrap(sub interface{}) *Base {
	return &Base{reflect.TypeOf(sub).Name(), sub}
}

type Base struct {
	Type    string      `json:"type"`    // Will hold the name of the command.
	Command interface{} `json:"command"` // The real command.
}

type Error struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}
type Acknowledgment struct {
	Serial       string `json:"serial"`
	Valid        bool   `json:"valid"`
	ErrorMessage string `json:"errorMessage"`
}
type Welcome struct {
	Message string `json:"message"`
}
type Registered struct {
	Username string `json:"username"`
}

type CombatList struct {
	Combats []util.MapHelper `json:"combats"`
}
type CombatJoin struct {
	Combat interface{} `json:"combat"` // The combat details.
}
type CombatPlayerJoined struct {
	Player util.MapHelper `json:"player"`
}
type CombatPlayerLeft struct {
	UUID string `json:"uuid"`
}
type CombatStart struct {
	UUID   string      `json:"uuid"` // The combat unique identifier on the server.
	Target interface{} `json:"target"`
	Units  interface{} `json:"units"`
	Pieces interface{} `json:"pieces"`
}
type CombatNewTurn struct {
	TurnID int `json:"turnId"`
	UnitID int `json:"unitId"`
}
type CombatPlayerTurn struct {
	PlayerUUID string      `json:"playerUUID"`
	PieceID    int         `json:"pieceId"`
	UnitID     int         `json:"unitId"`
	Unit       interface{} `json:"unit"`
}
type CombatEnd struct {
}
