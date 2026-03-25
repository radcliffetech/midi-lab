import pytest
from counterpoint_gen.main import app
import tempfile
import os


@pytest.fixture
def client():
    app.testing = True
    with app.test_client() as client:
        yield client


def test_harmonize_basic_melody(client):
    payload = {"melody": ["C4", "D4", "E4", "F4", "G4"], "key": "C"}
    response = client.post("/harmonize", json=payload)
    assert response.status_code == 200
    assert response.content_type == "audio/midi"


def test_harmonize_returns_midi_file(client):
    payload = {"melody": ["C4", "E4", "G4", "C5"], "key": "C"}
    response = client.post("/harmonize", json=payload)
    assert response.status_code == 200
    assert len(response.data) > 0
    # MIDI files start with "MThd" header
    assert response.data[:4] == b'MThd'


def test_harmonize_missing_melody(client):
    response = client.post("/harmonize", json={"key": "C"})
    assert response.status_code == 400
    assert "Missing melody" in response.json["error"]


def test_harmonize_different_key(client):
    payload = {"melody": ["G4", "A4", "B4", "C5", "D5"], "key": "G"}
    response = client.post("/harmonize", json=payload)
    assert response.status_code == 200
    assert response.content_type == "audio/midi"


def test_harmonize_single_note(client):
    payload = {"melody": ["C4"], "key": "C"}
    response = client.post("/harmonize", json=payload)
    assert response.status_code == 200
    assert response.content_type == "audio/midi"


def test_harmonize_defaults_to_c(client):
    payload = {"melody": ["C4", "E4", "G4"]}
    response = client.post("/harmonize", json=payload)
    assert response.status_code == 200
    assert response.content_type == "audio/midi"
