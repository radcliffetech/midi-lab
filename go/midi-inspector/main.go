package main

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/v2/smf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: midi-validate <file.mid>")
		return
	}

	filePath := os.Args[1]
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("❌ Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	m, err := smf.ReadFrom(f)
	if err != nil {
		fmt.Printf("❌ Error parsing MIDI file: %v\n", err)
		return
	}

	fmt.Println("✅ MIDI file header validated successfully")
	fmt.Printf("Format: %d\n", m.Format())
	fmt.Printf("Ticks per quarter note (TPQN): %s\n", m.TimeFormat.String())
	fmt.Printf("Number of tracks: %d\n", len(m.Tracks))

	fmt.Println("✅ MIDI file validated successfully")
}
