// chord-analyzer.ts
const keyboard = document.getElementById('keyboard')!;
const activeNotesDisplay = document.getElementById('active-notes')!;
const chordDisplay = document.getElementById('chord-display')!;
const statusDisplay = document.getElementById('status')!;

import { detectChord, midiNoteName } from './lib/midi-utils';

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
  const names = sortedNotes.map((n) => midiNoteName(n)).join(', ');
  activeNotesDisplay.textContent = `${names}`;
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
    statusDisplay.textContent = '✅ MIDI connected';
  } catch (err) {
    console.error('MIDI init error', err);
    statusDisplay.textContent = '❌ Failed to connect to MIDI';
  }
}

initMIDI();
