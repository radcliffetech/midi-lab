from flask import Blueprint, request, send_file, jsonify
from music21 import scale, pitch, note, stream, key
import tempfile
import os
import logging

logger = logging.getLogger(__name__)

bp = Blueprint("scales", __name__)

SCALE_TYPES = {
    "major": scale.MajorScale,
    "natural_minor": scale.MinorScale,
    "harmonic_minor": scale.HarmonicMinorScale,
    "melodic_minor": scale.MelodicMinorScale,
    "dorian": scale.DorianScale,
    "phrygian": scale.PhrygianScale,
    "lydian": scale.LydianScale,
    "mixolydian": scale.MixolydianScale,
    "aeolian": scale.MinorScale,
    "locrian": scale.LocrianScale,
}


def get_scale_info(key_name, scale_type):
    """Return notes, MIDI numbers, and intervals for a given key and scale type."""
    scale_class = SCALE_TYPES.get(scale_type)
    if not scale_class:
        raise ValueError(f"Unknown scale type: {scale_type}. Available: {list(SCALE_TYPES.keys())}")

    tonic = pitch.Pitch(key_name)
    if tonic.octave is None:
        tonic.octave = 4
    sc = scale_class(tonic)
    top = tonic.transpose(12)
    pitches = sc.getPitches(tonic, top)

    notes = [p.nameWithOctave for p in pitches]
    midi_numbers = [p.midi for p in pitches]

    intervals = []
    for i in range(len(pitches) - 1):
        semitones = pitches[i + 1].midi - pitches[i].midi
        intervals.append(semitones)

    return {
        "key": key_name,
        "type": scale_type,
        "notes": notes,
        "midi_numbers": midi_numbers,
        "intervals": intervals,
    }


def detect_scales(note_names):
    """Detect which scales contain all of the given notes."""
    pitches = [pitch.Pitch(n) for n in note_names]
    pitch_classes = sorted(set(p.pitchClass for p in pitches))

    matches = []
    for tonic_pc in range(12):
        tonic = pitch.Pitch(tonic_pc)
        tonic.octave = 4
        for scale_name, scale_class in SCALE_TYPES.items():
            sc = scale_class(tonic)
            scale_pcs = sorted(set(p.pitchClass for p in sc.getPitches(tonic, tonic.transpose(12))))
            if all(pc in scale_pcs for pc in pitch_classes):
                matches.append({
                    "key": tonic.name,
                    "type": scale_name,
                })

    return matches


def scale_to_midi(key_name, scale_type):
    """Generate a MIDI file of the given scale and return the temp file path."""
    tonic = pitch.Pitch(key_name)
    sc = SCALE_TYPES[scale_type](tonic)
    pitches = sc.getPitches(tonic, tonic.transpose(12))

    s = stream.Stream()
    for p in pitches:
        n = note.Note(p)
        n.quarterLength = 1.0
        s.append(n)

    tmp_file = tempfile.NamedTemporaryFile(delete=False, suffix=".mid")
    s.write("midi", fp=tmp_file.name)
    tmp_file.close()
    return tmp_file.name


@bp.route("/scales", methods=["POST"])
def get_scale():
    logger.info("Received /scales POST request")
    try:
        data = request.get_json(force=True)
    except Exception:
        return jsonify({"error": "Invalid JSON"}), 400

    key_name = data.get("key")
    scale_type = data.get("type")
    midi_out = data.get("midi", False)

    if not key_name or not scale_type:
        return jsonify({"error": "Missing 'key' and/or 'type' parameter"}), 400

    try:
        result = get_scale_info(key_name, scale_type)
    except Exception as e:
        return jsonify({"error": str(e)}), 400

    if midi_out:
        try:
            midi_path = scale_to_midi(key_name, scale_type)
            try:
                return send_file(midi_path, mimetype="audio/midi", as_attachment=True, download_name=f"{key_name}_{scale_type}.mid")
            finally:
                os.remove(midi_path)
        except Exception as e:
            return jsonify({"error": f"MIDI generation error: {str(e)}"}), 500

    return jsonify(result)


@bp.route("/detect-scale", methods=["POST"])
def detect():
    logger.info("Received /detect-scale POST request")
    try:
        data = request.get_json(force=True)
    except Exception:
        return jsonify({"error": "Invalid JSON"}), 400

    notes = data.get("notes", [])
    if not notes:
        return jsonify({"error": "Missing 'notes' parameter"}), 400

    try:
        matches = detect_scales(notes)
    except Exception as e:
        return jsonify({"error": f"Detection error: {str(e)}"}), 400

    return jsonify({"matches": matches})
