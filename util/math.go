package util

import "math"

func QuickRound(f float64) float64 { return math.Floor(f + .5) }
func QuickIntRound(f float64) int  { return int(QuickRound(f)) }
