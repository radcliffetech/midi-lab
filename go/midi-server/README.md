# MIDI Server 🎹

A lightweight, production-ready Go WebSocket server that connects a local MIDI input device to web clients in real-time.  
Clients can interact with MIDI note data live over WebSocket and trigger virtual pads from a responsive web UI.

---

## ✨ Features

- Streams MIDI `NoteOn` events to WebSocket clients
- Sends MIDI notes based on client interaction (pad presses)
- Tracks active MIDI notes to avoid duplication
- Automatic note-off handling
- Live LED indicator for activity
- Cue system to broadcast scenes to all clients
- Idle timeout disconnection for inactive WebSocket clients
- Clean shutdown handling (no channel panics)
- Mobile and desktop touch/mouse support
- Color-coded, standardized server logs
- Automatic client reconnect and error handling
- Dynamic per-scene button color theming (gradient + pressed state)
- Hot reload scenes at runtime via `/reload-scenes`
- Multiple broadcasting strategies (Default, Buffered, Batch, Lossy) for optimizing under load
- Modular broadcaster interface for easy A/B testing

---

## 🛠 Requirements

- **Go** 1.20+
- **PortMIDI C library** installed locally

### Install PortMIDI

On macOS:

```bash
brew install portmidi
```

On Ubuntu/Debian:

```bash
sudo apt-get install libportmidi-dev
```

Windows users will need to manually install the PortMIDI SDK.

---

## ⚡ Setup and Running

Clone or download this project:

```bash
git clone https://github.com/yourusername/midi-lab.git
cd midi-lab/go/midi-server
```

Install Go dependencies:

```bash
go mod tidy
```

Start the server:

```bash
sh run.sh
```

### Running with Correct Build Flags

To ensure correct linking with the native PortMIDI library, use the provided `run.sh` script:

```bash
sh ./run.sh
```

Or if running manually, set `CGO_LDFLAGS`:

```bash
CGO_LDFLAGS="-L/opt/homebrew/lib" go run *.go
```

This ensures Go can find and link against the installed PortMIDI library when compiling and running.

Server starts:
- WebSocket endpoint: `ws://localhost:8080/ws`
- Static frontend UI: `http://localhost:8080`

Open `http://localhost:8080` to access the controller interface!

---

## 📈 Performance

- Tested to support 250+ concurrent connections on an M4 MacBook Pro.
- Easily scalable to higher connection counts with proper server resources.
- Built-in idle timeout handling prevents resource leaks.
- MIDI messages protected against concurrency race conditions.

---

## 📡 Broadcast Modes

You can choose the server's WebSocket broadcast strategy at startup:

```bash
go run main.go --broadcast-mode=buffered
```

Available modes:
- `default` — Immediate send, close slow clients
- `buffered` — Allow 50ms grace before closing (recommended)
- `batch` — Parallel fan-out to clients
- `lossy` — Skip slow clients without closing

Default is `buffered`.

---

## 🧪 Testing

### Local Manual Testing

1. Connect a MIDI controller or virtual MIDI input (e.g., [IAC Driver](https://support.apple.com/en-us/HT201884) on macOS).
2. Launch the server.
3. Play notes on your MIDI controller — WebSocket broadcasts and LED flashes will appear in browser.
4. Tap the virtual pads — notes will send back to your MIDI device output.

### Load Testing

Use the [MIDI Load Tester](../midi-load-tester) tool to simulate hundreds of WebSocket clients:

```bash
cd ../midi-load-tester
go run main.go --server=localhost:8080 --clients=500
```

You can also set a `--max-notes` limit during load tests.

---

## 📋 Roadmap Ideas

- Full MIDI `NoteOn` + `NoteOff` lifecycle tracking
- Handle multiple connected MIDI devices
- Velocity-based pad color feedback
- MIDI session recording and replay
- OAuth-based secure client authentication

---

## 📖 License

MIT License — feel free to fork, modify, and use!