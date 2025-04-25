// chord-analyzer.ts
const keyboard = document.getElementById('keyboard')!;
const activeNotesDisplay = document.getElementById('active-notes')!;
const chordDisplay = document.getElementById('chord-display')!;
const statusDisplay = document.getElementById('status')!;

const keyMap = new Map<number, HTMLElement>();
const activeNotes = new Set<number>();
const blackKeys = [1, 3, 6, 8, 10];

for (let note = 21; note <= 108; note++) {
  const key = document.createElement('div');
  key.classList.add('key');
  if (blackKeys.includes(note % 12)) key.classList.add('black');
  key.dataset.note = note.toString();
  keyboard.appendChild(key);
  keyMap.set(note, key);
}

function updateDisplay() {
  const sortedNotes = Array.from(activeNotes).sort((a, b) => a - b);
  const names = sortedNotes.map(n => midiNoteName(n)).join(', ');
  activeNotesDisplay.textContent = `🎵 Active Notes: ${names}`;
  chordDisplay.textContent = detectChord(sortedNotes) ?? '–';
}

function handleMIDIMessage(e: MIDIMessageEvent) {
  const data = e.data;
  if (!data || data.length < 3) return;
  const [status, note, vel] = data;
  const cmd = status & 0xf0;

  if (cmd === 0x90 && vel > 0) {
    activeNotes.add(note);
    keyMap.get(note)?.classList.add('active');
  } else if (cmd === 0x80 || (cmd === 0x90 && vel === 0)) {
    activeNotes.delete(note);
    keyMap.get(note)?.classList.remove('active');
  }
  updateDisplay();
}

async function initMIDI() {
  try {
    const access = await navigator.requestMIDIAccess();
    for (const input of access.inputs.values()) {
      input.onmidimessage = handleMIDIMessage;
    }
    access.onstatechange = () => {
      for (const input of access.inputs.values()) {
        input.onmidimessage = handleMIDIMessage;
      }
    };
    statusDisplay.textContent = "✅ MIDI connected";
  } catch (err) {
    console.error("MIDI init error", err);
    statusDisplay.textContent = "❌ Failed to connect to MIDI";
  }
}

function midiNoteName(n: number): string {
  const names = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
  return names[n % 12] + Math.floor(n / 12 - 1);
}

// 🧠 Detect chord from active MIDI notes
function detectChord(notes: number[]): string | null {
  if (notes.length < 3) return null;
  const pitchClasses = [...new Set(notes.map(n => n % 12))].sort((a, b) => a - b);

  const chordTypes: [string, number[]][] = [
    ['Major', [0, 4, 7]],
    ['Minor', [0, 3, 7]],
    ['Diminished', [0, 3, 6]],
    ['Augmented', [0, 4, 8]],
    ['Sus2', [0, 2, 7]],
    ['Sus4', [0, 5, 7]],
    ['Major 7', [0, 4, 7, 11]],
    ['Minor 7', [0, 3, 7, 10]],
    ['Dominant 7', [0, 4, 7, 10]],
    ['Half-diminished 7', [0, 3, 6, 10]]
  ];

  for (let root = 0; root < 12; root++) {
    const rotated = pitchClasses.map(pc => (pc - root + 12) % 12).sort((a, b) => a - b);
    for (const [name, shape] of chordTypes) {
      if (shape.every(x => rotated.includes(x))) {
        return `${midiNoteName(root)} ${name}`;
      }
    }
  }

  return null;
}

initMIDI();