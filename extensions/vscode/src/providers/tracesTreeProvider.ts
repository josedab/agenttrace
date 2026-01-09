import * as vscode from 'vscode';
import { AgentTraceClient, Trace, Observation } from '../utils/client';

export class TraceTreeItem extends vscode.TreeItem {
    constructor(
        public readonly trace?: Trace,
        public readonly observation?: Observation,
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly itemType: 'trace' | 'observation' | 'info' | 'loading' | 'error'
    ) {
        super(label, collapsibleState);
        this.contextValue = itemType;
        this.setupItem();
    }

    private setupItem() {
        if (this.trace) {
            this.setupTraceItem();
        } else if (this.observation) {
            this.setupObservationItem();
        } else if (this.itemType === 'loading') {
            this.iconPath = new vscode.ThemeIcon('loading~spin');
        } else if (this.itemType === 'error') {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('errorForeground'));
        }
    }

    private setupTraceItem() {
        if (!this.trace) return;

        // Set icon based on status
        switch (this.trace.status) {
            case 'completed':
                this.iconPath = new vscode.ThemeIcon('check', new vscode.ThemeColor('charts.green'));
                break;
            case 'running':
                this.iconPath = new vscode.ThemeIcon('sync~spin', new vscode.ThemeColor('charts.blue'));
                break;
            case 'error':
                this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('charts.red'));
                break;
        }

        // Format duration
        const duration = this.trace.duration ? `${(this.trace.duration / 1000).toFixed(2)}s` : 'running';

        // Format cost
        const cost = this.trace.totalCost > 0 ? `$${this.trace.totalCost.toFixed(4)}` : '';

        // Build description
        const parts = [duration];
        if (cost) parts.push(cost);
        if (this.trace.inputTokens || this.trace.outputTokens) {
            parts.push(`${this.trace.inputTokens + this.trace.outputTokens} tokens`);
        }
        this.description = parts.join(' | ');

        // Tooltip with more details
        const tooltip = new vscode.MarkdownString();
        tooltip.appendMarkdown(`### ${this.trace.name}\n\n`);
        tooltip.appendMarkdown(`**ID:** \`${this.trace.id}\`\n\n`);
        tooltip.appendMarkdown(`**Status:** ${this.trace.status}\n\n`);
        tooltip.appendMarkdown(`**Duration:** ${duration}\n\n`);
        if (cost) tooltip.appendMarkdown(`**Cost:** ${cost}\n\n`);
        tooltip.appendMarkdown(`**Tokens:** ${this.trace.inputTokens} in / ${this.trace.outputTokens} out\n\n`);
        if (this.trace.gitCommitSha) {
            tooltip.appendMarkdown(`**Git:** \`${this.trace.gitCommitSha.substring(0, 7)}\` (${this.trace.gitBranch || 'unknown'})\n\n`);
        }
        tooltip.appendMarkdown(`**Started:** ${new Date(this.trace.startTime).toLocaleString()}`);
        this.tooltip = tooltip;

        // Command to view trace details
        this.command = {
            command: 'agenttrace.viewTrace',
            title: 'View Trace',
            arguments: [this.trace],
        };
    }

    private setupObservationItem() {
        if (!this.observation) return;

        // Set icon based on type
        switch (this.observation.type) {
            case 'generation':
                this.iconPath = new vscode.ThemeIcon('sparkle', new vscode.ThemeColor('charts.purple'));
                break;
            case 'span':
                this.iconPath = new vscode.ThemeIcon('symbol-function', new vscode.ThemeColor('charts.blue'));
                break;
            case 'event':
                this.iconPath = new vscode.ThemeIcon('symbol-event', new vscode.ThemeColor('charts.orange'));
                break;
        }

        // Build description
        const parts: string[] = [];
        if (this.observation.model) parts.push(this.observation.model);
        if (this.observation.cost) parts.push(`$${this.observation.cost.toFixed(4)}`);
        this.description = parts.join(' | ');

        // Tooltip
        const tooltip = new vscode.MarkdownString();
        tooltip.appendMarkdown(`### ${this.observation.name}\n\n`);
        tooltip.appendMarkdown(`**Type:** ${this.observation.type}\n\n`);
        if (this.observation.model) tooltip.appendMarkdown(`**Model:** ${this.observation.model}\n\n`);
        if (this.observation.inputTokens) tooltip.appendMarkdown(`**Input Tokens:** ${this.observation.inputTokens}\n\n`);
        if (this.observation.outputTokens) tooltip.appendMarkdown(`**Output Tokens:** ${this.observation.outputTokens}\n\n`);
        this.tooltip = tooltip;
    }
}

export class TracesTreeProvider implements vscode.TreeDataProvider<TraceTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<TraceTreeItem | undefined | null | void> = new vscode.EventEmitter<TraceTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<TraceTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    private traces: Trace[] = [];
    private loading = false;
    private error: string | null = null;
    private filter: {
        status?: string;
        search?: string;
        sessionId?: string;
        fromTime?: string;
    } = {};

    constructor(private client: AgentTraceClient) {
        this.loadTraces();
    }

    refresh(): void {
        this.loadTraces();
    }

    setFilter(filter: typeof this.filter) {
        this.filter = filter;
        this.loadTraces();
    }

    private async loadTraces() {
        if (!this.client.isConfigured()) {
            this.error = 'Not configured - set API key and project ID';
            this._onDidChangeTreeData.fire();
            return;
        }

        this.loading = true;
        this.error = null;
        this._onDidChangeTreeData.fire();

        try {
            const config = vscode.workspace.getConfiguration('agenttrace');
            const limit = config.get<number>('maxTraces', 50);

            const response = await this.client.listTraces({
                limit,
                ...this.filter,
            });

            this.traces = response.data;
            this.loading = false;
            this._onDidChangeTreeData.fire();
        } catch (error) {
            this.loading = false;
            this.error = 'Failed to load traces';
            this._onDidChangeTreeData.fire();
        }
    }

    getTreeItem(element: TraceTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: TraceTreeItem): Promise<TraceTreeItem[]> {
        if (!element) {
            // Root level - show traces or status messages
            if (this.loading) {
                return [new TraceTreeItem(undefined, undefined, 'Loading traces...', vscode.TreeItemCollapsibleState.None, 'loading')];
            }

            if (this.error) {
                return [new TraceTreeItem(undefined, undefined, this.error, vscode.TreeItemCollapsibleState.None, 'error')];
            }

            if (this.traces.length === 0) {
                return [new TraceTreeItem(undefined, undefined, 'No traces found', vscode.TreeItemCollapsibleState.None, 'info')];
            }

            return this.traces.map(trace =>
                new TraceTreeItem(
                    trace,
                    undefined,
                    trace.name || trace.id.substring(0, 8),
                    vscode.TreeItemCollapsibleState.Collapsed,
                    'trace'
                )
            );
        }

        // Child level - show observations for a trace
        if (element.trace) {
            const observations = await this.client.getTraceObservations(element.trace.id);

            // Build tree structure from observations
            const rootObservations = observations.filter(o => !o.parentObservationId);

            return rootObservations.map(obs =>
                new TraceTreeItem(
                    undefined,
                    obs,
                    obs.name || obs.type,
                    observations.some(o => o.parentObservationId === obs.id)
                        ? vscode.TreeItemCollapsibleState.Collapsed
                        : vscode.TreeItemCollapsibleState.None,
                    'observation'
                )
            );
        }

        return [];
    }

    getParent(element: TraceTreeItem): vscode.ProviderResult<TraceTreeItem> {
        // Not implementing parent navigation for simplicity
        return null;
    }
}
