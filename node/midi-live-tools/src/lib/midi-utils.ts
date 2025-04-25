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


export function getNoteName(n: number): string {
  return `${midiNoteName(n)} (${n})`;
}
function correlate(a: number[], b: number[]): number {
  const meanA = a.reduce((sum, v) => sum + v, 0) / a.length;
  const meanB = b.reduce((sum, v) => sum + v, 0) / b.length;

  let numerator = 0;
  let denomA = 0;
  let denomB = 0;

  for (let i = 0; i < a.length; i++) {
    const devA = a[i] - meanA;
    const devB = b[i] - meanB;
    numerator += devA * devB;
    denomA += devA * devA;
    denomB += devB * devB;
  }

  return numerator / Math.sqrt(denomA * denomB);
}

export function detectKey(pitchClassCounts: number[]): { key: string, confidence: number } {
  const majorProfile = [6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88];
  const minorProfile = [6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17];

  let bestScore = -Infinity;
  let bestKey = '';
  const names = ['C','C#','D','D#','E','F','F#','G','G#','A','A#','B'];

  for (let i = 0; i < 12; i++) {
    const rotateInput = (arr: number[]) => arr.map((_, j) => arr[(j - i + 12) % 12]);

    const majorScore = correlate(rotateInput(pitchClassCounts), majorProfile);
    if (majorScore > bestScore) {
      bestScore = majorScore;
      bestKey = `${names[i]} Major`;
    }

    const minorScore = correlate(rotateInput(pitchClassCounts), minorProfile);
    if (minorScore > bestScore) {
      bestScore = minorScore;
      bestKey = `${names[i]} Minor`;
    }
  }

  return { key: bestKey, confidence: bestScore };
}