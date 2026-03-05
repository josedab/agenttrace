"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tabs,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  BookOpen,
  Search,
  Plus,
  ChevronDown,
  ChevronRight,
  Tag,
  Calendar,
} from "lucide-react";

interface KBEntry {
  id: string;
  title: string;
  description: string;
  category: string;
  tags: string[];
  rootCause?: string;
  pattern?: string;
  fix?: string;
  createdAt: string;
}

const CATEGORIES = [
  "all",
  "root_cause",
  "pattern",
  "fix",
  "optimization",
  "anti_pattern",
] as const;

const CATEGORY_LABELS: Record<string, string> = {
  all: "All",
  root_cause: "Root Cause",
  pattern: "Pattern",
  fix: "Fix",
  optimization: "Optimization",
  anti_pattern: "Anti-Pattern",
};

const CATEGORY_COLORS: Record<string, string> = {
  root_cause: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
  pattern: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  fix: "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300",
  optimization: "bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300",
  anti_pattern: "bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300",
};

export function KnowledgeBaseBrowser() {
  const queryClient = useQueryClient();
  const [search, setSearch] = React.useState("");
  const [category, setCategory] = React.useState("all");
  const [expandedId, setExpandedId] = React.useState<string | null>(null);
  const [showForm, setShowForm] = React.useState(false);
  const [formData, setFormData] = React.useState({
    title: "",
    description: "",
    category: "pattern",
    tags: "",
    rootCause: "",
    pattern: "",
    fix: "",
  });

  const { data: entries, isLoading } = useQuery<KBEntry[]>({
    queryKey: ["knowledge-base", search, category],
    queryFn: () =>
      api.knowledgeBase.search({
        query: search,
        category: category === "all" ? undefined : category,
      }),
  });

  const createMutation = useMutation({
    mutationFn: (data: typeof formData) =>
      api.knowledgeBase.create({
        ...data,
        tags: data.tags.split(",").map((t) => t.trim()).filter(Boolean),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["knowledge-base"] });
      setShowForm(false);
      setFormData({ title: "", description: "", category: "pattern", tags: "", rootCause: "", pattern: "", fix: "" });
    },
  });

  return (
    <div className="space-y-6">
      {/* Search & Controls */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search knowledge base..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4 mr-1" />
          Add Entry
        </Button>
      </div>

      {/* Category Filter */}
      <Tabs value={category} onValueChange={setCategory}>
        <TabsList>
          {CATEGORIES.map((cat) => (
            <TabsTrigger key={cat} value={cat} className="text-xs">
              {CATEGORY_LABELS[cat]}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {/* Add Entry Form */}
      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">New Knowledge Base Entry</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-sm font-medium">Title</label>
                <Input
                  placeholder="Entry title"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Category</label>
                <Select
                  value={formData.category}
                  onValueChange={(v) => setFormData({ ...formData, category: v })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {CATEGORIES.filter((c) => c !== "all").map((cat) => (
                      <SelectItem key={cat} value={cat}>
                        {CATEGORY_LABELS[cat]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium">Description</label>
              <Input
                placeholder="Brief description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm font-medium">Tags (comma-separated)</label>
              <Input
                placeholder="e.g. latency, gpt-4, timeout"
                value={formData.tags}
                onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
              />
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-1">
                <label className="text-sm font-medium">Root Cause (optional)</label>
                <Input
                  placeholder="Root cause details"
                  value={formData.rootCause}
                  onChange={(e) => setFormData({ ...formData, rootCause: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Pattern (optional)</label>
                <Input
                  placeholder="Pattern details"
                  value={formData.pattern}
                  onChange={(e) => setFormData({ ...formData, pattern: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Fix (optional)</label>
                <Input
                  placeholder="Fix details"
                  value={formData.fix}
                  onChange={(e) => setFormData({ ...formData, fix: e.target.value })}
                />
              </div>
            </div>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" size="sm" onClick={() => setShowForm(false)}>
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={() => createMutation.mutate(formData)}
                disabled={createMutation.isPending || !formData.title}
              >
                Add Entry
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Entries List */}
      {isLoading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-20 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      ) : (
        <div className="space-y-3">
          {entries?.map((entry) => {
            const isExpanded = expandedId === entry.id;
            return (
              <Card key={entry.id}>
                <CardContent className="pt-4">
                  <button
                    className="w-full text-left"
                    onClick={() => setExpandedId(isExpanded ? null : entry.id)}
                  >
                    <div className="flex items-start gap-3">
                      {isExpanded ? (
                        <ChevronDown className="h-4 w-4 mt-0.5 shrink-0" />
                      ) : (
                        <ChevronRight className="h-4 w-4 mt-0.5 shrink-0" />
                      )}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-medium">{entry.title}</span>
                          <Badge
                            variant="secondary"
                            className={CATEGORY_COLORS[entry.category] ?? ""}
                          >
                            {CATEGORY_LABELS[entry.category] ?? entry.category}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                          {entry.description}
                        </p>
                        <div className="flex items-center gap-3 mt-2">
                          {entry.tags.length > 0 && (
                            <div className="flex items-center gap-1 flex-wrap">
                              <Tag className="h-3 w-3 text-muted-foreground" />
                              {entry.tags.map((tag) => (
                                <Badge key={tag} variant="outline" className="text-xs">
                                  {tag}
                                </Badge>
                              ))}
                            </div>
                          )}
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {new Date(entry.createdAt).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                    </div>
                  </button>

                  {/* Expanded Details */}
                  {isExpanded && (
                    <div className="mt-4 ml-7 space-y-3 border-t pt-3">
                      {entry.rootCause && (
                        <div>
                          <span className="text-xs font-medium text-muted-foreground uppercase">Root Cause</span>
                          <p className="text-sm mt-1">{entry.rootCause}</p>
                        </div>
                      )}
                      {entry.pattern && (
                        <div>
                          <span className="text-xs font-medium text-muted-foreground uppercase">Pattern</span>
                          <p className="text-sm mt-1">{entry.pattern}</p>
                        </div>
                      )}
                      {entry.fix && (
                        <div>
                          <span className="text-xs font-medium text-muted-foreground uppercase">Fix</span>
                          <p className="text-sm mt-1">{entry.fix}</p>
                        </div>
                      )}
                      {!entry.rootCause && !entry.pattern && !entry.fix && (
                        <p className="text-sm text-muted-foreground">No additional details available</p>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}

          {(!entries || entries.length === 0) && (
            <div className="text-center py-12 text-muted-foreground">
              <BookOpen className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No knowledge base entries found</p>
              <p className="text-xs mt-1">Add entries to build your annotation knowledge base</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
