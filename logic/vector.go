package logic

import "gogs.pierreqr.fr/doodloo/communitrix/util"

// Vector represents a point inside a integer coordinate system space.
type Vector struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// FromMap replaces the contents of the current object's values by the ones in the given map. The current object is then returned for chaining.
func (this *Vector) FromMap(m *util.MapHelper) *Vector {
	this.X, this.Y, this.Z = m.Int("x"), m.Int("y"), m.Int("z")
	return this
}

// NewVectorFromMap instanciates a new object given a map.
func NewVectorFromMap(data *util.MapHelper) *Vector {
	return (&Vector{}).FromMap(data)
}

// Allows to deep-copy a vector.
func (this *Vector) Copy() *Vector {
	return &Vector{this.X, this.Y, this.Z}
}
func (this *Vector) CopyTo(v *Vector) *Vector {
	v.X, v.Y, v.Z = this.X, this.Y, this.Z
	return v
}

// Translate applies a translation transformation to the current object. The current object is then returned for chaining.
func (this *Vector) Translate(t *Vector) *Vector {
	this.X += t.X
	this.Y += t.Y
	this.Z += t.Z
	return this
}

// Rotate applies a rotation transformation to the current object. The current object is then returned for chaining.
func (this *Vector) Rotate(q *Quaternion) *Vector {
	px, py, pz := float64(this.X), float64(this.Y), float64(this.Z)
	x, y, z, w := q.X, q.Y, q.Z, q.W
	this.X, this.Y, this.Z =
		util.QuickIntRound((w*w*px)+(2*y*w*pz)-(2*z*w*py)+(x*x*px)+(2*y*x*py)+(2*z*x*pz)-(z*z*px)-(y*y*px)),
		util.QuickIntRound((2*x*y*px)+(y*y*py)+(2*z*y*pz)+(2*w*z*px)-(z*z*py)+(w*w*py)-(2*x*w*pz)-(x*x*py)),
		util.QuickIntRound((2*x*z*px)+(2*y*z*py)+(z*z*pz)-(2*w*y*px)-(y*y*pz)+(2*w*x*py)-(x*x*pz)+(w*w*pz))
	return this
}
