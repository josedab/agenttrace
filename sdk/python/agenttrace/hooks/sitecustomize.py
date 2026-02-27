"""
sitecustomize hook for zero-config auto-instrumentation.

When this file is on the PYTHONPATH (e.g., via `agenttrace auto-instrument`),
it automatically detects and instruments AI frameworks at Python startup time.

Usage:
    PYTHONPATH=$(python -c "import agenttrace; print(agenttrace.__path__[0] + '/hooks')"):$PYTHONPATH python agent.py

Or via the CLI:
    agenttrace auto-instrument -- python agent.py
"""

import os


def _auto_instrument_on_startup():
    """Called automatically when Python starts if this file is on PYTHONPATH."""
    if os.environ.get("AGENTTRACE_DISABLE_AUTO_INSTRUMENT", "").lower() in ("1", "true"):
        return

    if not os.environ.get("AGENTTRACE_API_KEY") and not os.environ.get("AGENTTRACE_HOST"):
        return

    try:
        from agenttrace.auto_discovery import auto_instrument
        auto_instrument()
    except ImportError:
        pass
    except Exception:
        pass


_auto_instrument_on_startup()
