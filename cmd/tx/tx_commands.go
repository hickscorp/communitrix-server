package tx

import (
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"reflect"
)

func Wrap(sub interface{}) Base {
	return Base{reflect.TypeOf(sub).Name(), sub}
}

type Base struct {
	Type    string      `json:"type"`    // Will hold the name of the command.
	Command interface{} `json:"command"` // The real command.
}

type Error struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
}
type Welcome struct {
	Message string `json:"message"`
}
type Registered struct {
	Username string `json:"username"`
}

type CombatList struct {
	Combats *[]string `json:"combats"`
}
type CombatJoin struct {
	Combat interface{} `json:"combat"` // The combat details.
}
type CombatPlayerJoined struct {
	Player *util.MapHelper `json:"player"`
}
type CombatPlayerLeft struct {
	UUID string `json:"uuid"`
}
type CombatStart struct {
	UUID   string      `json:"uuid"` // The combat unique identifier on the server.
	Target interface{} `json:"target"`
	Pieces interface{} `json:"pieces"`
}
type CombatNewTurn struct{}
type CombatPlayerTurn struct {
	PlayerUUID string      `json:"playerUUID"`
	Contents   interface{} `json:"contents"`
}
type CombatEnd struct{}
