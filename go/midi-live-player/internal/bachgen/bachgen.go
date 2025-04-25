package bachgen

import (
	"math/rand"
	"time"
)

// GenerateMelodyFromRow generates a melody based on a tone row by randomly walking through it.
func GenerateMelodyFromRow(row []int, length int) []int {
	if len(row) == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())

	melody := make([]int, 0, length)
	current := row[rand.Intn(len(row))] // start from random note

	for i := 0; i < length; i++ {
		melody = append(melody, current)

		// 70% chance stepwise motion (next/previous)
		if rand.Float64() < 0.7 {
			dir := rand.Intn(2)*2 - 1 // -1 or +1
			nextIndex := -1
			for j, note := range row {
				if note == current {
					nextIndex = j + dir
					break
				}
			}
			if nextIndex >= 0 && nextIndex < len(row) {
				current = row[nextIndex]
			} else {
				// edge case: pick random again
				current = row[rand.Intn(len(row))]
			}
		} else {
			// 30% chance random jump
			current = row[rand.Intn(len(row))]
		}
	}

	return melody
}
