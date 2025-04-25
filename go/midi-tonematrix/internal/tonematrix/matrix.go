package tonematrix

import (
	"fmt"
	"strings"
)

func GenerateMatrix(row []int) [][]int {
	matrix := make([][]int, 12)
	for i := range matrix {
		matrix[i] = make([]int, 12)
	}

	for i := 0; i < 12; i++ {
		interval := (row[i] - row[0] + 12) % 12
		for j := 0; j < 12; j++ {
			matrix[i][j] = (row[j] - interval + 12) % 12
		}
	}

	return matrix
}

var noteNames = []string{"C", "C♯", "D", "D♯", "E", "F", "F♯", "G", "G♯", "A", "A♯", "B"}

func PrintMatrix(matrix [][]int, pretty bool) {
	// Print header
	fmt.Printf("    ")
	for i := 0; i < 12; i++ {
		if pretty {
			fmt.Printf("%3s ", noteNames[i])
		} else {
			fmt.Printf("%2d ", i)
		}
	}
	fmt.Println()

	fmt.Println("   +" + strings.Repeat("----", 12))

	for i, row := range matrix {
		if pretty {
			fmt.Printf("%3s| ", noteNames[i])
		} else {
			fmt.Printf("%2d | ", i)
		}
		for _, val := range row {
			if pretty {
				fmt.Printf("%3s ", noteNames[val])
			} else {
				fmt.Printf("%2d ", val)
			}
		}
		fmt.Println()
	}
}

func MatrixToString(matrix [][]int, pretty bool) string {
	var sb strings.Builder

	// Header
	sb.WriteString("    ")
	for i := 0; i < 12; i++ {
		if pretty {
			sb.WriteString(fmt.Sprintf("%3s ", noteNames[i]))
		} else {
			sb.WriteString(fmt.Sprintf("%2d ", i))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("   +" + strings.Repeat("----", 12) + "\n")

	for i, row := range matrix {
		if pretty {
			sb.WriteString(fmt.Sprintf("%3s| ", noteNames[i]))
		} else {
			sb.WriteString(fmt.Sprintf("%2d | ", i))
		}
		for _, val := range row {
			if pretty {
				sb.WriteString(fmt.Sprintf("%3s ", noteNames[val]))
			} else {
				sb.WriteString(fmt.Sprintf("%2d ", val))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
