"""AgentTrace JupyterLab Extension

Provides trace visualization and auto-instrumentation for JupyterLab notebooks.
"""

from ._version import __version__
from .handlers import setup_handlers


def _jupyter_labextension_paths():
    """Return the labextension paths."""
    return [{
        "src": "labextension",
        "dest": "@agenttrace/jupyter"
    }]


def _jupyter_server_extension_points():
    """Return the server extension points."""
    return [{
        "module": "agenttrace_jupyter"
    }]


def _load_jupyter_server_extension(server_app):
    """Load the Jupyter server extension."""
    setup_handlers(server_app.web_app)
    name = "agenttrace_jupyter"
    server_app.log.info(f"Registered {name} server extension")


# Backwards compatibility for notebook server
load_jupyter_server_extension = _load_jupyter_server_extension


__all__ = [
    "__version__",
    "_jupyter_labextension_paths",
    "_jupyter_server_extension_points",
    "_load_jupyter_server_extension",
]
