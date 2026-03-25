/** Chromatic note names, indexed by pitch class (0 = C, 11 = B). */
export const NOTE_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'] as const;

/** First MIDI note on a standard 88-key piano (A0). */
export const PIANO_MIDI_START = 21;

/** Last MIDI note on a standard 88-key piano (C8). */
export const PIANO_MIDI_END = 108;

/** Pitch classes that correspond to black keys on the piano. */
export const BLACK_KEY_PITCH_CLASSES = [1, 3, 6, 8, 10] as const;
