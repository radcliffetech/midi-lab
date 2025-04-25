// src/lib/midi-utils.ts

export function midiNoteName(n: number): string {
  const names = [
    'C',
    'C#',
    'D',
    'D#',
    'E',
    'F',
    'F#',
    'G',
    'G#',
    'A',
    'A#',
    'B',
  ];
  return names[n % 12] + Math.floor(n / 12 - 1);
}

export function detectChord(notes: number[]): string | null {
  if (notes.length < 3) return null;
  const pitchClasses = [...new Set(notes.map((n) => n % 12))].sort(
    (a, b) => a - b
  );

  const chordTypes: [string, number[]][] = [
    ['Major', [0, 4, 7]],
    ['Minor', [0, 3, 7]],
    ['Diminished', [0, 3, 6]],
    ['Augmented', [0, 4, 8]],
    ['Sus2', [0, 2, 7]],
    ['Sus4', [0, 5, 7]],
    ['Major 7', [0, 4, 7, 11]],
    ['Minor 7', [0, 3, 7, 10]],
    ['Dominant 7', [0, 4, 7, 10]],
    ['Half-diminished 7', [0, 3, 6, 10]],
  ];

  for (let root = 0; root < 12; root++) {
    const rotated = pitchClasses
      .map((pc) => (pc - root + 12) % 12)
      .sort((a, b) => a - b);
    for (const [name, shape] of chordTypes) {
      if (shape.every((x) => rotated.includes(x))) {
        return `${midiNoteName(root)} ${name}`;
      }
    }
  }

  return null;
}
