import * as vscode from 'vscode';
import { execSync } from 'child_process';
import { AgentTraceClient, Trace } from '../utils/client';
import { TracesTreeProvider, TraceTreeItem } from '../providers/tracesTreeProvider';
import { SessionsTreeProvider } from '../providers/sessionsTreeProvider';
import { GitHistoryTreeProvider } from '../providers/gitHistoryTreeProvider';

export function registerCommands(
    context: vscode.ExtensionContext,
    client: AgentTraceClient,
    tracesProvider: TracesTreeProvider,
    sessionsProvider: SessionsTreeProvider,
    gitHistoryProvider: GitHistoryTreeProvider
) {
    // Refresh traces
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.refreshTraces', () => {
            tracesProvider.refresh();
            sessionsProvider.refresh();
            gitHistoryProvider.refresh();
        })
    );

    // View trace details
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.viewTrace', async (traceOrItem: Trace | TraceTreeItem) => {
            const trace = traceOrItem instanceof TraceTreeItem ? traceOrItem.trace : traceOrItem;

            if (!trace) {
                vscode.window.showErrorMessage('No trace selected');
                return;
            }

            // Create and show a webview panel with trace details
            const panel = vscode.window.createWebviewPanel(
                'agenttraceTrace',
                `Trace: ${trace.name || trace.id.substring(0, 8)}`,
                vscode.ViewColumn.One,
                {
                    enableScripts: true,
                    retainContextWhenHidden: true,
                }
            );

            // Load observations
            const observations = await client.getTraceObservations(trace.id);

            panel.webview.html = getTraceDetailHtml(trace, observations);
        })
    );

    // Open in browser
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.openInBrowser', (traceOrItem: Trace | TraceTreeItem) => {
            const trace = traceOrItem instanceof TraceTreeItem ? traceOrItem.trace : traceOrItem;

            if (!trace) {
                vscode.window.showErrorMessage('No trace selected');
                return;
            }

            const config = vscode.workspace.getConfiguration('agenttrace');
            const dashboardUrl = config.get<string>('dashboardUrl', 'https://app.agenttrace.io');

            const url = `${dashboardUrl}/projects/${client.getProjectId()}/traces/${trace.id}`;
            vscode.env.openExternal(vscode.Uri.parse(url));
        })
    );

    // Configure AgentTrace
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.configure', async () => {
            const options = ['Set API Key', 'Set Project ID', 'Set API URL', 'Open Settings'];
            const selected = await vscode.window.showQuickPick(options, {
                placeHolder: 'Configure AgentTrace',
            });

            switch (selected) {
                case 'Set API Key':
                    const apiKey = await vscode.window.showInputBox({
                        prompt: 'Enter your AgentTrace API Key',
                        password: true,
                        placeHolder: 'sk-...',
                    });
                    if (apiKey) {
                        await vscode.workspace.getConfiguration('agenttrace').update('apiKey', apiKey, true);
                        vscode.window.showInformationMessage('API Key saved');
                    }
                    break;

                case 'Set Project ID':
                    const projectId = await vscode.window.showInputBox({
                        prompt: 'Enter your AgentTrace Project ID',
                        placeHolder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
                    });
                    if (projectId) {
                        await vscode.workspace.getConfiguration('agenttrace').update('projectId', projectId, true);
                        vscode.window.showInformationMessage('Project ID saved');
                    }
                    break;

                case 'Set API URL':
                    const apiUrl = await vscode.window.showInputBox({
                        prompt: 'Enter your AgentTrace API URL (for self-hosted)',
                        value: vscode.workspace.getConfiguration('agenttrace').get<string>('apiUrl'),
                    });
                    if (apiUrl) {
                        await vscode.workspace.getConfiguration('agenttrace').update('apiUrl', apiUrl, true);
                        vscode.window.showInformationMessage('API URL saved');
                    }
                    break;

                case 'Open Settings':
                    vscode.commands.executeCommand('workbench.action.openSettings', 'agenttrace');
                    break;
            }
        })
    );

    // Link current git commit to trace
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.linkGitCommit', async () => {
            // Get current git info
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                vscode.window.showErrorMessage('No workspace folder open');
                return;
            }

            try {
                const commitSha = execSync('git rev-parse HEAD', {
                    cwd: workspaceFolder.uri.fsPath,
                    encoding: 'utf-8',
                }).trim();

                const branch = execSync('git rev-parse --abbrev-ref HEAD', {
                    cwd: workspaceFolder.uri.fsPath,
                    encoding: 'utf-8',
                }).trim();

                const commitMessage = execSync('git log -1 --format=%s', {
                    cwd: workspaceFolder.uri.fsPath,
                    encoding: 'utf-8',
                }).trim();

                const commitAuthor = execSync('git log -1 --format=%an', {
                    cwd: workspaceFolder.uri.fsPath,
                    encoding: 'utf-8',
                }).trim();

                // Ask for trace ID
                const traceId = await vscode.window.showInputBox({
                    prompt: 'Enter the Trace ID to link',
                    placeHolder: 'trace-id',
                });

                if (!traceId) return;

                const gitLink = await client.createGitLink({
                    traceId,
                    commitSha,
                    branch,
                    commitMessage,
                    commitAuthor,
                });

                if (gitLink) {
                    vscode.window.showInformationMessage(`Linked commit ${commitSha.substring(0, 7)} to trace`);
                    gitHistoryProvider.refresh();
                } else {
                    vscode.window.showErrorMessage('Failed to create git link');
                }
            } catch (error) {
                vscode.window.showErrorMessage('Failed to get git info - is this a git repository?');
            }
        })
    );

    // Create checkpoint
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.createCheckpoint', async () => {
            const traceId = await vscode.window.showInputBox({
                prompt: 'Enter the Trace ID',
                placeHolder: 'trace-id',
            });

            if (!traceId) return;

            const name = await vscode.window.showInputBox({
                prompt: 'Enter checkpoint name',
                placeHolder: 'my-checkpoint',
            });

            if (!name) return;

            const description = await vscode.window.showInputBox({
                prompt: 'Enter checkpoint description (optional)',
                placeHolder: 'Description...',
            });

            const checkpoint = await client.createCheckpoint({
                traceId,
                name,
                description: description || undefined,
                type: 'manual',
            });

            if (checkpoint) {
                vscode.window.showInformationMessage(`Checkpoint "${name}" created`);
            } else {
                vscode.window.showErrorMessage('Failed to create checkpoint');
            }
        })
    );

    // Show cost summary
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.showCostSummary', async () => {
            const summary = await client.getCostSummary();

            if (!summary) {
                vscode.window.showErrorMessage('Failed to load cost summary');
                return;
            }

            const panel = vscode.window.createWebviewPanel(
                'agenttraceCosts',
                'AgentTrace Cost Summary',
                vscode.ViewColumn.One,
                { enableScripts: true }
            );

            panel.webview.html = getCostSummaryHtml(summary);
        })
    );

    // Filter by file
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.filterByFile', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showErrorMessage('No active editor');
                return;
            }

            const filePath = vscode.workspace.asRelativePath(editor.document.uri);
            const traces = await client.getTracesByFile(filePath);

            if (traces.length === 0) {
                vscode.window.showInformationMessage(`No traces found for ${filePath}`);
                return;
            }

            const selected = await vscode.window.showQuickPick(
                traces.map(t => ({
                    label: t.name || t.id.substring(0, 12),
                    description: new Date(t.startTime).toLocaleString(),
                    detail: `${t.status} | $${t.totalCost?.toFixed(4) || '0.00'}`,
                    trace: t,
                })),
                {
                    placeHolder: `Select a trace for ${filePath}`,
                }
            );

            if (selected) {
                vscode.commands.executeCommand('agenttrace.viewTrace', selected.trace);
            }
        })
    );

    // Copy trace ID
    context.subscriptions.push(
        vscode.commands.registerCommand('agenttrace.copyTraceId', (traceOrItem: Trace | TraceTreeItem) => {
            const trace = traceOrItem instanceof TraceTreeItem ? traceOrItem.trace : traceOrItem;

            if (!trace) {
                vscode.window.showErrorMessage('No trace selected');
                return;
            }

            vscode.env.clipboard.writeText(trace.id);
            vscode.window.showInformationMessage(`Copied trace ID: ${trace.id}`);
        })
    );
}

