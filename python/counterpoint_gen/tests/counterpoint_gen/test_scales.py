import pytest
from counterpoint_gen.main import app


@pytest.fixture
def client():
    app.testing = True
    with app.test_client() as client:
        yield client


def test_major_scale(client):
    response = client.post("/scales", json={"key": "C", "type": "major"})
    assert response.status_code == 200
    data = response.json
    assert data["key"] == "C"
    assert data["type"] == "major"
    assert len(data["notes"]) == 8  # includes octave
    assert data["notes"][0].startswith("C")


def test_dorian_mode(client):
    response = client.post("/scales", json={"key": "D", "type": "dorian"})
    assert response.status_code == 200
    data = response.json
    assert data["type"] == "dorian"
    assert data["notes"][0].startswith("D")


def test_harmonic_minor(client):
    response = client.post("/scales", json={"key": "A", "type": "harmonic_minor"})
    assert response.status_code == 200
    data = response.json
    assert len(data["intervals"]) == 7


def test_unknown_scale_type(client):
    response = client.post("/scales", json={"key": "C", "type": "superlocrian"})
    assert response.status_code == 400
    assert "Unknown scale type" in response.json["error"]


def test_missing_params(client):
    response = client.post("/scales", json={"key": "C"})
    assert response.status_code == 400


def test_detect_c_major_scale(client):
    response = client.post("/detect-scale", json={"notes": ["C4", "D4", "E4", "F4", "G4", "A4", "B4"]})
    assert response.status_code == 200
    matches = response.json["matches"]
    scale_names = [(m["key"], m["type"]) for m in matches]
    assert ("C", "major") in scale_names


def test_detect_scale_empty_notes(client):
    response = client.post("/detect-scale", json={"notes": []})
    assert response.status_code == 400


def test_scale_midi_output(client):
    response = client.post("/scales", json={"key": "C", "type": "major", "midi": True})
    assert response.status_code == 200
    assert response.content_type == "audio/midi"


def test_all_modes_return_results(client):
    modes = ["major", "natural_minor", "harmonic_minor", "melodic_minor",
             "dorian", "phrygian", "lydian", "mixolydian", "aeolian", "locrian"]
    for mode in modes:
        response = client.post("/scales", json={"key": "C", "type": mode})
        assert response.status_code == 200, f"Failed for mode: {mode}"
        assert len(response.json["notes"]) >= 7
