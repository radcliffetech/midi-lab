package tonematrix

import (
	"math/rand"
	"time"
)

// GenerateMarkovMelody generates a melody of a given length based on a simple Markov model built from the tone row.
func GenerateMarkovMelody(row []int, length int) []int {
	if length <= 0 {
		return []int{}
	}

	// Build a simple transition map from the original row
	transitions := make(map[int]int)
	for i := 0; i < len(row)-1; i++ {
		transitions[row[i]] = row[i+1]
	}
	transitions[row[len(row)-1]] = row[0] // wrap around

	// Start with a random note from the row
	rand.Seed(time.Now().UnixNano())
	current := row[rand.Intn(len(row))]
	melody := []int{current}

	for i := 1; i < length; i++ {
		next, exists := transitions[current]
		if !exists {
			// fallback: pick random note from row
			next = row[rand.Intn(len(row))]
		}
		melody = append(melody, next)
		current = next
	}

	return melody
}