function getTraceDetailHtml(trace: Trace, observations: any[]): string {
    const formatDuration = (ms?: number) => ms ? `${(ms / 1000).toFixed(2)}s` : 'N/A';
    const formatCost = (cost?: number) => cost ? `$${cost.toFixed(4)}` : '$0.00';

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Trace Details</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            padding: 20px;
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
        }
        h1 { margin-bottom: 20px; }
        .section {
            margin-bottom: 30px;
            padding: 15px;
            background: var(--vscode-editor-inactiveSelectionBackground);
            border-radius: 8px;
        }
        .section h2 {
            margin-top: 0;
            border-bottom: 1px solid var(--vscode-widget-border);
            padding-bottom: 10px;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        .stat {
            padding: 10px;
            background: var(--vscode-input-background);
            border-radius: 4px;
        }
        .stat-label {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
        }
        .stat-value {
            font-size: 18px;
            font-weight: bold;
        }
        .observation {
            padding: 10px;
            margin: 10px 0;
            background: var(--vscode-input-background);
            border-left: 3px solid var(--vscode-textLink-foreground);
            border-radius: 4px;
        }
        .observation.generation { border-color: #a78bfa; }
        .observation.span { border-color: #60a5fa; }
        .observation.event { border-color: #f59e0b; }
        .status-completed { color: #10b981; }
        .status-running { color: #3b82f6; }
        .status-error { color: #ef4444; }
        pre {
            background: var(--vscode-textBlockQuote-background);
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
        }
    </style>
</head>
<body>
    <h1>${trace.name || 'Trace'}</h1>

    <div class="section">
        <h2>Overview</h2>
        <div class="grid">
            <div class="stat">
                <div class="stat-label">Status</div>
                <div class="stat-value status-${trace.status}">${trace.status}</div>
            </div>
            <div class="stat">
                <div class="stat-label">Duration</div>
                <div class="stat-value">${formatDuration(trace.duration)}</div>
            </div>
            <div class="stat">
                <div class="stat-label">Total Cost</div>
                <div class="stat-value">${formatCost(trace.totalCost)}</div>
            </div>
            <div class="stat">
                <div class="stat-label">Tokens</div>
                <div class="stat-value">${(trace.inputTokens || 0) + (trace.outputTokens || 0)}</div>
            </div>
        </div>
    </div>

    <div class="section">
        <h2>Details</h2>
        <p><strong>ID:</strong> <code>${trace.id}</code></p>
        <p><strong>Started:</strong> ${new Date(trace.startTime).toLocaleString()}</p>
        ${trace.endTime ? `<p><strong>Ended:</strong> ${new Date(trace.endTime).toLocaleString()}</p>` : ''}
        ${trace.sessionId ? `<p><strong>Session:</strong> <code>${trace.sessionId}</code></p>` : ''}
        ${trace.gitCommitSha ? `<p><strong>Git:</strong> <code>${trace.gitCommitSha.substring(0, 7)}</code> (${trace.gitBranch || 'unknown'})</p>` : ''}
    </div>

    <div class="section">
        <h2>Observations (${observations.length})</h2>
        ${observations.map(obs => `
            <div class="observation ${obs.type}">
                <strong>${obs.name}</strong> <span style="opacity: 0.7">(${obs.type})</span>
                ${obs.model ? `<div style="font-size: 12px;">Model: ${obs.model}</div>` : ''}
                ${obs.cost ? `<div style="font-size: 12px;">Cost: ${formatCost(obs.cost)}</div>` : ''}
            </div>
        `).join('')}
    </div>

    ${trace.metadata ? `
    <div class="section">
        <h2>Metadata</h2>
        <pre>${JSON.stringify(trace.metadata, null, 2)}</pre>
    </div>
    ` : ''}
</body>
</html>`;
}

function getCostSummaryHtml(summary: any): string {
    const byModel = summary.byModel || {};
    const modelRows = Object.entries(byModel)
        .map(([model, cost]) => `<tr><td>${model}</td><td>$${(cost as number).toFixed(4)}</td></tr>`)
        .join('');

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cost Summary</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            padding: 20px;
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
        }
        h1 { margin-bottom: 30px; }
        .cards {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;
            margin-bottom: 30px;
        }
        .card {
            padding: 20px;
            background: var(--vscode-editor-inactiveSelectionBackground);
            border-radius: 8px;
            text-align: center;
        }
        .card-label {
            font-size: 14px;
            color: var(--vscode-descriptionForeground);
            margin-bottom: 10px;
        }
        .card-value {
            font-size: 32px;
            font-weight: bold;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid var(--vscode-widget-border);
        }
        th {
            background: var(--vscode-editor-inactiveSelectionBackground);
        }
    </style>
</head>
<body>
    <h1>Cost Summary</h1>

    <div class="cards">
        <div class="card">
            <div class="card-label">Today</div>
            <div class="card-value">$${summary.today?.toFixed(4) || '0.00'}</div>
        </div>
        <div class="card">
            <div class="card-label">This Week</div>
            <div class="card-value">$${summary.thisWeek?.toFixed(4) || '0.00'}</div>
        </div>
        <div class="card">
            <div class="card-label">This Month</div>
            <div class="card-value">$${summary.thisMonth?.toFixed(4) || '0.00'}</div>
        </div>
    </div>

    <h2>Cost by Model</h2>
    <table>
        <thead>
            <tr>
                <th>Model</th>
                <th>Cost</th>
            </tr>
        </thead>
        <tbody>
            ${modelRows || '<tr><td colspan="2">No data</td></tr>'}
        </tbody>
    </table>
</body>
</html>`;
}
