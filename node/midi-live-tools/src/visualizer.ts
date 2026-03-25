import { parseMidiMessage } from './lib/midi-message';
import { BLACK_KEY_PITCH_CLASSES, PIANO_MIDI_START, PIANO_MIDI_END } from './lib/midi-constants';

const keyboard = document.getElementById('keyboard')!;

const keyMap = new Map<number, HTMLElement>();

for (let note = PIANO_MIDI_START; note <= PIANO_MIDI_END; note++) {
  const key = document.createElement('div');
  key.classList.add('key');
  if (BLACK_KEY_PITCH_CLASSES.includes(note % 12 as typeof BLACK_KEY_PITCH_CLASSES[number])) {
    key.classList.add('black');
  }
  key.dataset.note = note.toString();
  keyboard.appendChild(key);
  keyMap.set(note, key);
}

function handleMIDIMessage(e: MIDIMessageEvent) {
  const msg = parseMidiMessage(e.data);
  if (!msg) return;

  if (msg.command === 'note_on') {
    keyMap.get(msg.note)?.classList.add('active');
  } else if (msg.command === 'note_off') {
    keyMap.get(msg.note)?.classList.remove('active');
  }
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
  } catch (err) {
    console.error('MIDI init error', err);
  }
}

initMIDI();
