# MIDI Load Tester

A lightweight Go tool to stress-test WebSocket-based MIDI servers by simulating hundreds or thousands of fake clients sending MIDI note and scene cue messages.

## Features

- Simulates many WebSocket clients (configurable)
- Sends randomized MIDI notes with human-like timing
- Some clients act as "scene masters" sending frequent cue changes
- Bursty "jam session" behavior for realistic traffic
- Live stats reporting (clients connected, notes sent)

## Requirements

- Go 1.20 or newer
- Internet access if testing remote servers
- WebSocket server compatible with the MIDI protocol (e.g., MIDI Lab server)

## Installation

Clone the repository and initialize Go modules:

```bash
git clone https://github.com/yourusername/midi-load-tester.git
cd midi-load-tester
go mod tidy
```

## Usage

To start the load test:

```bash
go run main.go --server=localhost:8080 --clients=50 --max-notes=500
```

### Options

| Flag | Description | Default |
|:-----|:------------|:--------|
| `--server` | WebSocket server address (host:port) | `localhost:8080` |
| `--clients` | Number of simulated clients to spawn | `50` |
| `--max-notes` | (optional) Stop each client after this sending number of notes | `none` | 

Example targeting a remote server:

```bash
go run main.go --server=yourserver.com:8080 --clients=2000
```

## How It Works

- Each client establishes a WebSocket connection.
- Clients send MIDI note messages randomly every 100ms–5s.
- 10% of clients ("scene masters") send cue change messages 30% of the time.
- Random velocity values simulate realistic note pressures.
- Disconnected clients automatically stop.
- Stats (clients connected, notes sent) print every 5 seconds.

## Notes

- Designed for load-testing lightweight WebSocket MIDI apps like MIDI Lab.
- Use responsibly — you can easily overwhelm servers if connection counts are too high!
