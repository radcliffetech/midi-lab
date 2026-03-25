import Chart from 'chart.js/auto';
import { detectKey } from './lib/midi-utils';
import { parseMidiMessage } from './lib/midi-message';
import { aggregatePitchClasses } from './lib/music-theory';

const ctx = document.getElementById('noteChart') as HTMLCanvasElement;
const noteCounts = new Array(128).fill(0);
const labels = Array.from({ length: 128 }, (_, i) => `MIDI ${i}`);

const chart = new Chart(ctx, {
    type: 'bar',
    data: {
        labels,
        datasets: [{
            label: 'Note Count',
            data: noteCounts,
            backgroundColor: 'rgba(54, 162, 235, 0.6)',
        }]
    },
    options: {
        responsive: true,
        scales: {
            x: {
                ticks: { autoSkip: true, maxTicksLimit: 24 }
            },
            y: {
                beginAtZero: true
            }
        },
        plugins: {
            legend: {
                display: false
            }
        }
    }
});

function handleMIDIMessage(e: MIDIMessageEvent) {
    const msg = parseMidiMessage(e.data);
    if (!msg || msg.command !== 'note_on') return;

    noteCounts[msg.note]++;
    chart.update();

    const pitchClassCounts = aggregatePitchClasses(noteCounts);
    const { key, confidence } = detectKey(pitchClassCounts);

    document.getElementById('keyOutput')!.textContent = key;
    document.getElementById('confidenceOutput')!.textContent = confidence.toFixed(2);
    document.getElementById('pitchClassesOutput')!.textContent = `[${pitchClassCounts.join(', ')}]`;
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

document.getElementById('resetButton')?.addEventListener('click', () => {
    for (let i = 0; i < noteCounts.length; i++) {
        noteCounts[i] = 0;
    }
    chart.update();
    document.getElementById('keyOutput')!.textContent = '–';
    document.getElementById('pitchClassesOutput')!.textContent = '–';
});

document.getElementById('copyButton')?.addEventListener('click', () => {
  const pitchText = document.getElementById('pitchClassesOutput')?.textContent ?? '';
  navigator.clipboard.writeText(pitchText).then(() => {
    console.log('Pitch class array copied to clipboard.');
  }).catch(err => {
    console.error('Failed to copy pitch class array:', err);
  });
});
