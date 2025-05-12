

VENV := .venv
PYTHON := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

setup:
	python3 -m venv $(VENV)
	$(PIP) install -r counterpoint_gen/requirements.txt

run:
	PYTHONUNBUFFERED=1 $(PYTHON) counterpoint_gen/main.py

test:
	PYTHONPATH=. PYTHONUNBUFFERED=1 $(PYTHON) -m pytest counterpoint_gen/tests

