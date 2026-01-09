import * as vscode from 'vscode';
import { AgentTraceClient } from './utils/client';
import { TracesTreeProvider } from './providers/tracesTreeProvider';
import { SessionsTreeProvider } from './providers/sessionsTreeProvider';
import { GitHistoryTreeProvider } from './providers/gitHistoryTreeProvider';
import { StatusBarManager } from './views/statusBar';
import { FileDecorationProvider } from './providers/fileDecorationProvider';
import { registerCommands } from './commands';

let client: AgentTraceClient;
let statusBarManager: StatusBarManager;
let refreshInterval: NodeJS.Timeout | undefined;

export async function activate(context: vscode.ExtensionContext) {
    console.log('AgentTrace extension is activating...');

    // Initialize configuration
    const config = vscode.workspace.getConfiguration('agenttrace');
    const apiUrl = config.get<string>('apiUrl', 'https://api.agenttrace.io');
    const apiKey = config.get<string>('apiKey', '');
    const projectId = config.get<string>('projectId', '');

    // Initialize client
    client = new AgentTraceClient(apiUrl, apiKey, projectId);

    // Set context for conditional UI
    await vscode.commands.executeCommand('setContext', 'agenttrace.isConfigured', !!apiKey && !!projectId);

    // Initialize tree providers
    const tracesProvider = new TracesTreeProvider(client);
    const sessionsProvider = new SessionsTreeProvider(client);
    const gitHistoryProvider = new GitHistoryTreeProvider(client);

    // Register tree views
    const tracesView = vscode.window.createTreeView('agenttrace.traces', {
        treeDataProvider: tracesProvider,
        showCollapseAll: true,
    });

    const sessionsView = vscode.window.createTreeView('agenttrace.sessions', {
        treeDataProvider: sessionsProvider,
        showCollapseAll: true,
    });

    const gitHistoryView = vscode.window.createTreeView('agenttrace.gitHistory', {
        treeDataProvider: gitHistoryProvider,
        showCollapseAll: true,
    });

    // Initialize status bar
    statusBarManager = new StatusBarManager(client);

    if (config.get<boolean>('showStatusBarItem', true)) {
        statusBarManager.show();
    }

    // Initialize file decorations
    const fileDecorationProvider = new FileDecorationProvider(client);

    if (config.get<boolean>('enableFileDecorations', true)) {
        context.subscriptions.push(
            vscode.window.registerFileDecorationProvider(fileDecorationProvider)
        );
    }

    // Register commands
    registerCommands(context, client, tracesProvider, sessionsProvider, gitHistoryProvider);

    // Set up auto-refresh
    if (config.get<boolean>('autoRefresh', true)) {
        const interval = config.get<number>('refreshInterval', 30) * 1000;
        refreshInterval = setInterval(() => {
            tracesProvider.refresh();
            statusBarManager.refresh();
        }, interval);
    }

    // Listen for configuration changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('agenttrace')) {
                handleConfigurationChange();
            }
        })
    );

    // Add disposables
    context.subscriptions.push(tracesView, sessionsView, gitHistoryView);
    context.subscriptions.push(statusBarManager);

    console.log('AgentTrace extension activated');
}

function handleConfigurationChange() {
    const config = vscode.workspace.getConfiguration('agenttrace');

    // Update client configuration
    const apiUrl = config.get<string>('apiUrl', 'https://api.agenttrace.io');
    const apiKey = config.get<string>('apiKey', '');
    const projectId = config.get<string>('projectId', '');

    client.updateConfig(apiUrl, apiKey, projectId);

    // Update context
    vscode.commands.executeCommand('setContext', 'agenttrace.isConfigured', !!apiKey && !!projectId);

    // Update status bar visibility
    if (config.get<boolean>('showStatusBarItem', true)) {
        statusBarManager.show();
    } else {
        statusBarManager.hide();
    }

    // Update refresh interval
    if (refreshInterval) {
        clearInterval(refreshInterval);
    }

    if (config.get<boolean>('autoRefresh', true)) {
        const interval = config.get<number>('refreshInterval', 30) * 1000;
        refreshInterval = setInterval(() => {
            vscode.commands.executeCommand('agenttrace.refreshTraces');
            statusBarManager.refresh();
        }, interval);
    }
}

export function deactivate() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
    }
}
