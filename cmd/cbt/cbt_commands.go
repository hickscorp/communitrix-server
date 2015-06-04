package cbt

import "communitrix/i"

func Wrap(sub interface{}) *Base {
	return &Base{Command: sub}
}

type Base struct{ Command interface{} }
type AddPlayer struct{ Player i.Player }
type RemovePlayer struct{ Player i.Player }
