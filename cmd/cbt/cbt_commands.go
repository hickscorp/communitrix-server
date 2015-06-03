package cbt

func Wrap(sub interface{}) *Base {
	return &Base{
		Command: sub,
	}
}

type Base struct {
	Command interface{}
}
type AddPlayer struct{ Player interface{} }
type RemovePlayer struct{ Player interface{} }
