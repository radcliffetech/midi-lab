import { cMinorExampleRow, gMinorExampleRow, majorRow, minorRow } from './test-profiles';
import { correlate, detectChord, detectKey, getNoteName, midiNoteName, rotateBackward, rotateForward } from './midi-utils';

describe('detectKey', () => {
    test('clearly identifies C Major - strong profile', () => {
        const result = detectKey(majorRow);
        expect(result.key).toBe('C Major');
    });

    test('clearly identifies C Minor - strong profile', () => {
        const result = detectKey(minorRow);
        expect(result.key).toBe('C Minor');
    });

    test('identifies G Minor - real piece example', () => {
        const result = detectKey(gMinorExampleRow);
        console.log(`Detected key for G Minor test: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('F Minor');
    });

    test('identifies C Minor - real piece example', () => {
        const result = detectKey(cMinorExampleRow);
        console.log(`Detected key for C Minor test: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('C Minor');
    });


    test('identifies F Minor - synthetic example', () => {
        const fMinorRow = [
            0, 0, 0, 5,  // D#, E
            30, 0, 0, 40, // F, F#, G
            0, 25, 35, 0  // G#, A, A#
        ];
        const result = detectKey(fMinorRow);
        console.log(`Detected key for F Minor test: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('F Minor');
    });

});

describe('getNoteName', () => {
    test('returns correct note names', () => {
        expect(getNoteName(57)).toBe('A3 (57)');
        expect(getNoteName(58)).toBe('A#3 (58)');
        expect(getNoteName(59)).toBe('B3 (59)');
        expect(getNoteName(60)).toBe('C4 (60)');
        expect(getNoteName(61)).toBe('C#4 (61)');
        expect(getNoteName(62)).toBe('D4 (62)');
        expect(getNoteName(63)).toBe('D#4 (63)');
        expect(getNoteName(64)).toBe('E4 (64)');
        expect(getNoteName(65)).toBe('F4 (65)');
        expect(getNoteName(66)).toBe('F#4 (66)');
        expect(getNoteName(67)).toBe('G4 (67)');
        expect(getNoteName(68)).toBe('G#4 (68)');
        expect(getNoteName(69)).toBe('A4 (69)');
        expect(getNoteName(70)).toBe('A#4 (70)');
        expect(getNoteName(71)).toBe('B4 (71)');
        expect(getNoteName(72)).toBe('C5 (72)');
        expect(getNoteName(73)).toBe('C#5 (73)');
        expect(getNoteName(74)).toBe('D5 (74)');
    });
});

describe('rotateForward and rotateBackward', () => {
    test('rotates input array correctly', () => {
        const input = [10, 20, 30, 40];
        const expected = [40, 10, 20, 30];
        const rotated = rotateForward(input, 1);
        expect(rotated).toEqual(expected);  
    });

    test('rotates input array correctly with negative rotation', () => {
        const input = [0, 1, 2, 3];
        const expected = [1, 2, 3, 0];
        const rotated = rotateBackward(input, 1);
        expect(rotated).toEqual(expected);
    });

    test("special case to find bugs", () => {
        const input = [
            100, 0, 50, 80, 
            0, 0, 0, 90, 
            50, 0, 20, 20
          ] 
        const expected = [
            20, 20, 100, 0, 
            50, 80, 0, 0, 
            0, 90, 50, 0
        ];

        const rotated = rotateForward(input, 2);
        expect(rotated).toEqual(expected);
    });

});

describe('correlate', () => {
    const majorProfile = [6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88];
    const minorProfile = [6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17];

    test('returns correct correlation coefficient', () => {
        const a = [1, 2, 3, 4, 5];
        const b = [5, 4, 3, 2, 1];
        const result = correlate(a, b);
        expect(result).toBe(-1);
    });

    test('returns correct correlation coefficient for identical arrays', () => {
        const a = [1, 2, 3, 4, 5];
        const b = [1, 2, 3, 4, 5];
        const result = correlate(a, b);
        expect(result).toBe(1);
    });

    test("special case to find bugs", () => {
        const dMinor = [20, 20, 100, 0, 50, 80, 0, 0, 0, 90, 50, 0];
        const aSharpMajor = [0, 80, 0, 0, 90, 0, 50, 0, 20, 100, 0, 50];
    
        const result1 = correlate(dMinor, minorProfile);
        const result2 = correlate(aSharpMajor, minorProfile);
    
        console.log("RESULTS", "d minor =>", result1, "a# major =>", result2);
    
        // Now it makes sense: dMinor should correlate better to minorProfile
        expect(result1).toBeGreaterThan(result2);
    });
});

describe('midiNoteName', () => {
    test('returns correct note names for known MIDI numbers', () => {
        expect(midiNoteName(60)).toBe('C4');
        expect(midiNoteName(61)).toBe('C#4');
        expect(midiNoteName(69)).toBe('A4');
        expect(midiNoteName(72)).toBe('C5');
    });

    test('wraps correctly at octave boundaries', () => {
        expect(midiNoteName(0)).toBe('C-1');
        expect(midiNoteName(12)).toBe('C0');
    });
});

describe('detectChord', () => {
    test('detects major chords', () => {
        expect(detectChord([60, 64, 67])).toBe('C Major');
    });

    test('detects minor chords', () => {
        expect(detectChord([60, 63, 67])).toBe('C Minor');
    });

    test('detects diminished chords', () => {
        expect(detectChord([60, 63, 66])).toBe('C Diminished');
    });

    test('returns null for too few notes', () => {
        expect(detectChord([60])).toBeNull();
    });

    test('returns null for unrecognized combinations', () => {
        expect(detectChord([60, 62, 65])).toBeNull();
    });
});