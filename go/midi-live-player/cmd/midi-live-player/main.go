package main

import (
	"fmt"
	"log"

	"github.com/radcliffetech/midi-lab/go/midi-live-player/internal/bachgen"
	"github.com/radcliffetech/midi-lab/go/midi-live-player/internal/liveplayer"
)

func main() {
	player, err := liveplayer.NewPlayer("IAC Driver Bus 1") // Adjust to your IAC Bus name
	if err != nil {
		log.Fatal(err)
	}

	row := []int{0, 2, 4, 5, 7, 9, 11, 0, 2, 4, 5, 7}
	melody := bachgen.GenerateMelodyFromRow(row, 32)
	fmt.Println("Generated Melody:", melody)
	err = player.PlayMelody(melody, 120) // Play at 120 BPM
	if err != nil {
		log.Fatal(err)
	}
}
