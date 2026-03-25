# Jeffrey's MIDI Lab

![CI](https://github.com/radcliffetech/midi-lab/actions/workflows/ci.yml/badge.svg)

A polyglot collection of MIDI tools written in **Go**, **TypeScript**, and **Python** for converting, transforming, generating, and visualizing MIDI data.

## Go CLI Tools

| Tool | Description |
|------|-------------|
| [midi-server](go/midi-server/) | WebSocket server streaming real-time MIDI to browser clients (250+ concurrent connections) |
| [midi-tonematrix](go/midi-tonematrix/) | 12-tone serialism matrix generator with Markov melody generation |
| [midi-player](go/midi-player/) | MIDI file playback to hardware/virtual ports |
| [midi-inspector](go/midi-inspector/) | MIDI file validation and metadata inspection |
| [midi-logger](go/midi-logger/) | Real-time MIDI input logging |
| [midi-thru](go/midi-thru/) | MIDI input-to-output passthrough routing |
| [midi-load-tester](go/midi-load-tester/) | WebSocket client simulator for load testing midi-server |

```bash
cd go/midi-inspector && go run main.go examples/bach_minuet.mid
```

## Browser Tools (TypeScript + Vite)

| Tool | Description |
|------|-------------|
| Monitor | Real-time MIDI message logging |
| Visualizer | Keyboard layout with glowing notes |
| Spiral Viewer | Animated spiral display with fading ghost trails |
| Drum Pad | Clickable grid that sends live MIDI output |
| Histogram | Live note frequency histogram |
| Chord Analyzer | Real-time chord and key detection |

```bash
cd node/midi-live-tools && npm install && npm run dev
```

## Python Music Theory API

| Endpoint | Description |
|----------|-------------|
| `POST /generate` | First-species counterpoint generation |
| `POST /analyze` | Interval analysis and voice leading evaluation |
| `POST /scales` | Scale and mode lookup by key |
| `POST /detect-scale` | Detect scales from a set of notes |
| `POST /harmonize` | Four-part Bach-style chorale harmonization |

```bash
cd python && make setup && make run
```

## Structure

- `go/` — Compiled CLI tools and libraries for MIDI manipulation
- `node/` — Browser-based live MIDI tools (WebMIDI API + Vite)
- `python/` — Music theory API powered by music21
- `examples/` — Sample MIDI files

## License

MIT License
