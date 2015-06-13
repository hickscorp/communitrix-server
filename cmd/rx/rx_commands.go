package rx

import "gogs.pierreqr.fr/doodloo/communitrix/i"

func Wrap(player i.Player, sub interface{}) *Base {
	return &Base{
		Player:  player,
		Command: sub,
	}
}

type Base struct {
	Player  i.Player
	Command interface{}
}

type Register struct {
	Username string
}
type Unregister struct{}

type CombatList struct{}
type CombatCreate struct {
	MinPlayers int
	MaxPlayers int
}
type CombatJoin struct {
	UUID string
}
type CombatEnd struct {
	UUID string
}
