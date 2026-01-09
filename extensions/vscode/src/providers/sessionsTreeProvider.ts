import * as vscode from 'vscode';
import { AgentTraceClient, Session } from '../utils/client';

export class SessionTreeItem extends vscode.TreeItem {
    constructor(
        public readonly session?: Session,
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly itemType: 'session' | 'info' | 'loading' | 'error'
    ) {
        super(label, collapsibleState);
        this.contextValue = itemType;
        this.setupItem();
    }

    private setupItem() {
        if (this.session) {
            this.setupSessionItem();
        } else if (this.itemType === 'loading') {
            this.iconPath = new vscode.ThemeIcon('loading~spin');
        } else if (this.itemType === 'error') {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('errorForeground'));
        }
    }

    private setupSessionItem() {
        if (!this.session) return;

        this.iconPath = new vscode.ThemeIcon('folder', new vscode.ThemeColor('charts.blue'));

        // Build description
        const cost = this.session.totalCost > 0 ? `$${this.session.totalCost.toFixed(4)}` : '';
        const traces = `${this.session.traceCount} traces`;
        this.description = [traces, cost].filter(Boolean).join(' | ');

        // Tooltip
        const tooltip = new vscode.MarkdownString();
        tooltip.appendMarkdown(`### ${this.session.name || this.session.id}\n\n`);
        tooltip.appendMarkdown(`**ID:** \`${this.session.id}\`\n\n`);
        tooltip.appendMarkdown(`**Traces:** ${this.session.traceCount}\n\n`);
        if (cost) tooltip.appendMarkdown(`**Total Cost:** ${cost}\n\n`);
        tooltip.appendMarkdown(`**First Trace:** ${new Date(this.session.firstTraceTime).toLocaleString()}\n\n`);
        tooltip.appendMarkdown(`**Last Trace:** ${new Date(this.session.lastTraceTime).toLocaleString()}`);
        this.tooltip = tooltip;
    }
}

export class SessionsTreeProvider implements vscode.TreeDataProvider<SessionTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<SessionTreeItem | undefined | null | void> = new vscode.EventEmitter<SessionTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<SessionTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    private sessions: Session[] = [];
    private loading = false;
    private error: string | null = null;

    constructor(private client: AgentTraceClient) {
        this.loadSessions();
    }

    refresh(): void {
        this.loadSessions();
    }

    private async loadSessions() {
        if (!this.client.isConfigured()) {
            this.error = 'Not configured';
            this._onDidChangeTreeData.fire();
            return;
        }

        this.loading = true;
        this.error = null;
        this._onDidChangeTreeData.fire();

        try {
            const response = await this.client.listSessions({ limit: 50 });
            this.sessions = response.data;
            this.loading = false;
            this._onDidChangeTreeData.fire();
        } catch (error) {
            this.loading = false;
            this.error = 'Failed to load sessions';
            this._onDidChangeTreeData.fire();
        }
    }

    getTreeItem(element: SessionTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: SessionTreeItem): Promise<SessionTreeItem[]> {
        if (element) {
            return [];
        }

        if (this.loading) {
            return [new SessionTreeItem(undefined, 'Loading sessions...', vscode.TreeItemCollapsibleState.None, 'loading')];
        }

        if (this.error) {
            return [new SessionTreeItem(undefined, this.error, vscode.TreeItemCollapsibleState.None, 'error')];
        }

        if (this.sessions.length === 0) {
            return [new SessionTreeItem(undefined, 'No sessions found', vscode.TreeItemCollapsibleState.None, 'info')];
        }

        return this.sessions.map(session =>
            new SessionTreeItem(
                session,
                session.name || session.id.substring(0, 8),
                vscode.TreeItemCollapsibleState.None,
                'session'
            )
        );
    }

    getParent(element: SessionTreeItem): vscode.ProviderResult<SessionTreeItem> {
        return null;
    }
}
