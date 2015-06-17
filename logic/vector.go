package logic

import (
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math"
)

// Vector represents a point inside a integer coordinate system space.
type Vector struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func (this *Vector) GetX() int  { return this.X }
func (this *Vector) GetY() int  { return this.Y }
func (this *Vector) GetZ() int  { return this.Z }
func (this *Vector) SetX(x int) { this.X = x }
func (this *Vector) SetY(y int) { this.Y = y }
func (this *Vector) SetZ(z int) { this.Z = z }

// Instanciator based on an existing vector.
func NewVectorFromVector(other *Vector) *Vector {
	return (&Vector{}).FromVector(other)
}

// Updates values from another vector.
func (this *Vector) FromVector(other *Vector) *Vector {
	this.X, this.Y, this.Z = other.X, other.Y, other.Z
	return this
}

// Instanciator based on a list of integers.
func NewVectorFromValues(x, y, z int) *Vector {
	return (&Vector{}).FromValues(x, y, z)
}

// Updates values from a list of integers.
func (this *Vector) FromValues(x, y, z int) *Vector {
	this.X, this.Y, this.Z = x, y, z
	return this
}

// Instanciator based on a map.
func NewVectorFromMap(data util.MapHelper) *Vector {
	ret := Vector{}
	ret.FromMap(data)
	return &ret
}

// Updates values from a map.
func (this *Vector) FromMap(m util.MapHelper) {
	this.X, this.Y, this.Z = m.Int("x"), m.Int("y"), m.Int("z")
}

// CollidesWith checks wether two vectors are holding identical values.
func (this *Vector) CollidesWith(other *Vector) bool {
	return this.X == other.GetX() && this.Y == other.GetY() && this.Z == other.GetZ()
}

// Clone allows to deep-copy a vector.
func (this *Vector) Clone() *Vector {
	return NewVectorFromVector(this)
}

func (this *Vector) Volume() int {
	return util.QuickIntRound(math.Abs(float64(this.X)) * math.Abs(float64(this.Y)) * math.Abs(float64(this.Z)))
}
func (this *Vector) Half() *Vector {
	this.X, this.Y, this.Z = this.X/2, this.Y/2, this.Z/2
	return this
}
func (this *Vector) Inv() *Vector {
	this.X, this.Y, this.Z = -this.X, -this.Y, -this.Z
	return this
}
func (this *Vector) Abs() *Vector {
	this.X, this.Y, this.Z = util.QuickIntRound(math.Abs(float64(this.X))), util.QuickIntRound(math.Abs(float64(this.Y))), util.QuickIntRound(math.Abs(float64(this.Z)))
	return this
}

// Translate applies a translation transformation to the current object. The current object is then returned for chaining.
func (this *Vector) Translate(t *Vector) *Vector {
	this.X, this.Y, this.Z = this.X+t.GetX(), this.Y+t.GetY(), this.Z+t.GetZ()
	return this
}

// Rotate applies a rotation transformation to the current object. The current object is then returned for chaining.
func (this *Vector) Rotate(iq interface{}) *Vector {
	q := iq.(*Quaternion)
	px, py, pz := float64(this.X), float64(this.Y), float64(this.Z)
	x, y, z, w := q.X, q.Y, q.Z, q.W
	this.X = util.QuickIntRound((w * w * px) + (2 * y * w * pz) - (2 * z * w * py) + (x * x * px) + (2 * y * x * py) + (2 * z * x * pz) - (z * z * px) - (y * y * px))
	this.Y = util.QuickIntRound((2 * x * y * px) + (y * y * py) + (2 * z * y * pz) + (2 * w * z * px) - (z * z * py) + (w * w * py) - (2 * x * w * pz) - (x * x * py))
	this.Z = util.QuickIntRound((2 * x * z * px) + (2 * y * z * py) + (z * z * pz) - (2 * w * y * px) - (y * y * pz) + (2 * w * x * py) - (x * x * pz) + (w * w * pz))
	return this
}
