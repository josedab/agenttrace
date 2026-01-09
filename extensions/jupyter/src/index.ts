import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';
import { ISettingRegistry } from '@jupyterlab/settingregistry';
import { INotebookTracker, NotebookPanel } from '@jupyterlab/notebook';
import { ToolbarButton } from '@jupyterlab/apputils';
import { Widget } from '@lumino/widgets';
import { requestAPI } from './handler';

/**
 * AgentTrace extension state
 */
interface AgentTraceState {
  configured: boolean;
  autoTrace: boolean;
  currentTraceId: string | null;
  sessionMetrics: SessionMetrics;
}

interface SessionMetrics {
  totalCost: number;
  totalTokens: number;
  cellExecutions: number;
  llmCalls: number;
}

/**
 * AgentTrace sidebar widget for trace visualization
 */
class AgentTraceSidebar extends Widget {
  private state: AgentTraceState;

  constructor() {
    super();
    this.addClass('agenttrace-sidebar');
    this.id = 'agenttrace-sidebar';
    this.title.label = 'AgentTrace';
    this.title.closable = true;

    this.state = {
      configured: false,
      autoTrace: true,
      currentTraceId: null,
      sessionMetrics: {
        totalCost: 0,
        totalTokens: 0,
        cellExecutions: 0,
        llmCalls: 0
      }
    };

    this.render();
  }

  private render(): void {
    this.node.innerHTML = `
      <div class="agenttrace-container">
        <div class="agenttrace-header">
          <h2>AgentTrace</h2>
          <div class="agenttrace-status ${this.state.configured ? 'connected' : 'disconnected'}">
            ${this.state.configured ? 'Connected' : 'Not Configured'}
          </div>
        </div>

        <div class="agenttrace-metrics">
          <h3>Session Metrics</h3>
          <div class="metric-grid">
            <div class="metric">
              <span class="metric-value">${this.state.sessionMetrics.cellExecutions}</span>
              <span class="metric-label">Cell Executions</span>
            </div>
            <div class="metric">
              <span class="metric-value">${this.state.sessionMetrics.llmCalls}</span>
              <span class="metric-label">LLM Calls</span>
            </div>
            <div class="metric">
              <span class="metric-value">${this.state.sessionMetrics.totalTokens}</span>
              <span class="metric-label">Total Tokens</span>
            </div>
            <div class="metric">
              <span class="metric-value">$${this.state.sessionMetrics.totalCost.toFixed(4)}</span>
              <span class="metric-label">Total Cost</span>
            </div>
          </div>
        </div>

        <div class="agenttrace-controls">
          <label class="toggle-control">
            <input type="checkbox" id="auto-trace-toggle" ${this.state.autoTrace ? 'checked' : ''}>
            <span>Auto-trace cell executions</span>
          </label>
        </div>

        <div class="agenttrace-traces">
          <h3>Recent Traces</h3>
          <div id="trace-list" class="trace-list">
            <p class="empty-state">No traces yet. Execute a cell to start tracing.</p>
          </div>
        </div>

        ${!this.state.configured ? `
          <div class="agenttrace-setup">
            <p>Configure AgentTrace to start tracing:</p>
            <code>export AGENTTRACE_API_KEY=your-api-key</code>
          </div>
        ` : ''}
      </div>
    `;

    // Add event listeners
    const toggle = this.node.querySelector('#auto-trace-toggle');
    if (toggle) {
      toggle.addEventListener('change', (e) => {
        this.state.autoTrace = (e.target as HTMLInputElement).checked;
        this.updateAutoTrace(this.state.autoTrace);
      });
    }
  }

  public async checkConfiguration(): Promise<void> {
    try {
      const config = await requestAPI<any>('config');
      this.state.configured = config.configured;
      this.state.autoTrace = config.autoTrace;
      this.render();
    } catch (error) {
      console.error('Failed to check AgentTrace configuration:', error);
    }
  }

  public updateMetrics(metrics: Partial<SessionMetrics>): void {
    this.state.sessionMetrics = {
      ...this.state.sessionMetrics,
      ...metrics
    };
    this.render();
  }

