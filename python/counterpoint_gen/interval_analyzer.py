from flask import Blueprint, request, jsonify
from music21 import note, interval
import logging

logger = logging.getLogger(__name__)

bp = Blueprint("interval_analyzer", __name__)

CONSONANT_INTERVALS = {'m3', 'M3', 'P5', 'm6', 'M6', 'P8'}
STEP_INTERVALS = {'m2', 'M2'}


def analyze_melody(note_names):
    """Analyze a melody's intervals, returning classifications and statistics.

    Returns interval sequence, consonance/dissonance classification,
    and aggregate statistics (largest leap, stepwise motion %, most common interval).
    """
    if len(note_names) < 2:
        return {
            "intervals": [],
            "classifications": [],
            "statistics": {
                "total_intervals": 0,
                "consonant_percent": 0.0,
                "largest_leap": None,
                "most_common_interval": None,
                "stepwise_motion_percent": 0.0,
            }
        }

    notes = [note.Note(n) for n in note_names]
    intervals = []
    classifications = []

    for i in range(len(notes) - 1):
        intvl = interval.Interval(noteStart=notes[i], noteEnd=notes[i + 1])
        interval_name = intvl.simpleName
        semitones = abs(intvl.semitones)
        is_consonant = interval_name in CONSONANT_INTERVALS
        is_step = interval_name in STEP_INTERVALS

        intervals.append({
            "from": note_names[i],
            "to": note_names[i + 1],
            "interval": interval_name,
            "semitones": semitones,
            "direction": intvl.direction.name if hasattr(intvl.direction, 'name') else str(intvl.direction),
        })
        classifications.append({
            "interval": interval_name,
            "consonant": is_consonant,
            "step": is_step,
            "leap": semitones > 2,
        })

    total = len(intervals)
    consonant_count = sum(1 for c in classifications if c["consonant"])
    step_count = sum(1 for c in classifications if c["step"])
    leap_intervals = [i for i, c in zip(intervals, classifications) if c["leap"]]
    largest_leap = max(leap_intervals, key=lambda x: x["semitones"]) if leap_intervals else None

    interval_counts = {}
    for i in intervals:
        name = i["interval"]
        interval_counts[name] = interval_counts.get(name, 0) + 1
    most_common = max(interval_counts, key=interval_counts.get) if interval_counts else None

    return {
        "intervals": intervals,
        "classifications": classifications,
        "statistics": {
            "total_intervals": total,
            "consonant_percent": round(consonant_count / total * 100, 1) if total else 0.0,
            "largest_leap": {
                "interval": largest_leap["interval"],
                "semitones": largest_leap["semitones"],
                "from": largest_leap["from"],
                "to": largest_leap["to"],
            } if largest_leap else None,
            "most_common_interval": most_common,
            "stepwise_motion_percent": round(step_count / total * 100, 1) if total else 0.0,
        }
    }


@bp.route("/analyze", methods=["POST"])
def analyze():
    logger.info("Received /analyze POST request")
    try:
        data = request.get_json(force=True)
    except Exception:
        return jsonify({"error": "Invalid JSON"}), 400

    melody = data.get("melody", [])
    if not melody:
        return jsonify({"error": "Missing melody input"}), 400

    try:
        result = analyze_melody(melody)
    except Exception as e:
        return jsonify({"error": f"Analysis error: {str(e)}"}), 400

    return jsonify(result)
