"use client";

import * as React from "react";
import {
  Link2,
  Plus,
  Play,
  Trash2,
  Copy,
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  Zap,
  Code2,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface Adapter {
  id: string;
  name: string;
  framework: string;
  version: string;
  status: "registered" | "active" | "inactive" | "deprecated";
  stats: {
    totalTraces: number;
    totalSpans: number;
    avgLatencyMs: number;
    errorRate: number;
    lastActiveAt?: string;
  };
  capabilities: string[];
  createdAt: string;
}

interface AdapterTemplate {
  framework: string;
  name: string;
  description: string;
  setupCode: string;
  language: string;
  dependencies: string[];
}

const frameworkIcons: Record<string, string> = {
  langchain: "🦜",
  crewai: "🚀",
  autogen: "🤖",
  langgraph: "📊",
  openhands: "🤲",
  semantic_kernel: "🧠",
  custom: "⚙️",
};

const statusConfig: Record<string, { label: string; variant: "default" | "secondary" | "destructive" | "outline" }> = {
  registered: { label: "Registered", variant: "outline" },
  active: { label: "Active", variant: "default" },
  inactive: { label: "Inactive", variant: "secondary" },
  deprecated: { label: "Deprecated", variant: "destructive" },
};

const sampleTemplates: AdapterTemplate[] = [
  {
    framework: "langchain",
    name: "LangChain",
    description: "Automatic trace capture for LangChain agents, chains, and tools",
    language: "python",
    setupCode: `from agenttrace.adapters import LangChainAdapter

adapter = LangChainAdapter(api_key="your-key")
adapter.instrument()

# Your LangChain code — traces are captured automatically
from langchain.agents import initialize_agent
agent = initialize_agent(tools, llm, agent="zero-shot-react-description")
agent.run("What is the weather in SF?")`,
    dependencies: ["agenttrace[langchain]", "langchain"],
  },
  {
    framework: "crewai",
    name: "CrewAI",
    description: "Trace capture for CrewAI multi-agent orchestrations",
    language: "python",
    setupCode: `from agenttrace.adapters import CrewAIAdapter

adapter = CrewAIAdapter(api_key="your-key")
adapter.instrument()

# Your CrewAI code — traces are captured automatically
from crewai import Agent, Task, Crew
crew = Crew(agents=[...], tasks=[...])
crew.kickoff()`,
    dependencies: ["agenttrace[crewai]", "crewai"],
  },
  {
    framework: "autogen",
    name: "AutoGen",
    description: "Trace capture for Microsoft AutoGen multi-agent conversations",
    language: "python",
    setupCode: `from agenttrace.adapters import AutoGenAdapter

adapter = AutoGenAdapter(api_key="your-key")
adapter.instrument()

# Your AutoGen code — traces are captured automatically
import autogen
assistant = autogen.AssistantAgent("assistant", llm_config=config)
user = autogen.UserProxyAgent("user")
user.initiate_chat(assistant, message="Hello")`,
    dependencies: ["agenttrace[autogen]", "pyautogen"],
  },
  {
    framework: "langgraph",
    name: "LangGraph",
    description: "Trace capture for LangGraph stateful agent workflows",
    language: "python",
    setupCode: `from agenttrace.adapters import LangGraphAdapter

adapter = LangGraphAdapter(api_key="your-key")
adapter.instrument()

# Your LangGraph code — traces are captured automatically
from langgraph.graph import StateGraph
graph = StateGraph(AgentState)
graph.add_node("agent", agent_node)
app = graph.compile()
app.invoke({"messages": [...]})`,
    dependencies: ["agenttrace[langgraph]", "langgraph"],
  },
];

export function AdapterManager() {
  const [adapters] = React.useState<Adapter[]>([]);
  const [showRegister, setShowRegister] = React.useState(false);
  const [selectedFramework, setSelectedFramework] = React.useState<string>("");
  const [adapterName, setAdapterName] = React.useState("");
  const [copiedCode, setCopiedCode] = React.useState<string | null>(null);

  const handleCopyCode = (code: string, framework: string) => {
    navigator.clipboard.writeText(code);
    setCopiedCode(framework);
    setTimeout(() => setCopiedCode(null), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Stats Overview */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Adapters</CardDescription>
            <CardTitle className="text-2xl">{adapters.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Active</CardDescription>
            <CardTitle className="text-2xl">
              {adapters.filter((a) => a.status === "active").length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Traces</CardDescription>
            <CardTitle className="text-2xl">
              {adapters.reduce((sum, a) => sum + a.stats.totalTraces, 0)}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Avg Error Rate</CardDescription>
            <CardTitle className="text-2xl">
              {adapters.length > 0
                ? (adapters.reduce((sum, a) => sum + a.stats.errorRate, 0) / adapters.length * 100).toFixed(1) + "%"
                : "—"}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      {/* Registered Adapters */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Link2 className="h-5 w-5" />
              Registered Adapters
            </CardTitle>
            <CardDescription>
              Manage your agent framework integrations
            </CardDescription>
          </div>
          <Dialog open={showRegister} onOpenChange={setShowRegister}>
            <DialogTrigger asChild>
              <Button size="sm">
                <Plus className="h-4 w-4 mr-1" />
                Register Adapter
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Register New Adapter</DialogTitle>
                <DialogDescription>
                  Connect an agent framework for automatic trace capture.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label>Adapter Name</Label>
                  <Input
                    placeholder="My LangChain Adapter"
                    value={adapterName}
                    onChange={(e) => setAdapterName(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Framework</Label>
                  <Select value={selectedFramework} onValueChange={setSelectedFramework}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select framework" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="langchain">🦜 LangChain</SelectItem>
                      <SelectItem value="crewai">🚀 CrewAI</SelectItem>
                      <SelectItem value="autogen">🤖 AutoGen</SelectItem>
                      <SelectItem value="langgraph">📊 LangGraph</SelectItem>
                      <SelectItem value="openhands">🤲 OpenHands</SelectItem>
                      <SelectItem value="semantic_kernel">🧠 Semantic Kernel</SelectItem>
                      <SelectItem value="custom">⚙️ Custom</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setShowRegister(false)}>
                  Cancel
                </Button>
                <Button
                  disabled={!adapterName || !selectedFramework}
                  onClick={() => setShowRegister(false)}
                >
                  Register
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          {adapters.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Zap className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No adapters registered yet.</p>
              <p className="text-xs mt-1">
                Register an adapter to start capturing traces from your agent framework.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Adapter</TableHead>
                  <TableHead>Framework</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Traces</TableHead>
                  <TableHead>Avg Latency</TableHead>
                  <TableHead>Error Rate</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {adapters.map((adapter) => (
                  <TableRow key={adapter.id}>
                    <TableCell className="font-medium">{adapter.name}</TableCell>
                    <TableCell>
                      <span className="mr-1">{frameworkIcons[adapter.framework] || "⚙️"}</span>
                      {adapter.framework}
                    </TableCell>
                    <TableCell>
                      <Badge variant={statusConfig[adapter.status]?.variant || "outline"}>
                        {statusConfig[adapter.status]?.label || adapter.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{adapter.stats.totalTraces}</TableCell>
                    <TableCell>{adapter.stats.avgLatencyMs.toFixed(0)}ms</TableCell>
                    <TableCell>{(adapter.stats.errorRate * 100).toFixed(1)}%</TableCell>
                    <TableCell className="text-right space-x-1">
                      <Button variant="ghost" size="sm" title="Test adapter">
                        <Play className="h-3 w-3" />
                      </Button>
                      <Button variant="ghost" size="sm" title="Delete adapter">
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Framework Setup Templates */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Code2 className="h-5 w-5" />
            Quick Setup
          </CardTitle>
          <CardDescription>
            Copy setup code to integrate AgentTrace with your framework in seconds.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="langchain">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="langchain">🦜 LangChain</TabsTrigger>
              <TabsTrigger value="crewai">🚀 CrewAI</TabsTrigger>
              <TabsTrigger value="autogen">🤖 AutoGen</TabsTrigger>
              <TabsTrigger value="langgraph">📊 LangGraph</TabsTrigger>
            </TabsList>
            {sampleTemplates.map((template) => (
              <TabsContent key={template.framework} value={template.framework}>
                <div className="space-y-3">
                  <p className="text-sm text-muted-foreground">{template.description}</p>
                  <div className="flex gap-2">
                    {template.dependencies.map((dep) => (
                      <Badge key={dep} variant="secondary" className="font-mono text-xs">
                        {dep}
                      </Badge>
                    ))}
                  </div>
                  <div className="relative">
                    <pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
                      <code>{template.setupCode}</code>
                    </pre>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="absolute top-2 right-2"
                      onClick={() => handleCopyCode(template.setupCode, template.framework)}
                    >
                      {copiedCode === template.framework ? (
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>

      {/* Lifecycle Hooks Reference */}
      <Card>
        <CardHeader>
          <CardTitle>Lifecycle Hooks</CardTitle>
          <CardDescription>
            Hooks available for all adapters to customize trace capture behavior.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
            {[
              { name: "on_start", desc: "Triggered when a trace begins", icon: <Play className="h-4 w-4 text-green-500" /> },
              { name: "on_complete", desc: "Triggered when a trace ends successfully", icon: <CheckCircle2 className="h-4 w-4 text-green-500" /> },
              { name: "on_error", desc: "Triggered when an error occurs", icon: <XCircle className="h-4 w-4 text-red-500" /> },
              { name: "on_tool_call", desc: "Triggered on agent tool invocations", icon: <Zap className="h-4 w-4 text-blue-500" /> },
              { name: "on_llm_call", desc: "Triggered on LLM API calls", icon: <Clock className="h-4 w-4 text-purple-500" /> },
              { name: "on_agent_action", desc: "Triggered on agent decision steps", icon: <AlertTriangle className="h-4 w-4 text-yellow-500" /> },
            ].map((hook) => (
              <div key={hook.name} className="flex items-start gap-3 rounded-md border p-3">
                {hook.icon}
                <div>
                  <p className="text-sm font-mono font-medium">{hook.name}</p>
                  <p className="text-xs text-muted-foreground">{hook.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
