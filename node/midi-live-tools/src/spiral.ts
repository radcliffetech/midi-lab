// spiral.ts
const BASE_RADIUS = 15;
const SPACING = 10;
const ANGLE_STEP = 0.35;
const DEFAULT_PULSE = 6.4;
const ACTIVE_PULSE_AMPLITUDE = 2;
const SHADOW_BLUR = 15;

const MIDI_START = 21; // A0
const MIDI_END = 108;  // C8

const canvas = document.getElementById('spiral-canvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;
canvas.width = window.innerWidth * 0.8;
canvas.height = window.innerHeight * 0.8;

const activeNotes = new Set<number>();
const noteDecay = new Map<number, number>();

let enableKeyHighlighting = true;
const keyNotes = new Set<number>();
const permanentlyHighlightedNotes = new Set<number>();

function setKey(keyName: string) {
  keyNotes.clear();
  permanentlyHighlightedNotes.clear();

  const keyMappings: { [key: string]: number[] } = {
    'C Minor': [0, 2, 3, 5, 7, 8, 10],
    'C Major': [0, 2, 4, 5, 7, 9, 11],
    'E Major': [4, 6, 8, 9, 11, 1, 3],
    'G Major': [7, 9, 11, 0, 2, 4, 6],
    'A Minor': [9, 11, 0, 2, 4, 5, 7],
  };

  const notes = keyMappings[keyName];
  if (notes) {
    notes.forEach(pc => keyNotes.add(pc));
  }
}

// Recalculate on resize
window.addEventListener('resize', () => {
  canvas.width = window.innerWidth * 0.8;
  canvas.height = window.innerHeight * 0.8;
  activeNotes.clear();
});

function drawSpiral() {
  const centerX = canvas.width / 2;
  const centerY = canvas.height / 2;

  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = 'rgba(255, 255, 255, 0.15)';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  const time = performance.now();

  for (let midiNote = MIDI_START; midiNote <= MIDI_END; midiNote++) {
    const i = midiNote - MIDI_START;
    const angle = i * ANGLE_STEP;
    const radius = BASE_RADIUS + SPACING * angle;

    const x = centerX + Math.cos(angle) * radius;
    const y = centerY + Math.sin(angle) * radius;

    const hue = (midiNote % 12) * 30;

    const decay = noteDecay.get(midiNote) ?? 0;
    const newDecay = Math.max(0, decay - 0.010);
    noteDecay.set(midiNote, newDecay);

    const isPermanentlyHighlighted = permanentlyHighlightedNotes.has(midiNote);
    const isActive = decay > 0 || isPermanentlyHighlighted;
    const pulse = 2 + i * 0.15;
    const animatedPulse = isActive
      ? pulse + Math.sin(time / 1000 + midiNote) * 0.5
      : pulse;

    ctx.beginPath();
    ctx.shadowColor = isActive ? `hsl(${hue}, 100%, 70%)` : 'transparent';
    ctx.shadowBlur = isActive ? SHADOW_BLUR : 0;
    ctx.arc(x, y, animatedPulse, 0, Math.PI * 2);

    const baseColor = '238, 238, 238'; // light grey
    const darkGrey = '200, 200, 200';  // darker for key member placeholders

    ctx.fillStyle = isActive
      ? `hsla(${hue}, 100%, 60%, ${decay})`
      : (enableKeyHighlighting && keyNotes.has(midiNote % 12))
        ? `rgba(${darkGrey}, 1)`
        : `rgba(${baseColor}, 1)`;

    ctx.fill();
    ctx.shadowBlur = 0;
  }

  requestAnimationFrame(drawSpiral);
}

function handleMIDIMessage(e: MIDIMessageEvent) {
    if (!e.data || e.data.length < 3) return;
  const [status, note, velocity] = e.data;
  const cmd = status & 0xf0;

  if (cmd === 0x90 && velocity > 0) {
    activeNotes.add(note);
    noteDecay.set(note, 1.0);

    if (enableKeyHighlighting && keyNotes.has(note % 12)) {
      permanentlyHighlightedNotes.add(note);
    }
  } else if (cmd === 0x80 || (cmd === 0x90 && velocity === 0)) {
    activeNotes.delete(note);
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
    console.error("MIDI init error", err);
  }
}

setKey('C Minor');

initMIDI();

const highlightToggle = document.getElementById('highlightToggle') as HTMLInputElement;
highlightToggle.addEventListener('change', () => {
  enableKeyHighlighting = highlightToggle.checked;
});

const keySelect = document.getElementById('keySelect') as HTMLSelectElement;
keySelect.addEventListener('change', () => {
  setKey(keySelect.value);
});

document.getElementById('resetButton')?.addEventListener('click', () => {
  noteDecay.clear();
});
drawSpiral();