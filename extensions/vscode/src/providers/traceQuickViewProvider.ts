import * as vscode from 'vscode';

interface TraceInfo {
  id: string;
  name: string;
  status: string;
  duration: number;
  cost: number;
  tokens: number;
  model: string;
  observations: number;
  errors: number;
  startTime: string;
}

export class TraceQuickViewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'agenttrace.traceQuickView';

  private webviewView?: vscode.WebviewView;
  private currentTrace?: TraceInfo;

  constructor(private readonly extensionUri: vscode.Uri) {}

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ): void {
    this.webviewView = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.extensionUri],
    };

    webviewView.webview.html = this.getHtmlContent();

    webviewView.webview.onDidReceiveMessage((message) => {
      if (message.command === 'openInBrowser') {
        const url = `${this.getApiUrl()}/traces/${message.traceId}`;
        vscode.env.openExternal(vscode.Uri.parse(url));
      }
    });
  }

  showTrace(trace: TraceInfo): void {
    this.currentTrace = trace;
    if (this.webviewView) {
      this.webviewView.webview.postMessage({ type: 'showTrace', trace });
    }
  }

  private getApiUrl(): string {
    return vscode.workspace.getConfiguration('agenttrace').get<string>('apiUrl', 'http://localhost:3000');
  }

  private getHtmlContent(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: var(--vscode-font-family); color: var(--vscode-foreground); padding: 8px; font-size: 12px; }
    .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
    .title { font-weight: bold; font-size: 14px; }
    .status { padding: 2px 6px; border-radius: 10px; font-size: 10px; text-transform: uppercase; }
    .status.success { background: var(--vscode-testing-iconPassed); color: white; }
    .status.error { background: var(--vscode-testing-iconFailed); color: white; }
    .status.running { background: var(--vscode-testing-iconQueued); color: white; }
    .metrics { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin: 8px 0; }
    .metric { background: var(--vscode-editor-background); padding: 8px; border-radius: 4px; border: 1px solid var(--vscode-widget-border); }
    .metric-label { font-size: 10px; color: var(--vscode-descriptionForeground); text-transform: uppercase; }
    .metric-value { font-size: 16px; font-weight: bold; margin-top: 2px; }
    .actions { margin-top: 12px; }
    .actions button { width: 100%; background: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; padding: 6px; cursor: pointer; border-radius: 2px; margin-bottom: 4px; }
    .empty { text-align: center; color: var(--vscode-descriptionForeground); padding: 20px; }
  </style>
</head>
<body>
  <div id="content"><div class="empty">Select a trace to view details</div></div>
  <script>
    const vscode = acquireVsCodeApi();
    window.addEventListener('message', event => {
      const { trace } = event.data;
      if (!trace) return;
      document.getElementById('content').innerHTML =
        '<div class="header"><span class="title">' + trace.name + '</span><span class="status ' + trace.status + '">' + trace.status + '</span></div>' +
        '<div class="metrics">' +
          '<div class="metric"><div class="metric-label">Duration</div><div class="metric-value">' + (trace.duration / 1000).toFixed(1) + 's</div></div>' +
          '<div class="metric"><div class="metric-label">Cost</div><div class="metric-value">$' + trace.cost.toFixed(4) + '</div></div>' +
          '<div class="metric"><div class="metric-label">Tokens</div><div class="metric-value">' + trace.tokens.toLocaleString() + '</div></div>' +
          '<div class="metric"><div class="metric-label">Observations</div><div class="metric-value">' + trace.observations + '</div></div>' +
        '</div>' +
        '<div class="actions"><button onclick="vscode.postMessage({command:\\'openInBrowser\\',traceId:\\'' + trace.id + '\\'})">Open in Dashboard</button></div>';
    });
  </script>
</body>
</html>`;
  }
}
