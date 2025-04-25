package tonematrix

import "testing"

func TestGenerateRandomRow(t *testing.T) {
	row := GenerateRandomRow()

	if len(row) != 12 {
		t.Errorf("expected 12 elements, got %d", len(row))
	}

	if err := ValidateRow(row); err != nil {
		t.Errorf("generated row is invalid: %v", err)
	}
}
