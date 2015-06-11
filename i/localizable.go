package i

type Localizable interface {
	GetX() int
	GetY() int
	GetZ() int
	SetX(int)
	SetY(int)
	SetZ(int)
	Volume() int                   // X * Y * Z
	Half()                         // X / 2, Y / 2, Z / 2
	Inv()                          // -X, -Y, -Z
	Abs()                          // SQRT(X*X), SQRT(Y*Y), SQRT(Z*Z)
	SameAs(other Localizable) bool // X==o.X && Y==o.Y && Z==o.Z
	Translate(offset Localizable)  // X+o.X, Y+o.Y, Z+o.Z
	Rotate(q interface{})          // Applies Q
}
