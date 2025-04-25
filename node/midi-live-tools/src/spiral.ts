// spiral.ts
const canvas = document.getElementById('spiral-canvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;
canvas.width = window.innerWidth * 0.8;
canvas.height = window.innerHeight * 0.8;

const activeNotes = new Set<number>();

// Recalculate on resize
window.addEventListener('resize', () => {
  canvas.width = window.innerWidth * 0.8;
  canvas.height = window.innerHeight * 0.8;
});

function drawSpiral() {
  const centerX = canvas.width / 2;
  const centerY = canvas.height / 2;
  const baseRadius = 15;
  const spacing = 10;

  ctx.clearRect(0, 0, canvas.width, canvas.height);

  for (let i = 0; i < 88; i++) {
    const note = i + 21;
    const angle = i * 0.35;
    const radius = baseRadius + spacing * angle;

    const x = centerX + Math.cos(angle) * radius;
    const y = centerY + Math.sin(angle) * radius;

    ctx.beginPath();
    ctx.arc(x, y, 8, 0, Math.PI * 2);
    ctx.fillStyle = activeNotes.has(note) ? '#ff3366' : '#cccccc';
    ctx.fill();
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