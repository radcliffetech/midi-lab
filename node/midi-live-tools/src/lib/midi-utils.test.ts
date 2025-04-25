import { detectKey } from './midi-utils';

describe('detectKey', () => {
    test('clearly identifies C Major - only chord tones', () => {
        const cMajorStrong: number[] = [
            // C  C# D  D# E  F  F# G  G# A  A# B
             10, 0, 0, 0, 8, 0, 0, 9, 0, 0, 0, 0
        ];
        const result = detectKey(cMajorStrong);
        console.log(`Strong C Major result: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('C Major');
    });
    test('clearly identifies D Minor - only chord tones', () => {
        const dMinorStrong: number[] = [
            // C C# D D# E F F# G G# A A# B
             0, 0, 9, 0, 0, 8, 0, 7, 0, 0, 0, 0
        ];
        const result = detectKey(dMinorStrong);
        console.log(`Strong D Minor result: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('D Minor');
    });

    test('clearly identifies G Major - only chord tones', () => {
        const gMajorStrong: number[] = [
            // C C# D D# E F F# G G# A A# B
             0, 0, 8, 0, 0, 0, 7, 9, 0, 0, 0, 0
        ];
        const result = detectKey(gMajorStrong);
        console.log(`Strong G Major result: ${result.key}, confidence: ${result.confidence}`);
        expect(result.key).toBe('G Major');
    });
});



//   test('identifies D Major', () => {
//     const dMajor: number[] = [
//       // C C# D D# E F F# G G# A A# B
//       2, 0, 1, 1, 2, 1, 0, 2, 1, 0, 1, 0  
//     ];
//     const result = detectKey(dMajor);
//     console.log(`D Major result: ${result}`);
//     expect(result).toContain('D');
//   });

//   test('identifies D Minor', () => {
//     const dMinor: number[] = [
//       1, 0, 2, 1, 1, 2, 1, 0, 1, 2, 0, 1
//     ];
//     const result = detectKey(dMinor);
//     console.log(`D Minor result: ${result}`);
//     expect(result).toContain('D');
//   });

//   test('identifies C Major', () => {
//     const cMajor: number[] = [
//         0, 0, 1, 1, 2, 2, 0, 2, 1, 0, 1, 0
//     ];
//     const result = detectKey(cMajor);
//     console.log(`C Major result: ${result}`);
//     expect(result).toContain('C');
//   });
// });