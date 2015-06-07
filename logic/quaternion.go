package logic

import "gogs.pierreqr.fr/doodloo/communitrix/util"

// Quaternion encapsulates rotation transforms.
type Quaternion struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

// NewVectorFromMap instanciates a new object given a map.
func NewQuaternionFromMap(data util.MapHelper) *Quaternion {
	return (&Quaternion{}).FromMap(data)
}

// FromMap replaces the contents of the current object's values by the ones in the given map. The current object is returned for chaining.
func (this *Quaternion) FromMap(m util.MapHelper) *Quaternion {
	this.X, this.Y, this.Z, this.W = m.Float("x"), m.Float("y"), m.Float("z"), m.Float("w")
	return this
}

// Allows to deep-copy a vector.
func (this *Quaternion) Copy() *Quaternion {
	return &Quaternion{this.X, this.Y, this.Z, this.W}
}
func (this *Quaternion) CopyTo(q *Quaternion) *Quaternion {
	q.X, q.Y, q.Z, q.W = this.X, this.Y, this.Z, this.W
	return q
}
