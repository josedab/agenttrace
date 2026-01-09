import * as vscode from 'vscode';
import { AgentTraceClient } from '../utils/client';

interface GitTimelineEntry {
    commitSha: string;
    commitMessage: string;
    commitAuthor: string;
    commitTime: string;
    branch: string;
    traceCount: number;
    traceIds: string[];
}

export class GitHistoryTreeItem extends vscode.TreeItem {
    constructor(
        public readonly entry?: GitTimelineEntry,
        public readonly label: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly itemType: 'commit' | 'trace' | 'info' | 'loading' | 'error'
    ) {
        super(label, collapsibleState);
        this.contextValue = itemType;
        this.setupItem();
    }

    private setupItem() {
        if (this.entry) {
            this.setupCommitItem();
        } else if (this.itemType === 'loading') {
            this.iconPath = new vscode.ThemeIcon('loading~spin');
        } else if (this.itemType === 'error') {
            this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('errorForeground'));
        }
    }

    private setupCommitItem() {
        if (!this.entry) return;

        this.iconPath = new vscode.ThemeIcon('git-commit', new vscode.ThemeColor('charts.green'));

        // Description: author, time ago, trace count
        const timeAgo = this.getTimeAgo(new Date(this.entry.commitTime));
        this.description = `${this.entry.commitAuthor} | ${timeAgo} | ${this.entry.traceCount} traces`;

        // Tooltip
        const tooltip = new vscode.MarkdownString();
        tooltip.appendMarkdown(`### ${this.entry.commitMessage.split('\n')[0]}\n\n`);
        tooltip.appendMarkdown(`**SHA:** \`${this.entry.commitSha}\`\n\n`);
        tooltip.appendMarkdown(`**Author:** ${this.entry.commitAuthor}\n\n`);
        tooltip.appendMarkdown(`**Branch:** ${this.entry.branch}\n\n`);
        tooltip.appendMarkdown(`**Time:** ${new Date(this.entry.commitTime).toLocaleString()}\n\n`);
        tooltip.appendMarkdown(`**Traces:** ${this.entry.traceCount}`);
        this.tooltip = tooltip;
    }

    private getTimeAgo(date: Date): string {
        const now = new Date();
        const diff = now.getTime() - date.getTime();
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);

        if (minutes < 1) return 'just now';
        if (minutes < 60) return `${minutes}m ago`;
        if (hours < 24) return `${hours}h ago`;
        if (days < 7) return `${days}d ago`;
        return date.toLocaleDateString();
    }
}

export class GitHistoryTreeProvider implements vscode.TreeDataProvider<GitHistoryTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<GitHistoryTreeItem | undefined | null | void> = new vscode.EventEmitter<GitHistoryTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<GitHistoryTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    private timeline: GitTimelineEntry[] = [];
    private loading = false;
    private error: string | null = null;
    private branch: string | undefined;

    constructor(private client: AgentTraceClient) {
        this.loadTimeline();
    }

    refresh(): void {
        this.loadTimeline();
    }

    setBranch(branch: string | undefined) {
        this.branch = branch;
        this.loadTimeline();
    }

    private async loadTimeline() {
        if (!this.client.isConfigured()) {
            this.error = 'Not configured';
            this._onDidChangeTreeData.fire();
            return;
        }

        this.loading = true;
        this.error = null;
        this._onDidChangeTreeData.fire();

        try {
            this.timeline = await this.client.getGitTimeline(this.branch, 50);
            this.loading = false;
            this._onDidChangeTreeData.fire();
        } catch (error) {
            this.loading = false;
            this.error = 'Failed to load git history';
            this._onDidChangeTreeData.fire();
        }
    }

    getTreeItem(element: GitHistoryTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: GitHistoryTreeItem): Promise<GitHistoryTreeItem[]> {
        if (element) {
            // Show trace IDs for a commit
            if (element.entry && element.entry.traceIds.length > 0) {
                return element.entry.traceIds.map(traceId =>
                    new GitHistoryTreeItem(
                        undefined,
                        `Trace: ${traceId.substring(0, 12)}...`,
                        vscode.TreeItemCollapsibleState.None,
                        'trace'
                    )
                );
            }
            return [];
        }

        if (this.loading) {
            return [new GitHistoryTreeItem(undefined, 'Loading git history...', vscode.TreeItemCollapsibleState.None, 'loading')];
        }

        if (this.error) {
            return [new GitHistoryTreeItem(undefined, this.error, vscode.TreeItemCollapsibleState.None, 'error')];
        }

        if (this.timeline.length === 0) {
            return [new GitHistoryTreeItem(undefined, 'No git history found', vscode.TreeItemCollapsibleState.None, 'info')];
        }

        return this.timeline.map(entry =>
            new GitHistoryTreeItem(
                entry,
                `${entry.commitSha.substring(0, 7)} ${entry.commitMessage.split('\n')[0].substring(0, 50)}`,
                entry.traceCount > 0 ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None,
                'commit'
            )
        );
    }

    getParent(element: GitHistoryTreeItem): vscode.ProviderResult<GitHistoryTreeItem> {
        return null;
    }
}
