package tonematrix

import "testing"

func TestGenerateMarkovMelody(t *testing.T) {
	row := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	melody := GenerateMarkovMelody(row, 32)

	if len(melody) != 32 {
		t.Errorf("expected melody of length 32, got %d", len(melody))
	}

	// Optional: check that all notes are in 0-11 range
	for _, note := range melody {
		if note < 0 || note > 11 {
			t.Errorf("note %d out of valid range 0-11", note)
		}
	}
}
