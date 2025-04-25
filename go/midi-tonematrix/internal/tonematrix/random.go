package tonematrix

import (
	"math/rand"
	"time"
)

func GenerateRandomRow() []int {
	rand.Seed(time.Now().UnixNano())
	row := rand.Perm(12)
	return row
}
