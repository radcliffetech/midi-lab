from flask import Flask, request, send_file, jsonify
from music21 import note, stream, interval, pitch
import tempfile
import mido
import time
import os
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s"
)
logging.getLogger('werkzeug').setLevel(logging.INFO)
logging.getLogger('flask.app').setLevel(logging.INFO)

# Explicitly add a StreamHandler to root logger to ensure logs go to stdout
console = logging.StreamHandler()
console.setLevel(logging.DEBUG)
formatter = logging.Formatter('%(asctime)s [%(levelname)s] %(message)s')
console.setFormatter(formatter)
logging.getLogger().addHandler(console)

logger = logging.getLogger(__name__)

app = Flask(__name__)

CONSONANT_INTERVALS = ['m3', 'M3', 'P5', 'm6', 'M6', 'P8']

def is_consonant(cantus_note, counter_note):
    intvl = interval.Interval(noteStart=cantus_note, noteEnd=counter_note)
    return intvl.simpleName in CONSONANT_INTERVALS

def generate_counterpoint(cantus_firmus):
    counterpoint = []
    for cf_note in cantus_firmus:
        candidates = []
        for semitone_offset in range(0, 20):
            cp_note = note.Note()
            cp_note.pitch = pitch.Pitch(cf_note.pitch.midi + semitone_offset)
            if is_consonant(cf_note, cp_note):
                candidates.append(cp_note)
        if not candidates:
            raise Exception(f"No valid counterpoint note for {cf_note}")
        counterpoint.append(candidates[0])
    return counterpoint

def play_realtime(cantus_notes, counterpoint_notes, port_name=None, duration=0.5, velocity=64):
    if port_name is None:
        ports = mido.get_output_names()
        if not ports:
            raise RuntimeError("No MIDI output ports found.")
        port_name = ports[0]
    logger.info(f"Sending to MIDI output: {port_name}")
    with mido.open_output(port_name) as outport:
        for cf, cp in zip(cantus_notes, counterpoint_notes):
            outport.send(mido.Message('note_on', note=cf.pitch.midi, velocity=velocity, channel=0))
            outport.send(mido.Message('note_on', note=cp.pitch.midi, velocity=velocity, channel=1))
            time.sleep(duration)
            outport.send(mido.Message('note_off', note=cf.pitch.midi, velocity=velocity, channel=0))
            outport.send(mido.Message('note_off', note=cp.pitch.midi, velocity=velocity, channel=1))

@app.route("/") 
def index():
    logger.info("Index page accessed")
    return "Counterpoint Generation API"

@app.route("/generate", methods=["POST"])
def generate():
    logger.info("Received /generate POST request")
    try:
        data = request.get_json(force=True)
    except Exception as e:
        logger.error("Failed to parse JSON", exc_info=True)
        return jsonify({"error": "Invalid JSON"}), 400
    logger.info(f"Incoming request from {request.remote_addr} using {request.method}")
    cantus_input = data.get("cantus", [])
    realtime = data.get("realtime", False)
    midi_out = data.get("midi", True)

    logger.info(f"Received cantus: {cantus_input}, realtime={realtime}, midi={midi_out}")

    if not cantus_input:
        logger.error("Missing cantus input", exc_info=True)
        return jsonify({"error": "Missing cantus input"}), 400

    cf_stream = stream.Part()
    try:
        for p in cantus_input:
            cf_stream.append(note.Note(p))
    except Exception as e:
        logger.error("Invalid note error", exc_info=True)
        return jsonify({"error": f"Invalid note: {str(e)}"}), 400

    try:
        cp_notes = generate_counterpoint(cf_stream.notes)
    except Exception as e:
        logger.error("Counterpoint generation error", exc_info=True)
        return jsonify({"error": f"Counterpoint error: {str(e)}"}), 500

    if realtime:
        try:
            play_realtime(cf_stream.notes, cp_notes)
        except Exception as e:
            logger.error("MIDI playback error", exc_info=True)
            return jsonify({"error": f"MIDI playback error: {str(e)}"}), 500

    if midi_out:
        cp_stream = stream.Part(cp_notes)
        score = stream.Score([cp_stream, cf_stream])
        tmp_file = tempfile.NamedTemporaryFile(delete=False, suffix=".mid")
        score.write("midi", fp=tmp_file.name)
        tmp_file.close()
        try:
            return send_file(tmp_file.name, mimetype="audio/midi", as_attachment=True, download_name="counterpoint.mid")
        finally:
            os.remove(tmp_file.name)

    return jsonify({"status": "success"})

if __name__ == "__main__":
    logging.getLogger().setLevel(logging.DEBUG)
    app.run(debug=True)