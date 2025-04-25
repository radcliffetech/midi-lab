

# MIDI Server

This Go application listens to a local MIDI input device and broadcasts MIDI NoteOn messages over a WebSocket server.  
Clients (like web browsers) can connect and receive live MIDI data in real time.

## 🛠 Requirements

- Go 1.20+
- PortMIDI C library installed on your system

### Install PortMIDI

**On macOS (using Homebrew):**

```bash
brew install portmidi
```

**On Ubuntu/Debian Linux:**

```bash
sudo apt-get install libportmidi-dev
```

(Windows setup requires manual PortMIDI installation.)

---

## ⚡ Setup and Run

Clone or download this project. Then:

```bash
# Inside the midi-server directory:
go mod tidy

# Run the server
go run main.go
```

The server will start:
- WebSocket endpoint: `ws://localhost:8080/ws`
- Static file server: `http://localhost:8080`

Open your browser to `http://localhost:8080` to view the live MIDI data stream.

---

## 🎛 Features

- Captures incoming MIDI notes from any connected device
- Broadcasts live NoteOn events to all WebSocket clients
- Simple frontend to view note numbers and velocities
- Small, fast, standalone Go server
- No database, no backend processing — pure WebMIDI stream

---

## 📋 Roadmap Ideas

- Add NoteOff support
- Support multiple simultaneous MIDI inputs
- Velocity-sensitive visualizations
- Record + Replay MIDI sessions