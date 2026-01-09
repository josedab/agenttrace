import * as vscode from 'vscode';
import { AgentTraceClient, CostSummary } from '../utils/client';

export class StatusBarManager implements vscode.Disposable {
    private statusBarItem: vscode.StatusBarItem;
    private costSummary: CostSummary | null = null;
    private loading = false;

    constructor(private client: AgentTraceClient) {
        this.statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Right,
            100
        );

        this.statusBarItem.command = 'agenttrace.showCostSummary';
        this.statusBarItem.tooltip = 'Click to view AgentTrace cost summary';

        this.refresh();
    }

    show() {
        this.statusBarItem.show();
    }

    hide() {
        this.statusBarItem.hide();
    }

    async refresh() {
        if (!this.client.isConfigured()) {
            this.statusBarItem.text = '$(warning) AgentTrace: Not configured';
            this.statusBarItem.tooltip = 'Click to configure AgentTrace';
            this.statusBarItem.command = 'agenttrace.configure';
            return;
        }

        this.loading = true;
        this.statusBarItem.text = '$(loading~spin) AgentTrace';

        try {
            this.costSummary = await this.client.getCostSummary();

            if (this.costSummary) {
                const todayCost = this.costSummary.today.toFixed(2);
                this.statusBarItem.text = `$(graph) AgentTrace: $${todayCost} today`;
                this.statusBarItem.tooltip = this.buildTooltip();
                this.statusBarItem.command = 'agenttrace.showCostSummary';
            } else {
                this.statusBarItem.text = '$(graph) AgentTrace';
                this.statusBarItem.tooltip = 'AgentTrace - Click to view dashboard';
            }
        } catch (error) {
            this.statusBarItem.text = '$(error) AgentTrace';
            this.statusBarItem.tooltip = 'Failed to load cost data';
        }

        this.loading = false;
    }

    private buildTooltip(): string {
        if (!this.costSummary) return 'AgentTrace';

        const lines = [
            'AgentTrace Cost Summary',
            '',
            `Today: $${this.costSummary.today.toFixed(4)}`,
            `This Week: $${this.costSummary.thisWeek.toFixed(4)}`,
            `This Month: $${this.costSummary.thisMonth.toFixed(4)}`,
        ];

        if (this.costSummary.byModel && Object.keys(this.costSummary.byModel).length > 0) {
            lines.push('', 'By Model:');
            for (const [model, cost] of Object.entries(this.costSummary.byModel)) {
                lines.push(`  ${model}: $${cost.toFixed(4)}`);
            }
        }

        lines.push('', 'Click to view full cost breakdown');

        return lines.join('\n');
    }

    getCostSummary(): CostSummary | null {
        return this.costSummary;
    }

    dispose() {
        this.statusBarItem.dispose();
    }
}
