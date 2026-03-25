import pytest
from counterpoint_gen.main import app


@pytest.fixture
def client():
    app.testing = True
    with app.test_client() as client:
        yield client


def test_generate_success(client):
    payload = {
        "cantus": ["C4", "D4", "E4", "F4", "G4", "F4", "E4", "D4", "C4"],
        "realtime": False,
        "midi": False
    }
    response = client.post("/generate", json=payload)
    assert response.status_code == 200
    assert response.json.get("status") == "success"


def test_generate_missing_cantus(client):
    response = client.post("/generate", json={})
    assert response.status_code == 400
    assert "Missing cantus input" in response.json.get("error", "")


def test_generate_invalid_note(client):
    payload = {"cantus": ["C4", "D4", "Z#9"], "realtime": False, "midi": False}
    response = client.post("/generate", json=payload)
    assert response.status_code == 400
    assert "Invalid note" in response.json.get("error", "")


def test_generate_midi_output(client):
    payload = {
        "cantus": ["C4", "D4", "E4", "F4", "G4"],
        "realtime": False,
        "midi": True
    }
    response = client.post("/generate", json=payload)
    assert response.status_code == 200
    assert response.content_type == "audio/midi"


def test_generate_single_note(client):
    payload = {"cantus": ["C4"], "realtime": False, "midi": False}
    response = client.post("/generate", json=payload)
    assert response.status_code == 200
    assert response.json.get("status") == "success"


def test_index(client):
    response = client.get("/")
    assert response.status_code == 200


def test_generate_non_json_body(client):
    response = client.post("/generate", data="not json", content_type="text/plain")
    # Flask's get_json(force=True) will try to parse and fail
    assert response.status_code in (400, 500)
