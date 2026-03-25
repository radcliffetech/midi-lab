export type MidiCommand = 'note_on' | 'note_off' | 'other';

export interface ParsedMidiMessage {
  command: MidiCommand;
  channel: number;
  note: number;
  velocity: number;
}

/**
 * Parse raw MIDI message bytes into a typed object.
 * Returns null if the data is too short to be a valid message.
 */
export function parseMidiMessage(data: Uint8Array | number[] | null | undefined): ParsedMidiMessage | null {
  if (!data || data.length < 3) return null;

  const status = data[0];
  const note = data[1];
  const velocity = data[2];
  const cmd = status & 0xf0;
  const channel = status & 0x0f;

  let command: MidiCommand;
  if (cmd === 0x90 && velocity > 0) {
    command = 'note_on';
  } else if (cmd === 0x80 || (cmd === 0x90 && velocity === 0)) {
    command = 'note_off';
  } else {
    command = 'other';
  }

  return { command, channel, note, velocity };
}
