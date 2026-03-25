from flask import Flask
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s"
)

logger = logging.getLogger(__name__)


def create_app():
    app = Flask(__name__)

    from counterpoint_gen.counterpoint import bp as counterpoint_bp
    from counterpoint_gen.interval_analyzer import bp as analyzer_bp
    from counterpoint_gen.scales import bp as scales_bp
    from counterpoint_gen.harmonizer import bp as harmonizer_bp

    app.register_blueprint(counterpoint_bp)
    app.register_blueprint(analyzer_bp)
    app.register_blueprint(scales_bp)
    app.register_blueprint(harmonizer_bp)

    @app.route("/")
    def index():
        return "MIDI Lab Music Theory API"

    return app


app = create_app()

if __name__ == "__main__":
    logging.getLogger().setLevel(logging.DEBUG)
    app.run(debug=True)
