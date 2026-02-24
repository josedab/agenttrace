import * as vscode from 'vscode';

interface TraceAnnotation {
  filePath: string;
  lineNumber: number;
  type: 'llm_call' | 'tool_call' | 'file_edit' | 'error';
  title: string;
  description: string;
  cost?: number;
  tokens?: number;
  model?: string;
  traceId: string;
  observationId: string;
  timestamp: string;
}

export class InlineTraceProvider implements vscode.Disposable {
  private decorationType: vscode.TextEditorDecorationType;
  private errorDecorationType: vscode.TextEditorDecorationType;
  private costDecorationType: vscode.TextEditorDecorationType;
  private annotations: Map<string, TraceAnnotation[]> = new Map();
  private disposables: vscode.Disposable[] = [];

  constructor() {
    this.decorationType = vscode.window.createTextEditorDecorationType({
      after: {
        margin: '0 0 0 1em',
        color: new vscode.ThemeColor('editorCodeLens.foreground'),
        fontStyle: 'italic',
      },
      isWholeLine: true,
    });

    this.errorDecorationType = vscode.window.createTextEditorDecorationType({
      after: {
        margin: '0 0 0 1em',
        color: new vscode.ThemeColor('errorForeground'),
        fontStyle: 'italic',
      },
      backgroundColor: new vscode.ThemeColor('diffEditor.removedTextBackground'),
      isWholeLine: true,
    });

    this.costDecorationType = vscode.window.createTextEditorDecorationType({
      after: {
        margin: '0 0 0 1em',
        color: new vscode.ThemeColor('charts.yellow'),
        fontStyle: 'italic',
      },
    });

    // Listen for editor changes
    this.disposables.push(
      vscode.window.onDidChangeActiveTextEditor((editor) => {
        if (editor) {
          this.updateDecorations(editor);
        }
      })
    );
  }

  setAnnotations(filePath: string, annotations: TraceAnnotation[]): void {
    this.annotations.set(filePath, annotations);
    const editor = vscode.window.activeTextEditor;
    if (editor && editor.document.uri.fsPath === filePath) {
      this.updateDecorations(editor);
    }
  }

  clearAnnotations(): void {
    this.annotations.clear();
    const editor = vscode.window.activeTextEditor;
    if (editor) {
      editor.setDecorations(this.decorationType, []);
      editor.setDecorations(this.errorDecorationType, []);
      editor.setDecorations(this.costDecorationType, []);
    }
  }

  private updateDecorations(editor: vscode.TextEditor): void {
    const filePath = editor.document.uri.fsPath;
    const fileAnnotations = this.annotations.get(filePath) || [];

    const normalDecorations: vscode.DecorationOptions[] = [];
    const errorDecorations: vscode.DecorationOptions[] = [];
    const costDecorations: vscode.DecorationOptions[] = [];

    for (const annotation of fileAnnotations) {
      const line = Math.max(0, annotation.lineNumber - 1);
      if (line >= editor.document.lineCount) continue;

      const range = new vscode.Range(line, 0, line, editor.document.lineAt(line).text.length);
      const hoverMessage = new vscode.MarkdownString();
      hoverMessage.isTrusted = true;
      hoverMessage.appendMarkdown(`**🤖 Agent Action:** ${annotation.title}\n\n`);
      if (annotation.description) {
        hoverMessage.appendMarkdown(`${annotation.description}\n\n`);
      }
      if (annotation.model) {
        hoverMessage.appendMarkdown(`**Model:** ${annotation.model}\n\n`);
      }
      if (annotation.tokens) {
        hoverMessage.appendMarkdown(`**Tokens:** ${annotation.tokens}\n\n`);
      }
      if (annotation.cost) {
        hoverMessage.appendMarkdown(`**Cost:** $${annotation.cost.toFixed(4)}\n\n`);
      }
      hoverMessage.appendMarkdown(
        `[Open Trace](command:agenttrace.openTrace?${encodeURIComponent(JSON.stringify(annotation.traceId))})`
      );

      const decoration: vscode.DecorationOptions = {
        range,
        hoverMessage,
        renderOptions: {
          after: {
            contentText: ` ◈ ${annotation.title}`,
          },
        },
      };

      if (annotation.type === 'error') {
        errorDecorations.push(decoration);
      } else if (annotation.cost && annotation.cost > 0.01) {
        costDecorations.push({
          ...decoration,
          renderOptions: {
            after: {
              contentText: ` ◈ ${annotation.title} ($${annotation.cost.toFixed(4)})`,
            },
          },
        });
      } else {
        normalDecorations.push(decoration);
      }
    }

    editor.setDecorations(this.decorationType, normalDecorations);
    editor.setDecorations(this.errorDecorationType, errorDecorations);
    editor.setDecorations(this.costDecorationType, costDecorations);
  }

  dispose(): void {
    this.decorationType.dispose();
    this.errorDecorationType.dispose();
    this.costDecorationType.dispose();
    this.disposables.forEach((d) => d.dispose());
  }
}
