package main

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	// Find a suitable input port
	in, err := midi.FindInPort("IAC Driver")
	if err != nil {
		fmt.Println("can't find IAC Driver")
		return
	}

	// Find a suitable output port that is not the same as the input
	var out drivers.Out
	for _, port := range midi.GetOutPorts() {
		if port.String() != in.String() {
			out = port
			break
		}
	}
	if out == nil {
		fmt.Println("no suitable output port found (must be different from input)")
		return
	}

	send, err := midi.SendTo(out)
	if err != nil {
		fmt.Printf("❌ Failed to open output port: %v\n", err)
		return
	}

	fmt.Println("Input Port:", in)
	fmt.Println("Output Port:", out)
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Printf("Forwarding message: %s\n", msg)
		if err := send(msg); err != nil {
			fmt.Printf("❌ Failed to send message: %v\n", err)
		}
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Println("🎹 Listening for MIDI input... Press Enter to stop.")
	_, _ = fmt.Scanln()

	stop()
}
