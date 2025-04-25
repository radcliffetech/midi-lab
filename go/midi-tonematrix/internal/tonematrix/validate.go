package tonematrix

import "errors"

func ValidateRow(row []int) error {
	if len(row) != 12 {
		return errors.New("row must have exactly 12 elements")
	}
	seen := make(map[int]bool)
	for _, num := range row {
		if num < 0 || num > 11 {
			return errors.New("row elements must be between 0 and 11")
		}
		if seen[num] {
			return errors.New("row must not contain duplicate elements")
		}
		seen[num] = true
	}
	return nil
}
