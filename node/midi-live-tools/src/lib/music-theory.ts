import { NOTE_NAMES } from './midi-constants';

const MAJOR_PATTERN = [0, 2, 4, 5, 7, 9, 11];
const MINOR_PATTERN = [0, 2, 3, 5, 7, 8, 10];

/**
 * Map a note name (e.g. 'C', 'F#') to its pitch class number (0-11).
 * Returns -1 for unrecognized names.
 */
export function noteNameToPitchClass(name: string): number {
  const index = NOTE_NAMES.indexOf(name as typeof NOTE_NAMES[number]);
  return index;
}

/**
 * Get the set of pitch classes (0-11) belonging to a given key.
 * @param root - Root note name (e.g. 'C', 'F#')
 * @param mode - 'Major' or 'Minor'
 */
export function getScaleNotes(root: string, mode: 'Major' | 'Minor'): Set<number> {
  const rootOffset = noteNameToPitchClass(root);
  if (rootOffset < 0) return new Set();

  const pattern = mode === 'Minor' ? MINOR_PATTERN : MAJOR_PATTERN;
  const notes = new Set<number>();
  for (const interval of pattern) {
    notes.add((rootOffset + interval) % 12);
  }
  return notes;
}

/**
 * Aggregate an array of 128 MIDI note counts into a 12-element pitch class histogram.
 */
export function aggregatePitchClasses(noteCounts: number[]): number[] {
  const pitchClassCounts = new Array(12).fill(0);
  for (let i = 0; i < noteCounts.length; i++) {
    pitchClassCounts[i % 12] += noteCounts[i];
  }
  return pitchClassCounts;
}
