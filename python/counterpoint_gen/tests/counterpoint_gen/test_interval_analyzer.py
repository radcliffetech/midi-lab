import pytest
from counterpoint_gen.main import app


@pytest.fixture
def client():
    app.testing = True
    with app.test_client() as client:
        yield client


def test_analyze_ascending_scale(client):
    payload = {"melody": ["C4", "D4", "E4", "F4", "G4"]}
    response = client.post("/analyze", json=payload)
    assert response.status_code == 200
    data = response.json
    assert data["statistics"]["total_intervals"] == 4
    assert data["statistics"]["stepwise_motion_percent"] == 100.0


def test_analyze_with_leaps(client):
    payload = {"melody": ["C4", "G4", "C5", "E4"]}
    response = client.post("/analyze", json=payload)
    assert response.status_code == 200
    data = response.json
    assert data["statistics"]["largest_leap"] is not None
    assert data["statistics"]["largest_leap"]["semitones"] > 2


def test_analyze_single_note(client):
    payload = {"melody": ["C4"]}
    response = client.post("/analyze", json=payload)
    assert response.status_code == 200
    data = response.json
    assert data["statistics"]["total_intervals"] == 0
    assert data["intervals"] == []


def test_analyze_empty_melody(client):
    payload = {"melody": []}
    response = client.post("/analyze", json=payload)
    assert response.status_code == 400


def test_analyze_missing_melody(client):
    response = client.post("/analyze", json={})
    assert response.status_code == 400


def test_analyze_consonance_stats(client):
    payload = {"melody": ["C4", "E4", "G4"]}
    response = client.post("/analyze", json=payload)
    assert response.status_code == 200
    data = response.json
    assert "consonant_percent" in data["statistics"]
    assert isinstance(data["statistics"]["consonant_percent"], float)
