import * as vscode from 'vscode';
import { AgentTraceClient } from '../utils/client';

export class FileDecorationProvider implements vscode.FileDecorationProvider {
    private _onDidChangeFileDecorations: vscode.EventEmitter<vscode.Uri | vscode.Uri[] | undefined> = new vscode.EventEmitter<vscode.Uri | vscode.Uri[] | undefined>();
    readonly onDidChangeFileDecorations: vscode.Event<vscode.Uri | vscode.Uri[] | undefined> = this._onDidChangeFileDecorations.event;

    private fileTraceCount: Map<string, number> = new Map();
    private lastRefresh: number = 0;
    private refreshInterval = 60000; // 1 minute

    constructor(private client: AgentTraceClient) {
        this.refreshData();
    }

    private async refreshData() {
        const now = Date.now();
        if (now - this.lastRefresh < this.refreshInterval) {
            return;
        }

        this.lastRefresh = now;

        if (!this.client.isConfigured()) {
            return;
        }

        try {
            // Get recent traces and count files
            const response = await this.client.listTraces({ limit: 100 });

            this.fileTraceCount.clear();

            for (const trace of response.data) {
                // Parse metadata for file operations if available
                if (trace.metadata?.filesModified) {
                    for (const file of trace.metadata.filesModified) {
                        const count = this.fileTraceCount.get(file) || 0;
                        this.fileTraceCount.set(file, count + 1);
                    }
                }
            }

            // Notify that decorations may have changed
            this._onDidChangeFileDecorations.fire(undefined);
        } catch (error) {
            console.error('Failed to refresh file decorations:', error);
        }
    }

    provideFileDecoration(uri: vscode.Uri, token: vscode.CancellationToken): vscode.ProviderResult<vscode.FileDecoration> {
        const relativePath = vscode.workspace.asRelativePath(uri);
        const count = this.fileTraceCount.get(relativePath);

        if (count && count > 0) {
            return {
                badge: count > 9 ? '9+' : count.toString(),
                tooltip: `${count} trace${count > 1 ? 's' : ''} modified this file`,
                color: new vscode.ThemeColor('agenttrace.traceSuccessBackground'),
            };
        }

        return undefined;
    }

    refresh() {
        this.lastRefresh = 0;
        this.refreshData();
    }
}
