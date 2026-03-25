import { detectChord, midiNoteName } from './lib/midi-utils';
import { parseMidiMessage } from './lib/midi-message';
import { BLACK_KEY_PITCH_CLASSES, PIANO_MIDI_START, PIANO_MIDI_END } from './lib/midi-constants';

const keyboard = document.getElementById('keyboard')!;
const activeNotesDisplay = document.getElementById('active-notes')!;
const chordDisplay = document.getElementById('chord-display')!;
const statusDisplay = document.getElementById('status')!;

const keyMap = new Map<number, HTMLElement>();
const activeNotes = new Set<number>();

for (let note = PIANO_MIDI_START; note <= PIANO_MIDI_END; note++) {
  const key = document.createElement('div');
  key.classList.add('key');
  if (BLACK_KEY_PITCH_CLASSES.includes(note % 12 as typeof BLACK_KEY_PITCH_CLASSES[number])) key.classList.add('black');
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
  const msg = parseMidiMessage(e.data);
  if (!msg) return;

  if (msg.command === 'note_on') {
    activeNotes.add(msg.note);
    keyMap.get(msg.note)?.classList.add('active');
  } else if (msg.command === 'note_off') {
    activeNotes.delete(msg.note);
    keyMap.get(msg.note)?.classList.remove('active');
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
