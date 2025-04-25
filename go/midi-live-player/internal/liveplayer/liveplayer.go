package liveplayer

import (
	"fmt"
	"time"

	"gitlab.com/gomidi/midi/v2"

	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

type Player struct {
	outputName string
	out        drivers.Out
}

func printPorts() {
	fmt.Println(midi.GetOutPorts())
}

// NewPlayer initializes a new Player connected to a specified MIDI output.
func NewPlayer(outputName string) (*Player, error) {
	printPorts()
	out, err := midi.FindOutPort("IAC Driver")
	if err != nil {
		return nil, fmt.Errorf("can't find IAC Driver")
	}
	if err := out.Open(); err != nil {
		return nil, fmt.Errorf("failed to open output: %w", err)
	}
	return &Player{outputName: outputName, out: out}, nil
}

// PlayMelody sends a melody to the output device live, according to the given BPM.
func (p *Player) PlayMelody(melody []int, bpm int) error {
	defer p.out.Close()

	ticksPerBeat := time.Minute / time.Duration(bpm)

	for _, note := range melody {
		pitch := uint8(60 + note) // Middle C + offset
		fmt.Printf("Playing note: %d\n", pitch)
		p.out.Send([]byte{0x90, pitch, 100}) // NoteOn (channel 0)
		time.Sleep(ticksPerBeat)
		p.out.Send([]byte{0x80, pitch, 0}) // NoteOff
	}

	return nil
}
