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

    const isActive = activeNotes.has(midiNote);
    const pulse = isActive
      ? Math.sin(time / 100 + midiNote) * ACTIVE_PULSE_AMPLITUDE + DEFAULT_PULSE
      : DEFAULT_PULSE;

    ctx.beginPath();
    ctx.shadowColor = isActive ? `hsl(${hue}, 100%, 70%)` : 'transparent';
    ctx.shadowBlur = isActive ? SHADOW_BLUR : 0;
    ctx.arc(x, y, pulse, 0, Math.PI * 2);
    ctx.fillStyle = isActive ? `hsl(${hue}, 100%, 60%)` : '#cccccc';
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

initMIDI();
drawSpiral();