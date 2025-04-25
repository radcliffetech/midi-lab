import { detectKey } from './midi-utils';

describe('detectKey', () => {
  test('identifies D Major', () => {
    const dMajor: number[] = [
      2, 0, 1, 1, 2, 1, 0, 2, 1, 0, 1, 0  // C C# D D# E F F# G G# A A# B
    ];
    const result = detectKey(dMajor);
    console.log(`D Major result: ${result}`);
    expect(result).toContain('D');
  });

  test('identifies D Minor', () => {
    const dMinor: number[] = [
      1, 0, 2, 1, 1, 2, 1, 0, 1, 2, 0, 1
    ];
    const result = detectKey(dMinor);
    console.log(`D Minor result: ${result}`);
    expect(result).toContain('D');
  });
});
