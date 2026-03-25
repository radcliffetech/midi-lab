## Music Theory API

A Flask-based REST API for music theory operations powered by [music21](https://web.mit.edu/music21/).

### Endpoints

#### `POST /generate` - Counterpoint Generation
Generate first-species counterpoint above a cantus firmus.

```bash
curl -X POST http://localhost:5000/generate \
  -H "Content-Type: application/json" \
  -d '{"cantus": ["C4", "D4", "E4", "F4", "G4"], "midi": false}'
```

Options: `realtime` (play via MIDI output), `midi` (return MIDI file, default true).

#### `POST /analyze` - Interval Analysis
Analyze a melody's intervals, consonance, and voice leading characteristics.

```bash
curl -X POST http://localhost:5000/analyze \
  -H "Content-Type: application/json" \
  -d '{"melody": ["C4", "E4", "G4", "F4", "D4"]}'
```

Returns: interval sequence, consonance/dissonance classification, statistics (largest leap, stepwise motion %, most common interval).

#### `POST /scales` - Scale & Mode Lookup
Get notes, MIDI numbers, and intervals for any scale or mode.

```bash
curl -X POST http://localhost:5000/scales \
  -H "Content-Type: application/json" \
  -d '{"key": "D", "type": "dorian"}'
```

Supported types: `major`, `natural_minor`, `harmonic_minor`, `melodic_minor`, `dorian`, `phrygian`, `lydian`, `mixolydian`, `aeolian`, `locrian`.

Add `"midi": true` to get a MIDI file of the scale.

#### `POST /detect-scale` - Scale Detection
Detect which scales a set of notes belongs to.

```bash
curl -X POST http://localhost:5000/detect-scale \
  -H "Content-Type: application/json" \
  -d '{"notes": ["C4", "D4", "Eb4", "F4", "G4", "A4", "Bb4"]}'
```

#### `POST /harmonize` - Four-Part Harmonization
Generate a Bach-style four-part chorale harmonization (soprano, alto, tenor, bass).

```bash
curl -X POST http://localhost:5000/harmonize \
  -H "Content-Type: application/json" \
  -d '{"melody": ["C4", "D4", "E4", "F4", "G4"], "key": "C"}'
```

Returns a MIDI file with four tracks.

### Setup & Run

```bash
cd python
make setup   # creates venv and installs dependencies
make run     # starts Flask dev server on :5000
make test    # runs pytest suite
```

### Dependencies

- **music21** - Music theory and analysis
- **mido** + **python-rtmidi** - MIDI file I/O and real-time output
- **Flask** - HTTP API framework
