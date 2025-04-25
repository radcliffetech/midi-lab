package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/radcliffetech/midi-lab/go/midi-tonematrix/internal/tonematrix"
)

func printHelp() {
	fmt.Print(`
	Usage:
	  tone-matrix generate [12 pitch classes] [--pretty] [--output filename.txt]
	  tone-matrix random [--pretty] [--output filename.txt]
	  tone-matrix generate-midi [12 pitch classes] [--length N]
	
	Options:
	  --pretty   Display note names (C, C♯, D, etc.)
	  --output   Save the matrix to a text file
	  --length   Length of the generated melody
	  --help     Show this help message
	`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	pretty := false
	outputFile := ""
	args := os.Args[2:]

	// Parse flags
	filteredArgs := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--pretty":
			pretty = true
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++ // Skip next arg because it's the filename
			} else {
				fmt.Println("Error: --output flag requires a filename")
				os.Exit(1)
			}
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	switch command {
	case "--help", "-h", "help":
		printHelp()
	case "generate":
		if len(filteredArgs) != 12 {
			fmt.Println("Error: Provide exactly 12 integers for the row.")
			os.Exit(1)
		}
		var row []int
		for _, arg := range filteredArgs {
			n, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("Invalid number: %s\n", arg)
				os.Exit(1)
			}
			row = append(row, n)
		}
		if err := tonematrix.ValidateRow(row); err != nil {
			fmt.Println("Validation error:", err)
			os.Exit(1)
		}
		matrix := tonematrix.GenerateMatrix(row)
		output := tonematrix.MatrixToString(matrix, pretty)
		if outputFile != "" {
			os.WriteFile(outputFile, []byte(output), 0644)
			fmt.Println("Matrix saved to", outputFile)
		} else {
			fmt.Print(output)
		}
	case "random":
		row := tonematrix.GenerateRandomRow()
		matrix := tonematrix.GenerateMatrix(row)
		output := tonematrix.MatrixToString(matrix, pretty)
		if outputFile != "" {
			os.WriteFile(outputFile, []byte(output), 0644)
			fmt.Println("Matrix saved to", outputFile)
		} else {
			fmt.Print(output)
		}
	case "generate-midi":
		length := 32 // default length
		filteredArgs := []string{}

		for i := 0; i < len(args); i++ {
			if args[i] == "--length" {
				if i+1 < len(args) {
					l, err := strconv.Atoi(args[i+1])
					if err != nil {
						fmt.Println("Invalid length value:", args[i+1])
						os.Exit(1)
					}
					length = l
					i++ // skip next arg
				} else {
					fmt.Println("Error: --length requires a number")
					os.Exit(1)
				}
			} else {
				filteredArgs = append(filteredArgs, args[i])
			}
		}

		if len(filteredArgs) != 12 {
			fmt.Println("Error: Provide exactly 12 integers for the row.")
			os.Exit(1)
		}

		var row []int
		for _, arg := range filteredArgs {
			n, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("Invalid number: %s\n", arg)
				os.Exit(1)
			}
			row = append(row, n)
		}

		if err := tonematrix.ValidateRow(row); err != nil {
			fmt.Println("Validation error:", err)
			os.Exit(1)
		}

		melody := tonematrix.GenerateMarkovMelody(row, length)
		fmt.Println("Generated Melody:")
		for _, n := range melody {
			fmt.Printf("%d ", n)
		}
		fmt.Println()
	default:
		fmt.Println("Unknown command:", command)
		printHelp()
		os.Exit(1)
	}
}
