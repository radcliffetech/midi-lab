export interface DrumPad {
  name: string;
  note: number;
}

/** General MIDI drum pad mapping. */
export const GM_DRUM_PADS: DrumPad[] = [
  { name: 'Kick', note: 36 },
  { name: 'Snare', note: 38 },
  { name: 'Clap', note: 39 },
  { name: 'Hi-Hat Closed', note: 42 },
  { name: 'Hi-Hat Open', note: 46 },
  { name: 'Tom Low', note: 41 },
  { name: 'Tom Mid', note: 45 },
  { name: 'Tom High', note: 48 },
  { name: 'Crash 1', note: 49 },
  { name: 'Crash 2', note: 57 },
  { name: 'Ride', note: 51 },
  { name: 'Perc 1', note: 47 },
  { name: 'Perc 2', note: 50 },
  { name: 'Perc 3', note: 53 },
  { name: 'Perc 4', note: 55 },
  { name: 'Perc 5', note: 60 },
];

const NOTE_DURATION_MS = 100;
const DEFAULT_VELOCITY = 100;

let midiOutput: MIDIOutput | null = null;

function sendMidiNote(note: number, velocity: number = DEFAULT_VELOCITY) {
  if (!midiOutput) return;
  midiOutput.send([0x90, note, velocity]);
  setTimeout(() => midiOutput?.send([0x80, note, 0]), NOTE_DURATION_MS);
}

async function init() {
  const outputSelect = document.getElementById('outputSelect') as HTMLSelectElement;
  const padGrid = document.getElementById('pad-grid')!;

  try {
    const midiAccess = await navigator.requestMIDIAccess();

    midiAccess.outputs.forEach((output) => {
      const option = document.createElement('option');
      option.value = output.id;
      option.textContent = output.name ?? '(unnamed output)';
      outputSelect.appendChild(option);
    });

    outputSelect.addEventListener('change', () => {
      const selectedId = outputSelect.value;
      midiOutput = [...midiAccess.outputs.values()].find(o => o.id === selectedId) ?? null;
    });

    if (midiAccess.outputs.size > 0) {
      midiOutput = [...midiAccess.outputs.values()][0];
      outputSelect.value = midiOutput.id;
    }
  } catch (err) {
    console.error('MIDI init error', err);
  }

  // Build pad grid
  for (const pad of GM_DRUM_PADS) {
    const button = document.createElement('button');
    button.className = 'pad';
    button.textContent = pad.name;
    button.addEventListener('click', () => sendMidiNote(pad.note));
    padGrid.appendChild(button);
  }
}

init();
