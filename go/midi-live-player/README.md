
# MIDI Live Player

A simple Go application that generates and performs live MIDI melodies based on algorithmic composition techniques, inspired by the style of Bach.

This project connects to the macOS IAC Driver and streams live MIDI notes to any DAW (Ableton, Logic, etc.) in real time.

---

## 🚀 Features

- Real-time MIDI performance using the IAC Driver
- Melody generation based on tone rows and probabilistic stepwise motion
- Configurable note generation speed (BPM)
- Modular design for easy future expansion (harmony, rhythm variations, dynamics)

---

## 📦 Project Structure

```
/cmd/midi-live-player/    → CLI entry point
/internal/liveplayer/     → Manages real-time MIDI output
/internal/bachgen/        → Melody generation logic
```

---

## 🛠️ Usage

1. Make sure the **IAC Driver** is enabled on macOS (Audio MIDI Setup > MIDI Studio > IAC Driver > Online).
2. Set your DAW to receive MIDI input from "IAC Driver Bus 1."
3. From this project directory:

```bash
go run ./cmd/midi-live-player
```

4. Listen to generated melodies live!

---

## 📜 Future Plans

- Add rhythmic variation (eighths, quarters, rests)
- Add dynamic velocity variation
- Generate simple 2-voice harmony
- WebSocket interface for interactive control

---

## ✨ Credits

Built with ❤️ for the joy of computational musicology and live coding creativity.