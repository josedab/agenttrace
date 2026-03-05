"use client";

import * as React from "react";
import {
  Search,
  Star,
  Download,
  ShieldCheck,
  Package,
  ArrowUpDown,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";

interface EvalDataset {
  id: string;
  name: string;
  description: string;
  author: string;
  taskType: string;
  downloads: number;
  rating: number;
  sampleCount: number;
  verified: boolean;
  category: string;
}

const categories = [
  "All",
  "Code Generation",
  "Text Summarization",
  "Question Answering",
  "Translation",
  "Classification",
  "Reasoning",
] as const;

type SortOption = "popular" | "newest" | "rating" | "downloads";

export function EvalMarketplaceBrowser() {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [activeCategory, setActiveCategory] = React.useState<string>("All");
  const [sortBy, setSortBy] = React.useState<SortOption>("popular");

  // TODO: Replace with useEvalMarketplace({ search, category, sort }) hook when available
  const isLoading = false;
  const datasets: EvalDataset[] = [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search datasets..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button
          variant="outline"
          size="sm"
          className="gap-2"
          onClick={() =>
            setSortBy((prev) =>
              prev === "popular"
                ? "newest"
                : prev === "newest"
                  ? "rating"
                  : prev === "rating"
                    ? "downloads"
                    : "popular"
            )
          }
        >
          <ArrowUpDown className="h-4 w-4" />
          Sort: {sortBy}
        </Button>
      </div>

      <Tabs value={activeCategory} onValueChange={setActiveCategory}>
        <TabsList>
          {categories.map((cat) => (
            <TabsTrigger key={cat} value={cat}>
              {cat}
            </TabsTrigger>
          ))}
        </TabsList>

        {categories.map((cat) => (
          <TabsContent key={cat} value={cat} className="mt-6">
            {isLoading ? (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {Array.from({ length: 6 }).map((_, i) => (
                  <div
                    key={i}
                    className="h-48 bg-muted animate-pulse rounded-lg"
                  />
                ))}
              </div>
            ) : datasets.length === 0 ? (
              <div className="text-center py-12 text-muted-foreground">
                <Package className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p className="text-lg font-medium">No datasets found</p>
                <p className="text-sm mt-2">
                  Try adjusting your search or browse a different category
                </p>
              </div>
            ) : (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {datasets.map((dataset) => (
                  <Card key={dataset.id} className="flex flex-col">
                    <CardHeader className="pb-3">
                      <div className="flex items-start justify-between">
                        <CardTitle className="text-base leading-tight">
                          {dataset.name}
                        </CardTitle>
                        {dataset.verified && (
                          <ShieldCheck className="h-4 w-4 shrink-0 text-blue-500" />
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground line-clamp-2">
                        {dataset.description}
                      </p>
                    </CardHeader>
                    <CardContent className="flex-1 flex flex-col justify-between gap-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge variant="outline" className="text-xs">
                          {dataset.taskType}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          by {dataset.author}
                        </span>
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Star className="h-3 w-3 text-yellow-500 fill-yellow-500" />
                            {dataset.rating.toFixed(1)}
                          </span>
                          <span className="flex items-center gap-1">
                            <Download className="h-3 w-3" />
                            {dataset.downloads.toLocaleString()}
                          </span>
                          <span>
                            {dataset.sampleCount.toLocaleString()} samples
                          </span>
                        </div>
                        <Button size="sm" variant="outline">
                          Import
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}
