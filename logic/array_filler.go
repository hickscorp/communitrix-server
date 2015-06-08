package logic

import "math/rand"

// Define our array filling template method.
type ArrayFiller func(at *Vector) int

// NewIntArrayFiller creates a new routine which will fill an array with a constant value.
func NewIntArrayFiller(val int) ArrayFiller {
	return func(*Vector) int { return val }
}

// NewRandomArrayFiller creates a new routine which will randomly fill an array with values between given bounds.
func NewRandomArrayFiller(min int, max int) ArrayFiller {
	rnd := max - min + 1
	return func(*Vector) int { return rand.Intn(rnd) + min }
}

// NewCopyArrayFiller will copy values from a given tri-ensional array.
func NewCopyArrayFiller(other ContentArray) ArrayFiller {
	return func(at *Vector) int { return other.Content[at.X][at.Y][at.Z] }
}
