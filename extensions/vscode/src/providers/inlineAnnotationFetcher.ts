import * as vscode from 'vscode';
import { AgentTraceClient, FileTraceMapping, LineAnnotation } from '../utils/client';
import { InlineTraceProvider } from './inlineTraceProvider';

/**
 * InlineAnnotationFetcher connects the IDE Trace View API to the
 * InlineTraceProvider, fetching file-level annotations from the
 * AgentTrace backend and rendering them as inline decorations.
 */
export class InlineAnnotationFetcher implements vscode.Disposable {
  private client: AgentTraceClient | null = null;
  private inlineProvider: InlineTraceProvider;
  private disposables: vscode.Disposable[] = [];
  private cache: Map<string, { data: FileTraceMapping; fetchedAt: number }> = new Map();
  private cacheTTLMs = 30_000; // 30 seconds

  constructor(inlineProvider: InlineTraceProvider) {
    this.inlineProvider = inlineProvider;

    // Fetch annotations when the active editor changes
    this.disposables.push(
      vscode.window.onDidChangeActiveTextEditor((editor) => {
        if (editor) {
          this.fetchAndApply(editor.document.uri.fsPath);
        }
      })
    );

    // Fetch annotations when a file is saved
    this.disposables.push(
      vscode.workspace.onDidSaveTextDocument((doc) => {
        // Invalidate cache on save
        this.cache.delete(doc.uri.fsPath);
        if (vscode.window.activeTextEditor?.document.uri.fsPath === doc.uri.fsPath) {
          this.fetchAndApply(doc.uri.fsPath);
        }
      })
    );
  }

  setClient(client: AgentTraceClient): void {
    this.client = client;
  }

  /**
   * Fetch annotations for a file from the API and apply them to the editor
   */
  async fetchAndApply(filePath: string): Promise<void> {
    if (!this.client || !this.client.isConfigured()) {
      return;
    }

    // Check cache
    const cached = this.cache.get(filePath);
    if (cached && Date.now() - cached.fetchedAt < this.cacheTTLMs) {
      this.applyMapping(cached.data);
      return;
    }

    try {
      // Get workspace-relative path
      const workspaceFolders = vscode.workspace.workspaceFolders;
      if (!workspaceFolders) return;

      let relativePath = filePath;
      for (const folder of workspaceFolders) {
        if (filePath.startsWith(folder.uri.fsPath)) {
          relativePath = filePath.substring(folder.uri.fsPath.length + 1);
          break;
        }
      }

      const mapping = await this.client.getFileMapping(relativePath);
      if (mapping && mapping.annotations.length > 0) {
        this.cache.set(filePath, { data: mapping, fetchedAt: Date.now() });
        this.applyMapping(mapping);
      }
    } catch (error) {
      console.error('AgentTrace: Failed to fetch file annotations', error);
    }
  }

  /**
   * Fetch annotations for all currently open files
   */
  async fetchAllOpen(): Promise<void> {
    if (!this.client || !this.client.isConfigured()) return;

    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (!workspaceFolders) return;

    const openFiles = vscode.workspace.textDocuments
      .filter((doc) => doc.uri.scheme === 'file')
      .map((doc) => {
        let relativePath = doc.uri.fsPath;
        for (const folder of workspaceFolders) {
          if (relativePath.startsWith(folder.uri.fsPath)) {
            relativePath = relativePath.substring(folder.uri.fsPath.length + 1);
            break;
          }
        }
        return relativePath;
      });

    if (openFiles.length === 0) return;

    try {
      const mappings = await this.client.getBatchFileMappings(openFiles);
      for (const mapping of mappings) {
        this.applyMapping(mapping);
      }
    } catch (error) {
      console.error('AgentTrace: Failed to fetch batch annotations', error);
    }
  }

  /**
   * Show detailed trace context for a specific trace ID
   */
  async showTraceContext(traceId: string): Promise<void> {
    if (!this.client) return;

    try {
      const context = await this.client.getIDETraceContext(traceId);
      if (!context) {
        vscode.window.showWarningMessage('Trace context not found');
        return;
      }

      // Show in a webview panel
      const panel = vscode.window.createWebviewPanel(
        'agenttraceContext',
        `Trace: ${context.traceName}`,
        vscode.ViewColumn.Beside,
        { enableScripts: false }
      );

      panel.webview.html = this.renderTraceContextHTML(context);
    } catch (error) {
      console.error('AgentTrace: Failed to fetch trace context', error);
    }
  }

  private applyMapping(mapping: FileTraceMapping): void {
    const annotations = mapping.annotations.map((ann: LineAnnotation) => ({
      filePath: mapping.filePath,
      lineNumber: ann.line,
      type: this.mapAnnotationType(ann.type),
      title: `${ann.agentName}: ${ann.type}`,
      description: `Trace: ${ann.traceName} | Confidence: ${(ann.confidence * 100).toFixed(0)}%`,
      cost: ann.cost,
      tokens: undefined,
      model: undefined,
      traceId: ann.traceId,
      observationId: '',
      timestamp: ann.timestamp,
    }));

    // Apply to the first matching editor
    for (const editor of vscode.window.visibleTextEditors) {
      if (editor.document.uri.fsPath.endsWith(mapping.filePath)) {
        this.inlineProvider.setAnnotations(editor.document.uri.fsPath, annotations);
        break;
      }
    }
  }

  private mapAnnotationType(type: string): 'llm_call' | 'tool_call' | 'file_edit' | 'error' {
    switch (type) {
      case 'created':
      case 'modified':
        return 'file_edit';
      case 'read':
        return 'tool_call';
      default:
        return 'llm_call';
    }
  }

  private renderTraceContextHTML(context: import('../utils/client').IDETraceContext): string {
    const fileChangesHTML = context.fileChanges
      .map(
        (fc) => `
        <div class="file-change">
          <strong>${fc.operation}</strong>: <code>${fc.path}</code>
          <p>${fc.diffSummary}</p>
        </div>`
      )
      .join('');

    return `<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: var(--vscode-font-family); padding: 16px; color: var(--vscode-foreground); }
    h2 { color: var(--vscode-textLink-foreground); }
    .metric { display: inline-block; margin-right: 24px; }
    .metric-label { font-size: 0.85em; opacity: 0.7; }
    .metric-value { font-size: 1.2em; font-weight: bold; }
    .reasoning { background: var(--vscode-textBlockQuote-background); padding: 12px; border-radius: 4px; margin: 12px 0; }
    .file-change { border-left: 3px solid var(--vscode-textLink-foreground); padding: 8px 12px; margin: 8px 0; }
    code { background: var(--vscode-textCodeBlock-background); padding: 2px 6px; border-radius: 3px; }
  </style>
</head>
<body>
  <h2>🤖 ${context.traceName}</h2>
  <p><strong>Session:</strong> ${context.agentSession}</p>
  <div>
    <div class="metric">
      <div class="metric-label">Cost</div>
      <div class="metric-value">$${context.cost.toFixed(4)}</div>
    </div>
    <div class="metric">
      <div class="metric-label">Latency</div>
      <div class="metric-value">${context.latencyMs.toFixed(0)}ms</div>
    </div>
  </div>
  <h3>Reasoning</h3>
  <div class="reasoning">${context.reasoning || 'No reasoning captured'}</div>
  <h3>File Changes (${context.fileChanges.length})</h3>
  ${fileChangesHTML || '<p>No file changes recorded</p>'}
</body>
</html>`;
  }

  dispose(): void {
    this.disposables.forEach((d) => d.dispose());
    this.cache.clear();
  }
}
