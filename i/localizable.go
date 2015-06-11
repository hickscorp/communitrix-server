package i

type Localizable interface {
	GetX() int
	GetY() int
	GetZ() int
	SetX(int)
	SetY(int)
	SetZ(int)

	Volume() int
	Half()
	Inv()
	Abs()
	SameAs(other Localizable) bool
	Translate(offset Localizable)
	Rotate(q interface{})
}
