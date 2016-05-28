package logic

import (
	"github.com/hickscorp/communitrix-server/util"
	"math"
)

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

// Usable even if the quaternion is not normalized
func (this *Quaternion) ToEulerAngles() *Vector {
	sqw := this.W * this.W
	sqx := this.X * this.X
	sqy := this.Y * this.Y
	sqz := this.Z * this.Z

	// If normalised it is one, otherwise it is the correction factor.
	unit := sqx + sqy + sqz + sqw
	test := this.X*this.Y + this.Z*this.W

	// Rotation around Y axis, around Z axis and around  around X axis, RESPECTIVELY
	var heading, attitude, bank float64

	if test > 0.499*unit { // Singularity at north pole
		heading = 2 * math.Atan2(this.X, this.W) * 180 / math.Pi
		attitude = math.Pi / 2 * 180 / math.Pi
		bank = 0
	} else if test < -0.499*unit { // Singularity at south pole
		heading = -2 * math.Atan2(this.X, this.W) * 180 / math.Pi
		attitude = -math.Pi / 2 * 180 / math.Pi
		bank = 0
	} else {
		heading = math.Atan2(2*this.Y*this.W-2*this.X*this.Z, 1-2*sqy-2*sqz) * 180 / math.Pi
		attitude = math.Asin(2*test) * 180 / math.Pi
		bank = math.Atan2(2*this.X*this.W-2*this.Y*this.Z, 1-2*sqx-2*sqz) * 180 / math.Pi
	}
	return (&Vector{}).FromValues(util.QuickIntRound(bank), util.QuickIntRound(heading), util.QuickIntRound(attitude))
}
