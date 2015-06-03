package logic

import (
	"math"
)

type Vector struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func NewVectorFromMap(data interface{}) *Vector {
	m := data.(map[string]interface{})
	return &Vector{int(m["x"].(float64)), int(m["y"].(float64)), int(m["z"].(float64))}
}

type Quaternion struct {
	X, Y, Z, W float64
}

func NewQuaternionFromMap(data interface{}) *Quaternion {
	m := data.(map[string]interface{})
	return &Quaternion{m["x"].(float64), m["y"].(float64), m["z"].(float64), m["w"].(float64)}
}

func (this *Vector) MultiplyByQuaternion(q *Quaternion) *Vector {
	px, py, pz := float64(this.X), float64(this.Y), float64(this.Z)
	x, y, z, w := q.X, q.Y, q.Z, q.W
	this.X = int(math.Floor(0.5 + (w * w * px) + (2 * y * w * pz) - (2 * z * w * py) + (x * x * px) + (2 * y * x * py) + (2 * z * x * pz) - (z * z * px) - (y * y * px)))
	this.Y = int(math.Floor(0.5 + (2 * x * y * px) + (y * y * py) + (2 * z * y * pz) + (2 * w * z * px) - (z * z * py) + (w * w * py) - (2 * x * w * pz) - (x * x * py)))
	this.Z = int(math.Floor(0.5 + (2 * x * z * px) + (2 * y * z * py) + (z * z * pz) - (2 * w * y * px) - (y * y * pz) + (2 * w * x * py) - (x * x * pz) + (w * w * pz)))
	return this
}
func (this *Vector) Apply(q *Quaternion) {
}

func NewSamplePiece() *[]*Vector {
	return &[]*Vector{
		&Vector{0, 0, 0},
		&Vector{1, 0, 0},
		&Vector{2, 0, 0},
		&Vector{1, 1, 0},
	}
}
