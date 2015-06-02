package tx

import (
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
	UUID       string    `json:"uuid"`       // The combat unique identifier on the server.
	MinPlayers int       `json:"minPlayers"` // The minimum number of players that can join.
	MaxPlayers int       `json:"maxPlayers"` // The maximum number of players that can join.
	Players    *[]string `json:"players"`
}
type CombatPlayerJoined struct {
	Player string `json:"player"`
}
type CombatPlayerLeft struct {
	Player string `json:"player"`
}
type CombatStart struct {
	UUID    string    `json:"uuid"`    // The combat unique identifier on the server.
	Players *[]string `json:"players"` // The list of players.
}
type CombatNewTurn struct{}
type CombatPlayerTurn struct {
	UUID   string `json:"uuid"`
	Player string `json:"player"`
}
type CombatEnd struct{}
