# MIDI Tone Matrix

A simple CLI tool for generating and displaying 12-tone matrices.

## Usage

```bash
go run main.go generate [12 pitch classes]
go run main.go random
```
# MIDI Tone Matrix

A simple Go CLI tool for working with 12-tone serialism techniques, including:

- Generating and displaying 12-tone matrices
- Randomly generating valid tone rows
- Generating Markov-chain-based melodies from a 12-tone row
- (Coming soon) Exporting melodies to MIDI files

---

## 📦 Commands

### Generate a Matrix from a Row

```bash
go run main.go generate [12 pitch classes] [--pretty] [--output filename.txt]
```

Example:

```bash
go run main.go generate 0 2 1 5 4 3 7 6 8 10 9 11
```

Flags:
- `--pretty`: Show note names (C, C♯, D, etc.) instead of numbers
- `--output filename.txt`: Save the matrix to a text file

---

### Generate a Random Tone Row and Matrix

```bash
go run main.go random [--pretty] [--output filename.txt]
```

---

### Generate a Markov Melody

```bash
go run main.go generate-midi [12 pitch classes] [--length N]
```

Example:

```bash
go run main.go generate-midi 0 2 1 5 4 3 7 6 8 10 9 11 --length 32
```

This generates a Markov-chain melody from the given row and prints it.

---

## 🔧 Global Installation

To install the CLI globally:

```bash
cd path/to/midi-tonematrix
go install

---

## 📜 Coming Soon

- Export generated melodies as MIDI files
- More control over rhythm and dynamics
- Live MIDI playback

---

## ✨ Project Goals

- Explore algorithmic composition
- Provide tools for music analysis and creativity
- Build fun, educational, open-source music tech