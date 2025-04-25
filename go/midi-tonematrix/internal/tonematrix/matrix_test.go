package tonematrix

import "testing"

func TestGenerateMatrix(t *testing.T) {
	row := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	matrix := GenerateMatrix(row)

	if len(matrix) != 12 {
		t.Errorf("matrix should have 12 rows, got %d", len(matrix))
	}

	for _, row := range matrix {
		if len(row) != 12 {
			t.Errorf("each matrix row should have 12 elements, got %d", len(row))
		}
	}
}
