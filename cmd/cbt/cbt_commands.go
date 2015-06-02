package cbt

func Wrap(sub interface{}) *Base {
	return &Base{
		Command: sub,
	}
}

type Base struct {
	Command interface{}
}
type AddClient struct{ Client interface{} }
type RemoveClient struct{ Client interface{} }
