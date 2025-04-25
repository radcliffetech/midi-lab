package tonematrix

import "testing"

func TestValidateRowGood(t *testing.T) {
	row := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	if err := ValidateRow(row); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateRowBadDuplicates(t *testing.T) {
	row := []int{0, 1, 2, 2, 4, 5, 6, 7, 8, 9, 10, 11}
	if err := ValidateRow(row); err == nil {
		t.Errorf("expected error for duplicate values, got nil")
	}
}

func TestValidateRowBadRange(t *testing.T) {
	row := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12}
	if err := ValidateRow(row); err == nil {
		t.Errorf("expected error for out-of-range values, got nil")
	}
}
