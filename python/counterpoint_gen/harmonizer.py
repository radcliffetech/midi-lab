from flask import Blueprint, request, send_file, jsonify
from music21 import note, stream, pitch, key as m21key, chord, roman, harmony
import tempfile
import os
import logging

logger = logging.getLogger(__name__)

bp = Blueprint("harmonizer", __name__)

# Voice ranges (MIDI numbers)
RANGES = {
    "soprano": (60, 81),  # C4 - A5
    "alto": (55, 74),     # G3 - D5
    "tenor": (48, 67),    # C3 - G4
    "bass": (36, 60),     # C2 - C4
}

# Common chord progressions for harmonization
PROGRESSION_MAP = {
    1: [("I", [0, 4, 7]), ("vi", [9, 0, 4])],
    2: [("ii", [2, 5, 9]), ("IV", [5, 9, 0])],
    3: [("iii", [4, 7, 11]), ("I", [0, 4, 7])],
    4: [("IV", [5, 9, 0]), ("ii", [2, 5, 9])],
    5: [("V", [7, 11, 2]), ("I", [0, 4, 7])],
    6: [("vi", [9, 0, 4]), ("IV", [5, 9, 0])],
    7: [("V", [7, 11, 2]), ("viio", [11, 2, 5])],
    0: [("I", [0, 4, 7])],
}


def _find_nearest_pitch(target_pc, low, high, prefer_midi=None):
    """Find a pitch with the given pitch class within the range, preferring proximity to prefer_midi."""
    candidates = []
    for midi_num in range(low, high + 1):
        if midi_num % 12 == target_pc:
            candidates.append(midi_num)
    if not candidates:
        return low
    if prefer_midi is not None:
        return min(candidates, key=lambda m: abs(m - prefer_midi))
    return candidates[len(candidates) // 2]


def _check_parallels(prev_voices, curr_voices):
    """Check for parallel fifths and octaves between two sets of voices. Returns True if found."""
    for i in range(len(prev_voices)):
        for j in range(i + 1, len(prev_voices)):
            prev_interval = abs(prev_voices[i] - prev_voices[j]) % 12
            curr_interval = abs(curr_voices[i] - curr_voices[j]) % 12
            if prev_interval == curr_interval and prev_interval in (0, 7):
                prev_motion_i = curr_voices[i] - prev_voices[i]
                prev_motion_j = curr_voices[j] - prev_voices[j]
                if prev_motion_i != 0 and prev_motion_j != 0:
                    if (prev_motion_i > 0) == (prev_motion_j > 0):
                        return True
    return False


def harmonize_melody(melody_names, key_name="C"):
    """Generate a 4-part harmonization for a given melody and key."""
    k = m21key.Key(key_name)
    tonic_pc = pitch.Pitch(key_name).pitchClass

    melody_notes = [note.Note(n) for n in melody_names]

    soprano_part = stream.Part()
    soprano_part.id = "Soprano"
    alto_part = stream.Part()
    alto_part.id = "Alto"
    tenor_part = stream.Part()
    tenor_part.id = "Tenor"
    bass_part = stream.Part()
    bass_part.id = "Bass"

    prev_voices = None

    for i, mel_note in enumerate(melody_notes):
        soprano_midi = mel_note.pitch.midi
        mel_pc = mel_note.pitch.pitchClass
        scale_degree = (mel_pc - tonic_pc) % 12

        # Choose chord tones based on scale degree
        progressions = PROGRESSION_MAP.get(scale_degree, [("I", [0, 4, 7])])
        _, chord_pcs_raw = progressions[0]
        chord_pcs = [(pc + tonic_pc) % 12 for pc in chord_pcs_raw]

        # Assign voices: soprano is given, distribute remaining chord tones
        remaining_pcs = [pc for pc in chord_pcs if pc != mel_pc % 12]
        if len(remaining_pcs) < 2:
            remaining_pcs = chord_pcs[:2]

        # Prefer smooth voice leading from previous beat
        prev_alto = prev_voices[1] if prev_voices else None
        prev_tenor = prev_voices[2] if prev_voices else None
        prev_bass = prev_voices[3] if prev_voices else None

        bass_midi = _find_nearest_pitch(chord_pcs[0], *RANGES["bass"], prefer_midi=prev_bass)

        # Alto and tenor get remaining chord tones
        alto_pc = remaining_pcs[0] if remaining_pcs else chord_pcs[1]
        tenor_pc = remaining_pcs[1] if len(remaining_pcs) > 1 else chord_pcs[0]

        alto_midi = _find_nearest_pitch(alto_pc, *RANGES["alto"], prefer_midi=prev_alto)
        tenor_midi = _find_nearest_pitch(tenor_pc, *RANGES["tenor"], prefer_midi=prev_tenor)

        curr_voices = [soprano_midi, alto_midi, tenor_midi, bass_midi]

        # Attempt to fix parallel fifths/octaves by shifting inner voices
        if prev_voices and _check_parallels(prev_voices, curr_voices):
            for alt_pc in chord_pcs:
                alt_alto = _find_nearest_pitch(alt_pc, *RANGES["alto"], prefer_midi=prev_alto)
                test_voices = [soprano_midi, alt_alto, tenor_midi, bass_midi]
                if not _check_parallels(prev_voices, test_voices):
                    alto_midi = alt_alto
                    curr_voices = test_voices
                    break

        prev_voices = curr_voices

        s_note = note.Note(soprano_midi, quarterLength=1.0)
        a_note = note.Note(alto_midi, quarterLength=1.0)
        t_note = note.Note(tenor_midi, quarterLength=1.0)
        b_note = note.Note(bass_midi, quarterLength=1.0)

        soprano_part.append(s_note)
        alto_part.append(a_note)
        tenor_part.append(t_note)
        bass_part.append(b_note)

    score = stream.Score([soprano_part, alto_part, tenor_part, bass_part])
    return score


@bp.route("/harmonize", methods=["POST"])
def harmonize():
    logger.info("Received /harmonize POST request")
    try:
        data = request.get_json(force=True)
    except Exception:
        return jsonify({"error": "Invalid JSON"}), 400

    melody = data.get("melody", [])
    key_name = data.get("key", "C")

    if not melody:
        return jsonify({"error": "Missing melody input"}), 400

    try:
        score = harmonize_melody(melody, key_name)
    except Exception as e:
        logger.error("Harmonization error", exc_info=True)
        return jsonify({"error": f"Harmonization error: {str(e)}"}), 400

    tmp_file = tempfile.NamedTemporaryFile(delete=False, suffix=".mid")
    score.write("midi", fp=tmp_file.name)
    tmp_file.close()
    try:
        return send_file(tmp_file.name, mimetype="audio/midi", as_attachment=True, download_name="harmonized.mid")
    finally:
        os.remove(tmp_file.name)
