const keyboard = document.getElementById('keyboard')!;

// Create keys (MIDI note numbers 21–108)
const keyMap = new Map<number, HTMLElement>();
const blackKeys = [1, 3, 6, 8, 10]; // Relative to octave start

for (let note = 21; note <= 108; note++) {
  const key = document.createElement('div');
  key.classList.add('key');
  if (blackKeys.includes(note % 12)) {
    key.classList.add('black');
  }
  key.dataset.note = note.toString();
  keyboard.appendChild(key);
  keyMap.set(note, key);
}

// MIDI setup
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

function handleMIDIMessage(event: MIDIMessageEvent) {
  const data = event.data;
  if (!data || data.length < 3) return;
  const [status, note, velocity] = data;
  const command = status & 0xf0;

  if (command === 0x90 && velocity > 0) {
    // Note On
    keyMap.get(note)?.classList.add('active');
  } else if (command === 0x80 || (command === 0x90 && velocity === 0)) {
    // Note Off
    keyMap.get(note)?.classList.remove('active');
  }
}

initMIDI();
