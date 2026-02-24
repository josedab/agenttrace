import * as vscode from 'vscode';

interface TimeTravelEntry {
  timestamp: string;
  eventType: string;
  title: string;
  reasoning: string;
  input?: string;
  output?: string;
  model?: string;
  cost?: number;
  tokens?: number;
  filePath?: string;
  lineNumber?: number;
}

export class TimeTravelProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'agenttrace.timeTravel';

  private webviewView?: vscode.WebviewView;
  private entries: TimeTravelEntry[] = [];
  private currentIndex: number = -1;

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
      switch (message.command) {
        case 'stepForward':
          this.stepForward();
          break;
        case 'stepBackward':
          this.stepBackward();
          break;
        case 'jumpTo':
          this.jumpTo(message.index);
          break;
        case 'openFile':
          this.openFileAtLocation(message.filePath, message.lineNumber);
          break;
      }
    });
  }

  setEntries(entries: TimeTravelEntry[]): void {
    this.entries = entries;
    this.currentIndex = entries.length > 0 ? 0 : -1;
    this.updateWebview();
  }

  stepForward(): void {
    if (this.currentIndex < this.entries.length - 1) {
      this.currentIndex++;
      this.updateWebview();
      this.navigateToEntry(this.entries[this.currentIndex]);
    }
  }

  stepBackward(): void {
    if (this.currentIndex > 0) {
      this.currentIndex--;
      this.updateWebview();
      this.navigateToEntry(this.entries[this.currentIndex]);
    }
  }

  jumpTo(index: number): void {
    if (index >= 0 && index < this.entries.length) {
      this.currentIndex = index;
      this.updateWebview();
      this.navigateToEntry(this.entries[index]);
    }
  }

  private async navigateToEntry(entry: TimeTravelEntry): Promise<void> {
    if (entry.filePath && entry.lineNumber) {
      await this.openFileAtLocation(entry.filePath, entry.lineNumber);
    }
  }

  private async openFileAtLocation(filePath: string, lineNumber: number): Promise<void> {
    try {
      const uri = vscode.Uri.file(filePath);
      const document = await vscode.workspace.openTextDocument(uri);
      const editor = await vscode.window.showTextDocument(document);
      const line = Math.max(0, lineNumber - 1);
      const range = new vscode.Range(line, 0, line, 0);
      editor.revealRange(range, vscode.TextEditorRevealType.InCenter);
      editor.selection = new vscode.Selection(range.start, range.start);
    } catch {
      // File may not exist locally
    }
  }

  private updateWebview(): void {
    if (!this.webviewView) return;

    this.webviewView.webview.postMessage({
      type: 'update',
      entries: this.entries,
      currentIndex: this.currentIndex,
    });
  }

  private getHtmlContent(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: var(--vscode-font-family); color: var(--vscode-foreground); padding: 8px; font-size: 12px; }
    .controls { display: flex; gap: 4px; margin-bottom: 8px; align-items: center; }
    .controls button { background: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; padding: 4px 8px; cursor: pointer; border-radius: 2px; }
    .controls button:hover { background: var(--vscode-button-hoverBackground); }
    .controls button:disabled { opacity: 0.5; cursor: default; }
    .counter { flex: 1; text-align: center; color: var(--vscode-descriptionForeground); }
    .entry { padding: 6px; margin: 4px 0; border-left: 3px solid var(--vscode-textLink-foreground); background: var(--vscode-editor-background); cursor: pointer; border-radius: 2px; }
    .entry.current { border-left-color: var(--vscode-focusBorder); background: var(--vscode-list-activeSelectionBackground); }
    .entry.error { border-left-color: var(--vscode-errorForeground); }
    .entry-type { font-size: 10px; text-transform: uppercase; color: var(--vscode-descriptionForeground); }
    .entry-title { font-weight: bold; margin: 2px 0; }
    .entry-reasoning { font-style: italic; color: var(--vscode-descriptionForeground); margin-top: 2px; }
    .entry-meta { display: flex; gap: 8px; font-size: 10px; color: var(--vscode-descriptionForeground); margin-top: 4px; }
    .empty { text-align: center; color: var(--vscode-descriptionForeground); padding: 20px; }
  </style>
</head>
<body>
  <div class="controls">
    <button id="prev" title="Step backward">◀</button>
    <span class="counter" id="counter">No entries</span>
    <button id="next" title="Step forward">▶</button>
  </div>
  <div id="entries"><div class="empty">Load a trace to begin time travel debugging</div></div>
  <script>
    const vscode = acquireVsCodeApi();
    document.getElementById('prev').addEventListener('click', () => vscode.postMessage({ command: 'stepBackward' }));
    document.getElementById('next').addEventListener('click', () => vscode.postMessage({ command: 'stepForward' }));

    window.addEventListener('message', event => {
      const { entries, currentIndex } = event.data;
      const container = document.getElementById('entries');
      const counter = document.getElementById('counter');
      
      if (!entries || entries.length === 0) {
        container.innerHTML = '<div class="empty">No entries</div>';
        counter.textContent = 'No entries';
        return;
      }

      counter.textContent = (currentIndex + 1) + ' / ' + entries.length;
      document.getElementById('prev').disabled = currentIndex <= 0;
      document.getElementById('next').disabled = currentIndex >= entries.length - 1;

      container.innerHTML = entries.map((e, i) => {
        const cls = ['entry'];
        if (i === currentIndex) cls.push('current');
        if (e.eventType === 'error') cls.push('error');
        return '<div class="' + cls.join(' ') + '" onclick="vscode.postMessage({command:\\'jumpTo\\',index:' + i + '})">' +
          '<div class="entry-type">' + e.eventType + '</div>' +
          '<div class="entry-title">' + e.title + '</div>' +
          (e.reasoning ? '<div class="entry-reasoning">' + e.reasoning + '</div>' : '') +
          '<div class="entry-meta">' +
            (e.model ? '<span>🤖 ' + e.model + '</span>' : '') +
            (e.tokens ? '<span>📊 ' + e.tokens + ' tokens</span>' : '') +
            (e.cost ? '<span>💰 $' + e.cost.toFixed(4) + '</span>' : '') +
          '</div></div>';
      }).join('');
    });
  </script>
</body>
</html>`;
  }
}
