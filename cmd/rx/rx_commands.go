package rx

func Wrap(client interface{}, sub interface{}) *Base {
	return &Base{
		Client:  client,
		Command: sub,
	}
}

type Base struct {
	Client  interface{} // Pointer to a Client.
	Command interface{} // Command.
}
type Error struct {
	Code   int
	Reason string
}
type Register struct{}
type Unregister struct{}

type CombatList struct{}
type CombatCreate struct {
	MinPlayers int
	MaxPlayers int
}
type CombatJoin struct {
	UUID string
}
type CombatLeave struct{}
type CombatPlayTurn struct{}
