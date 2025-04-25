const select = document.getElementById('midi-select') as HTMLSelectElement;
const startStopBtn = document.getElementById('start-stop') as HTMLButtonElement;
const statusDiv = document.getElementById('status')!;

function log(device: string, message: string) {
  const tbody = document.querySelector('#output tbody')!;
  const row = document.createElement('tr');
  const deviceCell = document.createElement('td');
  const msgCell = document.createElement('td');
  const timeCell = document.createElement('td');

  deviceCell.textContent = device;
  msgCell.textContent = message;
  timeCell.textContent = new Date().toLocaleTimeString();

  row.appendChild(deviceCell);
  row.appendChild(msgCell);
  row.appendChild(timeCell);
  tbody.insertBefore(row, tbody.firstChild);

  // Fade out after 5 seconds
  setTimeout(() => {
    row.classList.add('fade-out');
  }, 5000);

  // Remove after 10 seconds
  setTimeout(() => {
    row.remove();
  }, 10000);
}

let currentInput: MIDIInput | null = null;

async function initMIDI() {
  try {
    const access = await navigator.requestMIDIAccess();
    log("system", "✅ MIDI access granted");

    populateInputOptions(access);

    access.onstatechange = (e) => {
      log("system", `🔌 Device ${e.port.name} ${e.port.state}`);
      populateInputOptions(access);
    };

    select.addEventListener('change', () => {
      if (currentInput) {
        currentInput.onmidimessage = null;
        currentInput = null;
        updateStatus('');
      }

      const selectedId = select.value;
      const newInput = Array.from(access.inputs.values()).find(i => i.id === selectedId);

      if (newInput) {
        currentInput = newInput;
        updateStatus(`Ready to start: ${currentInput.name}`);
      }
    });
  } catch (err) {
    log("system", "❌ Could not access MIDI devices");
    console.error(err);
  }
}

function populateInputOptions(access: MIDIAccess) {
  const inputs = Array.from(access.inputs.values());
  select.innerHTML = '';

  if (inputs.length === 0) {
    const opt = document.createElement('option');
    opt.text = '(no inputs found)';
    opt.disabled = true;
    select.appendChild(opt);
    return;
  }

  inputs.forEach((input, index) => {
    const opt = document.createElement('option');
    opt.value = input.id;
    opt.text = input.name ?? '(unnamed device)';
    if (index === 0) opt.selected = true;
    select.appendChild(opt);
  });

  // Auto-select and register first input if available
  if (inputs.length > 0) {
    selectInputById(inputs[0].id, access);
  }
}

function selectInputById(id: string, access: MIDIAccess) {
  const input = Array.from(access.inputs.values()).find(i => i.id === id);
  if (input) {
    currentInput = input;
    updateStatus(`Ready to start: ${input.name}`);
  }
}

function updateStatus(msg: string) {
  statusDiv.textContent = msg;
}

let isRunning = false;

startStopBtn.addEventListener('click', () => {
  if (!currentInput) {
    updateStatus("⚠️ No MIDI input selected");
    return;
  }

  if (!isRunning) {
    startStopBtn.textContent = "Stop";
    updateStatus(`🎛 Listening to: ${currentInput.name}`);
    currentInput.onmidimessage = (e) => {
      const [status, data1, data2] = e.data;
      const command = status & 0xf0;
      let noteInfo = '';

      if (command === 0x90 && data2 > 0) {
        noteInfo = `Note On ${getNoteName(data1)} (velocity: ${data2})`;
      } else if (command === 0x80 || (command === 0x90 && data2 === 0)) {
        noteInfo = `Note Off ${getNoteName(data1)}`;
      }

      const logMessage = noteInfo
        ? `🎶 ${noteInfo}`
        : `Status: ${status.toString(16)}, Data1: ${data1}, Data2: ${data2}`;

      log(currentInput?.name || 'unknown', logMessage);
    };
    isRunning = true;
  } else {
    startStopBtn.textContent = "Start";
    updateStatus(`Paused: ${currentInput.name}`);
    if (currentInput) {
      currentInput.onmidimessage = null;
    }
    isRunning = false;
  }
});

initMIDI();
function getNoteName(noteNumber: number): string {
  const notes = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
  const note = notes[noteNumber % 12];
  const octave = Math.floor(noteNumber / 12) - 1;
  return `${note}${octave}`;
}