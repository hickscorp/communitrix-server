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
	// Our direction checks.
	directions = []*logic.Vector{
		&logic.Vector{-1, 0, 0}, &logic.Vector{+1, 0, 0}, // Left / Right.
		&logic.Vector{0, -1, 0}, &logic.Vector{0, +1, 0}, // Top / Bottom.
		&logic.Vector{0, 0, -1}, &logic.Vector{0, 0, +1}, // Forward / Backward
	}
	log    = logging.MustGetLogger("communitrix")
	format = logging.MustStringFormatter("%{color}%{level:.1s} %{shortfunc}%{color:reset} %{message}")
)

type CellularAutomata struct {
	size            *logic.Vector
	result          *array.ContentArray
	probabilities   *array.ContentArray
	spreadingFactor float64
}

func NewCellularAutomata(size *logic.Vector) *CellularAutomata {
	return &CellularAutomata{
		size:            size,
		spreadingFactor: 0.1,
	}
}

// Run creates the unit.
func (this *CellularAutomata) Run(density float64) *logic.Piece {
	// Normalize inputs.
	density = math.Min(1.0, math.Max(0.0, density))
	// Prepare the total number of blocks to be created.
	targetSize := int(float64(this.size.Volume()) * density)
	// Prepare the result and probabilities array.
	this.probabilities = array.NewContentArray(this.size, nil)
	this.result = array.NewContentArray(this.size, nil)

	// Aim for the center, and start filling our initial cell.
	at := this.size.Copy().Half()
	// Set the first block inside the results array.
	this.fillCell(at, 1)
	// Keep track of the total number of cells we've added.
	totalCellsAdded := 1

	var cellsPerIteration, freeLocCount, probSum int
	for {
		// Compute the target cells to generate during this iteration.
		cellsPerIteration = util.QuickIntRound(math.Max(1, math.Floor(float64(totalCellsAdded)*this.spreadingFactor)))
		// Whenever we're reaching the target number of cells, cap it correctly.
		if totalCellsAdded+cellsPerIteration > targetSize {
			cellsPerIteration = targetSize - totalCellsAdded
		}

		// Prepare iteration.
		freeLocCount, probSum = 0, 0
		groups := map[int][]logic.Vector{}
		// Look at every single array element.
		var pro int
		for x := 0; x < this.probabilities.Size.X; x++ {
			for y := 0; y < this.probabilities.Size.Y; y++ {
				for z := 0; z < this.probabilities.Size.Z; z++ {
					pro = this.probabilities.Content[x][y][z]
					if pro != 0 {
						freeLocCount++
						if locations, ok := groups[pro]; ok {
							groups[pro] = append(locations, logic.Vector{x, y, z})
						} else {
							probSum += pro
							groups[pro] = []logic.Vector{logic.Vector{x, y, z}}
						}
					}
				}
			}
		}
		// Build a dice array.
		dice := make([]int, 0, probSum)
		for pro := range groups {
			for c := 0; c < pro; c++ {
				dice = append(dice, pro)
			}
		}
		for i := 0; i < cellsPerIteration; i++ {
			pro := dice[rand.Intn(probSum)]
			locations := groups[pro]
			location := locations[rand.Intn(len(locations))]
			if !this.fillCell(&location, 1) {
				i--
			}
		}
		// Whenever we reach the target size, stop.
		totalCellsAdded += cellsPerIteration
		log.Info("Generated %d/%d pieces, factor is now %f (%d per pass)", totalCellsAdded, targetSize, this.spreadingFactor, cellsPerIteration)
		if totalCellsAdded >= targetSize {
			break
		}
	}

	piece := &logic.Piece{
		Size:    this.size,
		Content: make([]*logic.Vector, 0, this.size.Volume()),
	}
	this.result.Each(func(at *logic.Vector, val int) {
		if val != 0 {
			piece.Content = append(piece.Content, at.Copy())
		}
	})
	this.result, this.probabilities = nil, nil
	return piece
}

func (this *CellularAutomata) fillCell(at *logic.Vector, val int) bool {
	if this.result.Content[at.X][at.Y][at.Z] != 0 {
		return false
	}
	this.probabilities.Content[at.X][at.Y][at.Z] = 0
	this.result.Content[at.X][at.Y][at.Z] = val
	for _, d := range directions {
		if pos := at.Copy().Translate(d); pos.X >= 0 && pos.X < this.size.X && pos.Y >= 0 && pos.Y < this.size.Y && pos.Z >= 0 && pos.Z < this.size.Z {
			// If there is already a probability there, increase it.
			actual := this.probabilities.Content[pos.X][pos.Y][pos.Z]
			this.probabilities.Content[pos.X][pos.Y][pos.Z] = actual + 1
		}
	}
	return true
}

func (this *CellularAutomata) VectorTo1DCoords(v *logic.Vector) int {
	return v.X + (v.Y * this.size.X) + (v.Z * this.size.X * this.size.Y)
}
