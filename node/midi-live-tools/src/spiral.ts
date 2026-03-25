import { NOTE_NAMES, PIANO_MIDI_START, PIANO_MIDI_END } from './lib/midi-constants';
import { parseMidiMessage } from './lib/midi-message';
import { getScaleNotes } from './lib/music-theory';

// --- Constants and Setup ---
const BASE_RADIUS = 15;
const SPACING = 10;
const ANGLE_STEP = 0.35;
const SHADOW_BLUR = 15;

const canvas = document.getElementById('spiral-canvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;
canvas.width = window.innerWidth * 0.8;
canvas.height = window.innerHeight * 0.8;

// --- State ---
const activeNotes = new Set<number>();
const noteDecay = new Map<number, number>();
let enableKeyHighlighting = true;
let keyNotes = new Set<number>();
const permanentlyHighlightedNotes = new Set<number>();

// --- Key Definitions ---
const rootSlider = document.getElementById('rootSlider') as HTMLInputElement;
const rootLabel = document.getElementById('rootLabel')!;
const modeSelect = document.getElementById('modeSelect') as HTMLSelectElement;

/**
 * Sets the current musical key by root and mode.
 * Clears previous key notes and permanently highlighted notes,
 * then calculates the new key notes based on major/minor pattern.
 */
function setKey(root: string, mode: string) {
  permanentlyHighlightedNotes.clear();
  keyNotes = getScaleNotes(root, mode as 'Major' | 'Minor');
}

function updateKey() {
  const root = NOTE_NAMES[parseInt(rootSlider.value)];
  const mode = modeSelect.value;
  rootLabel.textContent = root;
  setKey(root, mode);
  localStorage.setItem('root', rootSlider.value);
  localStorage.setItem('mode', modeSelect.value);
}

// --- MIDI Setup ---

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

function handleMIDIMessage(e: MIDIMessageEvent) {
  const msg = parseMidiMessage(e.data);
  if (!msg) return;

  if (msg.command === 'note_on') {
    activeNotes.add(msg.note);
    noteDecay.set(msg.note, 1.0);
    if (enableKeyHighlighting && keyNotes.has(msg.note % 12)) {
      permanentlyHighlightedNotes.add(msg.note);
    }
  } else if (msg.command === 'note_off') {
    activeNotes.delete(msg.note);
  }
}

// --- Drawing ---

function drawSpiral() {
  const centerX = canvas.width / 2;
  const centerY = canvas.height / 2;

  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = 'rgba(255, 255, 255, 0.15)';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  const time = performance.now();

  for (let midiNote = PIANO_MIDI_START; midiNote <= PIANO_MIDI_END; midiNote++) {
    const i = midiNote - PIANO_MIDI_START;
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

    const baseColor = '238, 238, 238';
    const darkGrey = '200, 200, 200';

    if (isActive && decay > 0.2) {
      ctx.fillStyle = `hsla(${hue}, 100%, 60%, ${decay})`;
    } else if (enableKeyHighlighting && keyNotes.has(midiNote % 12)) {
      ctx.fillStyle = `rgba(${darkGrey}, 1)`;
    } else {
      ctx.fillStyle = `rgba(${baseColor}, 1)`;
    }

    ctx.fill();
    ctx.shadowBlur = 0;
  }

  requestAnimationFrame(drawSpiral);
}

// --- Initialize Everything ---

window.addEventListener('resize', () => {
  canvas.width = window.innerWidth * 0.8;
  canvas.height = window.innerHeight * 0.8;
  activeNotes.clear();
});

rootSlider.addEventListener('input', updateKey);
modeSelect.addEventListener('change', updateKey);

const highlightToggle = document.getElementById('highlightToggle') as HTMLInputElement;
const controlsFieldset = document.getElementById('controlsFieldset') as HTMLFieldSetElement;

highlightToggle.addEventListener('change', () => {
  enableKeyHighlighting = highlightToggle.checked;
  if (controlsFieldset) controlsFieldset.disabled = !highlightToggle.checked;
  localStorage.setItem('highlight', enableKeyHighlighting ? 'true' : 'false');
});

// --- Restore persisted settings before first updateKey() ---
const savedRoot = localStorage.getItem('root');
const savedMode = localStorage.getItem('mode');
const savedHighlight = localStorage.getItem('highlight');

if (savedRoot !== null) rootSlider.value = savedRoot;
if (savedMode !== null) modeSelect.value = savedMode;
if (savedHighlight !== null) {
  highlightToggle.checked = (savedHighlight === 'true');
  enableKeyHighlighting = highlightToggle.checked;
  if (controlsFieldset) controlsFieldset.disabled = !highlightToggle.checked;
}

updateKey();
initMIDI();
drawSpiral();
