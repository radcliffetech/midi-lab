import { getScaleNotes, aggregatePitchClasses, noteNameToPitchClass } from './music-theory';

describe('noteNameToPitchClass', () => {
    test('maps all 12 note names correctly', () => {
        expect(noteNameToPitchClass('C')).toBe(0);
        expect(noteNameToPitchClass('C#')).toBe(1);
        expect(noteNameToPitchClass('D')).toBe(2);
        expect(noteNameToPitchClass('D#')).toBe(3);
        expect(noteNameToPitchClass('E')).toBe(4);
        expect(noteNameToPitchClass('F')).toBe(5);
        expect(noteNameToPitchClass('F#')).toBe(6);
        expect(noteNameToPitchClass('G')).toBe(7);
        expect(noteNameToPitchClass('G#')).toBe(8);
        expect(noteNameToPitchClass('A')).toBe(9);
        expect(noteNameToPitchClass('A#')).toBe(10);
        expect(noteNameToPitchClass('B')).toBe(11);
    });

    test('returns -1 for unrecognized names', () => {
        expect(noteNameToPitchClass('Cb')).toBe(-1);
        expect(noteNameToPitchClass('X')).toBe(-1);
        expect(noteNameToPitchClass('')).toBe(-1);
    });
});

describe('getScaleNotes', () => {
    test('returns C Major scale notes', () => {
        const notes = getScaleNotes('C', 'Major');
        expect(notes).toEqual(new Set([0, 2, 4, 5, 7, 9, 11]));
    });

    test('returns C Minor scale notes', () => {
        const notes = getScaleNotes('C', 'Minor');
        expect(notes).toEqual(new Set([0, 2, 3, 5, 7, 8, 10]));
    });

    test('returns F# Major scale notes (wrapping)', () => {
        // F# = 6, pattern: 6, 8, 10, 11, 1, 3, 5
        const notes = getScaleNotes('F#', 'Major');
        expect(notes).toEqual(new Set([6, 8, 10, 11, 1, 3, 5]));
    });

    test('returns A Minor scale notes', () => {
        // A = 9, minor pattern: 9, 11, 0, 2, 4, 5, 7
        const notes = getScaleNotes('A', 'Minor');
        expect(notes).toEqual(new Set([9, 11, 0, 2, 4, 5, 7]));
    });

    test('returns 7 pitch classes for any valid key', () => {
        const allRoots = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'];
        for (const root of allRoots) {
            expect(getScaleNotes(root, 'Major').size).toBe(7);
            expect(getScaleNotes(root, 'Minor').size).toBe(7);
        }
    });

    test('returns empty set for unrecognized root', () => {
        expect(getScaleNotes('X', 'Major').size).toBe(0);
    });
});

describe('aggregatePitchClasses', () => {
    test('returns all zeros for empty note counts', () => {
        const counts = new Array(128).fill(0);
        expect(aggregatePitchClasses(counts)).toEqual(new Array(12).fill(0));
    });

    test('aggregates single note correctly', () => {
        const counts = new Array(128).fill(0);
        counts[60] = 5; // C4 -> pitch class 0
        const result = aggregatePitchClasses(counts);
        expect(result[0]).toBe(5);
    });

    test('aggregates notes across octaves', () => {
        const counts = new Array(128).fill(0);
        counts[60] = 3; // C4 -> pitch class 0
        counts[72] = 7; // C5 -> pitch class 0
        counts[48] = 2; // C3 -> pitch class 0
        const result = aggregatePitchClasses(counts);
        expect(result[0]).toBe(12); // 3 + 7 + 2
    });

    test('distributes across all 12 pitch classes', () => {
        const counts = new Array(128).fill(0);
        // Set one note per pitch class in octave 4 (MIDI 60-71)
        for (let i = 60; i < 72; i++) {
            counts[i] = 1;
        }
        const result = aggregatePitchClasses(counts);
        expect(result).toEqual(new Array(12).fill(1));
    });

    test('handles empty array', () => {
        const result = aggregatePitchClasses([]);
        expect(result).toEqual(new Array(12).fill(0));
    });
});
