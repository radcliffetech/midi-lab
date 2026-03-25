## 🎛 MIDI Live Tools (TypeScript + Vite)

A browser-based WebMIDI playground with real-time visualizations, diagnostics, and creative interfaces.

Built with:
- Vite + TypeScript
- Web MIDI API
- Bootstrap for layout and styling
- HTML Canvas for animation

### 🔗 Live Tools

Each tool is a standalone HTML+TS+CSS interface:

| Tool              | Description                                         |
|-------------------|-----------------------------------------------------|
| 🎛 Monitor         | Real-time logging of incoming MIDI messages         |
| 🎹 Visualizer      | Keyboard layout with glowing notes                 |
| 🌀 Spiral Viewer   | Animated spiral note display with fading ghost trails |
| 🥁 Drum Pad        | Clickable drum grid that sends live MIDI output      |
| 📊 Histogram       | Live note frequency histogram from input stream     |
| 🎵 Chord Analyzer  | Real-time chord and key detection from MIDI input    |

### ⚡ Setup and Run

To get started locally:

```bash
# Install dependencies
npm install

# Start the Vite development server
npm run dev

# Open your browser at http://localhost:5173
```

### 📁 Path

```bash
node/midi-live-tools/
```

### ✅ Features

- No backend required – pure browser WebMIDI magic
- Live MIDI input *and* output (drum pad sends real MIDI notes)
- Dynamic spiral visualizer with note ghosting and key highlighting
- Responsive Bootstrap layouts
- MIDI device selection menus
- Real-time MIDI message logging
- Playable browser-based MIDI drum pad (works with DAWs!)
- TypeScript + Vite project structure for rapid loading and development



This uses Vite for fast reloading during development.  
Ensure you use a Chromium-based browser (like Chrome) to enable WebMIDI support.

### 🚀 Why This Project Matters

This project showcases real-time audio interaction, Web API mastery, TypeScript architecture, and dynamic UI design — all without needing server infrastructure.

**Perfect for:**
- Music tech startups
- Browser-based creative tools
- WebMIDI-enabled product demos
- Interactive sound experiments
