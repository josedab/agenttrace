"""Request handlers for the AgentTrace Jupyter extension."""

import json
import os
from datetime import datetime
from typing import Any, Dict, List, Optional

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado.web

try:
    from agenttrace import AgentTrace
    AGENTTRACE_AVAILABLE = True
except ImportError:
    AGENTTRACE_AVAILABLE = False


class BaseAgentTraceHandler(APIHandler):
    """Base handler for AgentTrace API endpoints."""

    def get_client(self) -> Optional[Any]:
        """Get the AgentTrace client instance."""
        if not AGENTTRACE_AVAILABLE:
            return None

        api_key = os.environ.get("AGENTTRACE_API_KEY")
        base_url = os.environ.get("AGENTTRACE_BASE_URL", "https://api.agenttrace.io")

        if not api_key:
            return None

        return AgentTrace(api_key=api_key, base_url=base_url)


class ConfigHandler(BaseAgentTraceHandler):
    """Handler for extension configuration."""

    @tornado.web.authenticated
    def get(self):
        """Get the current configuration."""
        config = {
            "available": AGENTTRACE_AVAILABLE,
            "configured": bool(os.environ.get("AGENTTRACE_API_KEY")),
            "baseUrl": os.environ.get("AGENTTRACE_BASE_URL", "https://api.agenttrace.io"),
            "projectId": os.environ.get("AGENTTRACE_PROJECT_ID"),
            "autoTrace": os.environ.get("AGENTTRACE_AUTO_TRACE", "true").lower() == "true",
        }
        self.finish(json.dumps(config))

    @tornado.web.authenticated
    def post(self):
        """Update the configuration."""
        data = self.get_json_body()

        if "apiKey" in data:
            os.environ["AGENTTRACE_API_KEY"] = data["apiKey"]
        if "baseUrl" in data:
            os.environ["AGENTTRACE_BASE_URL"] = data["baseUrl"]
        if "projectId" in data:
            os.environ["AGENTTRACE_PROJECT_ID"] = data["projectId"]
        if "autoTrace" in data:
            os.environ["AGENTTRACE_AUTO_TRACE"] = str(data["autoTrace"]).lower()

        self.finish(json.dumps({"status": "ok"}))


class TracesHandler(BaseAgentTraceHandler):
    """Handler for listing and creating traces."""

    @tornado.web.authenticated
    def get(self):
        """List recent traces."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        limit = int(self.get_argument("limit", "20"))

        # In real implementation, fetch from AgentTrace API
        traces = {
            "traces": [],
            "totalCount": 0,
            "hasMore": False,
        }

        self.finish(json.dumps(traces))

    @tornado.web.authenticated
    def post(self):
        """Create a new trace."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        data = self.get_json_body()

        # In real implementation, create trace via SDK
        trace = {
            "id": "trace-id",
            "name": data.get("name", "notebook-trace"),
            "createdAt": datetime.now().isoformat(),
        }

        self.set_status(201)
        self.finish(json.dumps(trace))


class TraceDetailHandler(BaseAgentTraceHandler):
    """Handler for individual trace details."""

    @tornado.web.authenticated
    def get(self, trace_id: str):
        """Get trace details."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        # In real implementation, fetch from AgentTrace API
        self.set_status(404)
        self.finish(json.dumps({"error": "Trace not found"}))


class CellTraceHandler(BaseAgentTraceHandler):
    """Handler for cell-level tracing."""

    @tornado.web.authenticated
    def post(self):
        """Start tracing a cell execution."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        data = self.get_json_body()
        cell_id = data.get("cellId")
        notebook_path = data.get("notebookPath")
        cell_content = data.get("content")

        # In real implementation, create span for cell
        span = {
            "id": f"span-{cell_id}",
            "cellId": cell_id,
            "startTime": datetime.now().isoformat(),
        }

        self.set_status(201)
        self.finish(json.dumps(span))

    @tornado.web.authenticated
    def patch(self):
        """Update a cell trace with results."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        data = self.get_json_body()
        span_id = data.get("spanId")
        output = data.get("output")
        error = data.get("error")

        # In real implementation, update span
        span = {
            "id": span_id,
            "endTime": datetime.now().isoformat(),
            "output": output,
            "error": error,
        }

        self.finish(json.dumps(span))


class NotebookTraceHandler(BaseAgentTraceHandler):
    """Handler for notebook-level tracing."""

    @tornado.web.authenticated
    def post(self):
        """Create a trace for a notebook session."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        data = self.get_json_body()
        notebook_path = data.get("notebookPath")
        notebook_name = os.path.basename(notebook_path) if notebook_path else "untitled"

        # In real implementation, create trace via SDK
        trace = {
            "id": "trace-id",
            "name": f"notebook:{notebook_name}",
            "notebookPath": notebook_path,
            "createdAt": datetime.now().isoformat(),
        }

        self.set_status(201)
        self.finish(json.dumps(trace))


class VisualizationHandler(BaseAgentTraceHandler):
    """Handler for trace visualization data."""

    @tornado.web.authenticated
    def get(self, trace_id: str):
        """Get visualization data for a trace."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        # In real implementation, fetch and format for visualization
        viz_data = {
            "traceId": trace_id,
            "timeline": [],
            "spans": [],
            "metrics": {
                "totalDuration": 0,
                "totalCost": 0,
                "totalTokens": 0,
            },
        }

        self.finish(json.dumps(viz_data))


class MetricsHandler(BaseAgentTraceHandler):
    """Handler for notebook metrics."""

    @tornado.web.authenticated
    def get(self):
        """Get aggregated metrics for the current session."""
        client = self.get_client()
        if not client:
            self.set_status(503)
            self.finish(json.dumps({"error": "AgentTrace not configured"}))
            return

        metrics = {
            "totalTraces": 0,
            "totalCost": 0.0,
            "totalTokens": 0,
            "averageLatency": 0.0,
            "cellExecutions": 0,
            "llmCalls": 0,
        }

        self.finish(json.dumps(metrics))


def setup_handlers(web_app):
    """Set up the request handlers for the extension."""
    host_pattern = ".*$"
    base_url = web_app.settings["base_url"]

    handlers = [
        (url_path_join(base_url, "agenttrace", "config"), ConfigHandler),
        (url_path_join(base_url, "agenttrace", "traces"), TracesHandler),
        (url_path_join(base_url, "agenttrace", "traces", "(.+)"), TraceDetailHandler),
        (url_path_join(base_url, "agenttrace", "cell-trace"), CellTraceHandler),
        (url_path_join(base_url, "agenttrace", "notebook-trace"), NotebookTraceHandler),
        (url_path_join(base_url, "agenttrace", "visualization", "(.+)"), VisualizationHandler),
        (url_path_join(base_url, "agenttrace", "metrics"), MetricsHandler),
    ]

    web_app.add_handlers(host_pattern, handlers)
