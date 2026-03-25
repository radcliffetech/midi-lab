import { parseMidiMessage } from './midi-message';

describe('parseMidiMessage', () => {
    test('parses note on message', () => {
        const result = parseMidiMessage([0x90, 60, 100]);
        expect(result).toEqual({
            command: 'note_on',
            channel: 0,
            note: 60,
            velocity: 100,
        });
    });

    test('parses note off via 0x80', () => {
        const result = parseMidiMessage([0x80, 60, 0]);
        expect(result).toEqual({
            command: 'note_off',
            channel: 0,
            note: 60,
            velocity: 0,
        });
    });

    test('parses note off via velocity 0', () => {
        const result = parseMidiMessage([0x90, 60, 0]);
        expect(result).toEqual({
            command: 'note_off',
            channel: 0,
            note: 60,
            velocity: 0,
        });
    });

    test('extracts channel correctly', () => {
        const result = parseMidiMessage([0x91, 60, 100]);
        expect(result?.channel).toBe(1);

        const result2 = parseMidiMessage([0x9F, 60, 100]);
        expect(result2?.channel).toBe(15);
    });

    test('returns null for short data', () => {
        expect(parseMidiMessage([0x90])).toBeNull();
        expect(parseMidiMessage([0x90, 60])).toBeNull();
    });

    test('returns null for null/undefined input', () => {
        expect(parseMidiMessage(null)).toBeNull();
        expect(parseMidiMessage(undefined)).toBeNull();
    });

    test('returns null for empty array', () => {
        expect(parseMidiMessage([])).toBeNull();
    });

    test('classifies control change as other', () => {
        const result = parseMidiMessage([0xB0, 1, 64]);
        expect(result).toEqual({
            command: 'other',
            channel: 0,
            note: 1,
            velocity: 64,
        });
    });

    test('classifies program change as other', () => {
        const result = parseMidiMessage([0xC0, 5, 0]);
        expect(result?.command).toBe('other');
    });

    test('handles max velocity note on', () => {
        const result = parseMidiMessage([0x90, 127, 127]);
        expect(result?.command).toBe('note_on');
        expect(result?.note).toBe(127);
        expect(result?.velocity).toBe(127);
    });

    test('handles note off with non-zero velocity', () => {
        const result = parseMidiMessage([0x80, 60, 64]);
        expect(result?.command).toBe('note_off');
        expect(result?.velocity).toBe(64);
    });
});
