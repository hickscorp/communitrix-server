package gen

import (
	"github.com/op/go-logging"
	"gogs.pierreqr.fr/doodloo/communitrix/array"
	"gogs.pierreqr.fr/doodloo/communitrix/logic"
	"gogs.pierreqr.fr/doodloo/communitrix/util"
	"math"
	"math/rand"
)

var (
	log        = logging.MustGetLogger("communitrix")
	format     = logging.MustStringFormatter("%{color}%{level:.1s} %{shortfunc}%{color:reset} %{message}")
	directions []*logic.Vector
)

func init() {
	directions = make([]*logic.Vector, 6)
	directions[0] = &logic.Vector{-1, 0, 0}
	directions[1] = &logic.Vector{+1, 0, 0}
	directions[2] = &logic.Vector{0, -1, 0}
	directions[3] = &logic.Vector{0, +1, 0}
	directions[4] = &logic.Vector{0, 0, -1}
	directions[5] = &logic.Vector{0, 0, +1}
}

type CellularAutomata struct {
	size            *logic.Vector
	result          *array.ContentArray
	probabilities   *array.ContentArray
	spreadingFactor float64
	targetSize      int
}

func NewCellularAutomata(size *logic.Vector) *CellularAutomata {
	return &CellularAutomata{
		size:            size,
		targetSize:      size.Volume(),
		spreadingFactor: 0.01,
		result:          array.NewContentArray(size, nil),
	}
}

func (this *CellularAutomata) Run() array.ContentArrayFiller {
	this.probabilities = array.NewContentArray(this.size, array.NewIntContentArrayFiller(0))
	at := this.size.Copy().Half()
	this.result.Content[at.X][at.Y][at.Z] = 1

	totalCellsAdded, cellsPerIteration := 1, 0

	for {
		this.computeProbabilities(at)
		cellsPerIteration = util.QuickIntRound(math.Max(1, math.Floor(float64(totalCellsAdded)*this.spreadingFactor)))

		for i := 0; i < cellsPerIteration; i++ {
			freeLocCount := 0
			probSum := 0
			groups := map[int][]logic.Vector{}
			this.probabilities.Each(func(at *logic.Vector, pro int) {
				if pro == 0 {
					return
				}
				freeLocCount++
				if locations, ok := groups[pro]; ok {
					groups[pro] = append(locations, *at.Copy())
				} else {
					probSum += pro * pro
					groups[pro] = []logic.Vector{*at.Copy()}
				}
			})
			// Build a dice array.
			dice := make([]int, 0, probSum)
			for pro := range groups {
				for c := 0; c < pro*pro; c++ {
					dice = append(dice, pro)
				}
			}
			log.Warning("Dice: %d", dice)
			pro := dice[rand.Intn(probSum)]
			locations := groups[pro]
			location := locations[rand.Intn(len(locations))]
			this.result.Content[location.X][location.Y][location.Z] = 1
			this.probabilities.Content[location.X][location.Y][location.Z] = 0
			this.computeProbabilities(&location)
		}
		// Whenever we reach the target size, stop.
		if totalCellsAdded += cellsPerIteration; totalCellsAdded >= this.targetSize {
			break
		}
	}

	// Clone the result to prevent locking memory from the closure.
	result := this.result.Clone()
	// Free memory up.
	this.result = nil
	this.probabilities = nil
	// Return content builder.
	return func(at *logic.Vector) int { return result.Clone().Content[at.X][at.Y][at.Z] }
}

func (this *CellularAutomata) computeProbabilities(at *logic.Vector) {
	for _, d := range directions {
		if pos := at.Copy().Translate(d); pos.X >= 0 && pos.X < this.size.X && pos.Y >= 0 && pos.Y < this.size.Y && pos.Z >= 0 && pos.Z < this.size.Z {
			// If there is already a probability there, increase it.
			actual := this.probabilities.Content[pos.X][pos.Y][pos.Z]
			this.probabilities.Content[pos.X][pos.Y][pos.Z] = actual + 1
		}
	}
}

func (this *CellularAutomata) VectorTo1DCoords(v *logic.Vector) int {
	return v.X + (v.Y * this.size.X) + (v.Z * this.size.X * this.size.Y)
}