  private async updateAutoTrace(enabled: boolean): Promise<void> {
    try {
      await requestAPI<any>('config', {
        method: 'POST',
        body: JSON.stringify({ autoTrace: enabled })
      });
    } catch (error) {
      console.error('Failed to update auto-trace setting:', error);
    }
  }

  public async refreshTraces(): Promise<void> {
    if (!this.state.configured) return;

    try {
      const traces = await requestAPI<any>('traces?limit=10');
      this.renderTraceList(traces.traces);
    } catch (error) {
      console.error('Failed to fetch traces:', error);
    }
  }

  private renderTraceList(traces: any[]): void {
    const container = this.node.querySelector('#trace-list');
    if (!container) return;

    if (traces.length === 0) {
      container.innerHTML = '<p class="empty-state">No traces yet. Execute a cell to start tracing.</p>';
      return;
    }

    container.innerHTML = traces.map(trace => `
      <div class="trace-item" data-trace-id="${trace.id}">
        <div class="trace-name">${trace.name}</div>
        <div class="trace-meta">
          <span class="trace-time">${new Date(trace.createdAt).toLocaleTimeString()}</span>
          ${trace.totalCost ? `<span class="trace-cost">$${trace.totalCost.toFixed(4)}</span>` : ''}
        </div>
      </div>
    `).join('');
  }
}

/**
 * Cell execution tracer
 */
class CellTracer {
  private sidebar: AgentTraceSidebar;
  private activeSpans: Map<string, string> = new Map();

  constructor(sidebar: AgentTraceSidebar) {
    this.sidebar = sidebar;
  }

  public async startCellTrace(
    cellId: string,
    notebookPath: string,
    content: string
  ): Promise<string | null> {
    try {
      const result = await requestAPI<any>('cell-trace', {
        method: 'POST',
        body: JSON.stringify({
          cellId,
          notebookPath,
          content
        })
      });
      this.activeSpans.set(cellId, result.id);
      return result.id;
    } catch (error) {
      console.error('Failed to start cell trace:', error);
      return null;
    }
  }

  public async endCellTrace(
    cellId: string,
    output: string | null,
    error: string | null
  ): Promise<void> {
    const spanId = this.activeSpans.get(cellId);
    if (!spanId) return;

    try {
      await requestAPI<any>('cell-trace', {
        method: 'PATCH',
        body: JSON.stringify({
          spanId,
          output,
          error
        })
      });
      this.activeSpans.delete(cellId);

      // Update sidebar metrics
      const metrics = await requestAPI<any>('metrics');
      this.sidebar.updateMetrics(metrics);
      this.sidebar.refreshTraces();
    } catch (error) {
      console.error('Failed to end cell trace:', error);
    }
  }
}

/**
 * Initialization data for the agenttrace-jupyter extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: '@agenttrace/jupyter:plugin',
  description: 'AgentTrace integration for JupyterLab',
  autoStart: true,
  requires: [INotebookTracker],
  optional: [ISettingRegistry],
  activate: (
    app: JupyterFrontEnd,
    notebookTracker: INotebookTracker,
    settingRegistry: ISettingRegistry | null
  ) => {
    console.log('AgentTrace extension activated');

    // Create sidebar widget
    const sidebar = new AgentTraceSidebar();
    sidebar.checkConfiguration();

    // Add to right sidebar
    app.shell.add(sidebar, 'right', { rank: 1000 });

    // Create cell tracer
    const tracer = new CellTracer(sidebar);

    // Add toolbar button to notebooks
    notebookTracker.widgetAdded.connect((sender, panel: NotebookPanel) => {
      const button = new ToolbarButton({
        label: 'AgentTrace',
        tooltip: 'Open AgentTrace panel',
        onClick: () => {
          app.shell.activateById(sidebar.id);
        }
      });
      panel.toolbar.insertItem(10, 'agenttrace', button);

      // Track cell executions
      const notebook = panel.content;
      notebook.model?.cells.changed.connect(() => {
        // Handle cell changes
      });
    });

    // Load settings if available
    if (settingRegistry) {
      settingRegistry
        .load(plugin.id)
        .then(settings => {
          console.log('AgentTrace settings loaded:', settings.composite);
        })
        .catch(reason => {
          console.error('Failed to load AgentTrace settings:', reason);
        });
    }
  }
};

export default plugin;
