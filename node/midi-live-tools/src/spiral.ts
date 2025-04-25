// --- Constants and Setup ---
// Base radius for the spiral
const BASE_RADIUS = 15;
// Distance between successive spiral arms
const SPACING = 10;
// Angle increment per note (radians)
const ANGLE_STEP = 0.35;
// Blur radius for shadow effect on active notes
const SHADOW_BLUR = 15;

// MIDI note range (A0 to C8)
const MIDI_START = 21; // A0
const MIDI_END = 108;  // C8

// Get canvas and context, set initial size to 80% of window
const canvas = document.getElementById('spiral-canvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;
canvas.width = window.innerWidth * 0.8;
canvas.height = window.innerHeight * 0.8;

// --- State ---
// Set of currently active MIDI notes
const activeNotes = new Set<number>();
// Map to track decay values for notes (for fading effect)
const noteDecay = new Map<number, number>();
// Flag to enable or disable key highlighting
let enableKeyHighlighting = true;
// Set of notes (0-11) that belong to the current key
const keyNotes = new Set<number>();
// Set of notes permanently highlighted due to key membership and activation
const permanentlyHighlightedNotes = new Set<number>();

// --- Key Definitions ---
// Musical roots for selection
const roots = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
// DOM elements for key selection UI
const rootSlider = document.getElementById('rootSlider') as HTMLInputElement;
const rootLabel = document.getElementById('rootLabel')!;
const modeSelect = document.getElementById('modeSelect') as HTMLSelectElement;

// --- Helper Functions ---

/**
 * Sets the current musical key by root and mode.
 * Clears previous key notes and permanently highlighted notes,
 * then calculates the new key notes based on major/minor pattern.
 * @param root - Root note as string (e.g., 'C', 'F#')
 * @param mode - Mode string ('Major' or 'Minor')
 */
function setKey(root: string, mode: string) {
  keyNotes.clear();
  permanentlyHighlightedNotes.clear();

  // Major and minor scale intervals in semitones
  const majorPattern = [0, 2, 4, 5, 7, 9, 11];
  const minorPattern = [0, 2, 3, 5, 7, 8, 10];

  // Map root note names to semitone offsets
  const rootSemitones: { [key: string]: number } = {
    'C': 0, 'C#': 1, 'D': 2, 'D#': 3, 'E': 4, 'F': 5,
    'F#': 6, 'G': 7, 'G#': 8, 'A': 9, 'A#': 10, 'B': 11
  };

  const rootOffset = rootSemitones[root];
  const pattern = mode === 'Minor' ? minorPattern : majorPattern;

  // Add all notes in the key to the set (mod 12 for octave equivalence)
  pattern.forEach(interval => {
    keyNotes.add((rootOffset + interval) % 12);
  });
}

/**
 * Updates the current key based on UI controls.
 * Updates the root label and recalculates key notes.
 */
function updateKey() {
  const root = roots[parseInt(rootSlider.value)];
  const mode = modeSelect.value;
  rootLabel.textContent = root;
  setKey(root, mode);
  // Persist root and mode to localStorage
  localStorage.setItem('root', rootSlider.value);
  localStorage.setItem('mode', modeSelect.value);
}

// --- MIDI Setup ---

/**
 * Initializes MIDI access and sets up event handlers for MIDI input.
 * On receiving MIDI messages, calls handleMIDIMessage.
 */
async function initMIDI() {
  try {
    const access = await navigator.requestMIDIAccess();
    // Attach message handler to each MIDI input device
    for (const input of access.inputs.values()) {
      input.onmidimessage = handleMIDIMessage;
    }
    // Reattach handlers if MIDI devices change state
    access.onstatechange = () => {
      for (const input of access.inputs.values()) {
        input.onmidimessage = handleMIDIMessage;
      }
    };
  } catch (err) {
    console.error("MIDI init error", err);
  }
}

/**
 * Handles incoming MIDI messages.
 * Tracks note on/off events and updates active notes and decay.
 * Also manages permanent highlighting for key notes when enabled.
 * @param e - MIDIMessageEvent from MIDI input
 */
function handleMIDIMessage(e: MIDIMessageEvent) {
  if (!e.data || e.data.length < 3) return;

  const [status, note, velocity] = e.data;
  const cmd = status & 0xf0;

  if (cmd === 0x90 && velocity > 0) { // Note on
    activeNotes.add(note);
    noteDecay.set(note, 1.0);

    if (enableKeyHighlighting && keyNotes.has(note % 12)) {
      permanentlyHighlightedNotes.add(note);
    }
  } else if (cmd === 0x80 || (cmd === 0x90 && velocity === 0)) { // Note off
    activeNotes.delete(note);
  }
}

// --- Drawing ---

/**
 * Draws the spiral of notes on the canvas.
 * Each MIDI note is represented as a circle positioned along a spiral.
 * Active notes animate with pulsing and glow effects.
 * Notes belonging to the current key can be highlighted.
 * The function continuously requests animation frames.
 */
function drawSpiral() {
  const centerX = canvas.width / 2;
  const centerY = canvas.height / 2;

  // Clear canvas with a translucent white background for fade effect
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = 'rgba(255, 255, 255, 0.15)';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  const time = performance.now();

  for (let midiNote = MIDI_START; midiNote <= MIDI_END; midiNote++) {
    const i = midiNote - MIDI_START;
    const angle = i * ANGLE_STEP;
    const radius = BASE_RADIUS + SPACING * angle;

    // Calculate spiral coordinates
    const x = centerX + Math.cos(angle) * radius;
    const y = centerY + Math.sin(angle) * radius;

    // Hue based on note within octave
    const hue = (midiNote % 12) * 30;

    // Get current decay value and reduce it gradually for fade out
    const decay = noteDecay.get(midiNote) ?? 0;
    const newDecay = Math.max(0, decay - 0.010);
    noteDecay.set(midiNote, newDecay);

    // Determine if note is permanently highlighted (key member and activated)
    const isPermanentlyHighlighted = permanentlyHighlightedNotes.has(midiNote);
    // Note is active if decay > 0 or permanently highlighted
    const isActive = decay > 0 || isPermanentlyHighlighted;

    // Base pulse size grows slightly with note index for visual variation
    const pulse = 2 + i * 0.15;
    // Animate pulse size with sine wave if active
    const animatedPulse = isActive
      ? pulse + Math.sin(time / 1000 + midiNote) * 0.5
      : pulse;

    // Begin drawing the note circle
    ctx.beginPath();
    ctx.shadowColor = isActive ? `hsl(${hue}, 100%, 70%)` : 'transparent';
    ctx.shadowBlur = isActive ? SHADOW_BLUR : 0;
    ctx.arc(x, y, animatedPulse, 0, Math.PI * 2);

    // Base colors for notes
    const baseColor = '238, 238, 238'; // light grey
    const darkGrey = '200, 200, 200';  // darker grey for key members

    // Set fill style based on note state
    if (isActive && decay > 0.2) {
      // Active notes fade out with decay alpha
      ctx.fillStyle = `hsla(${hue}, 100%, 60%, ${decay})`;
    } else if (enableKeyHighlighting && keyNotes.has(midiNote % 12)) {
      // Key notes get darker grey fill
      ctx.fillStyle = `rgba(${darkGrey}, 1)`;
    } else {
      // Default fill color for other notes
      ctx.fillStyle = `rgba(${baseColor}, 1)`;
    }

    ctx.fill();
    ctx.shadowBlur = 0;
  }

  // Request next frame for continuous animation
  requestAnimationFrame(drawSpiral);
}

// --- Initialize Everything ---

// Update canvas size and clear active notes on window resize
window.addEventListener('resize', () => {
  canvas.width = window.innerWidth * 0.8;
  canvas.height = window.innerHeight * 0.8;
  activeNotes.clear();
});

// Attach event listeners to UI controls for key changes
rootSlider.addEventListener('input', updateKey);
modeSelect.addEventListener('change', updateKey);

// Highlight toggle and controls fieldset
const highlightToggle = document.getElementById('highlightToggle') as HTMLInputElement;
const controlsFieldset = document.getElementById('controlsFieldset') as HTMLFieldSetElement;

// Add event listener for highlight toggle, with persistence
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

// Initial setup calls
updateKey();
initMIDI();
drawSpiral();
