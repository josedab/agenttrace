"use client";

import * as React from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { FileUp, Loader2, Upload } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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

interface ImportItemsDialogProps {
  datasetId: string;
}

export function ImportItemsDialog({ datasetId }: ImportItemsDialogProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = React.useState(false);
  const [jsonContent, setJsonContent] = React.useState("");
  const [csvContent, setCsvContent] = React.useState("");
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const importMutation = useMutation({
    mutationFn: async (items: any[]) => {
      // Import items one by one (could be optimized with batch API)
      for (const item of items) {
        await api.datasets.addItem(datasetId, item);
      }
      return items.length;
    },
    onSuccess: (count) => {
      toast.success(`${count} items imported successfully`);
      queryClient.invalidateQueries({ queryKey: ["dataset-items", datasetId] });
      queryClient.invalidateQueries({ queryKey: ["dataset", datasetId] });
      setOpen(false);
      setJsonContent("");
      setCsvContent("");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to import items");
    },
  });

  const handleJsonImport = () => {
    try {
      const items = JSON.parse(jsonContent);
      if (!Array.isArray(items)) {
        throw new Error("JSON must be an array of items");
      }
      importMutation.mutate(items);
    } catch (error) {
      toast.error("Invalid JSON format");
    }
  };

  const handleCsvImport = () => {
    try {
      const lines = csvContent.trim().split("\n");
      if (lines.length < 2) {
        throw new Error("CSV must have a header row and at least one data row");
      }

      const headers = lines[0].split(",").map((h) => h.trim());
      const items = lines.slice(1).map((line) => {
        const values = line.split(",").map((v) => v.trim());
        const item: Record<string, string> = {};
        headers.forEach((header, index) => {
          item[header] = values[index] || "";
        });
        return {
          input: item.input || item,
          expectedOutput: item.expectedOutput || item.expected_output,
        };
      });

      importMutation.mutate(items);
    } catch (error) {
      toast.error("Invalid CSV format");
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      if (file.name.endsWith(".json")) {
        setJsonContent(content);
      } else if (file.name.endsWith(".csv")) {
        setCsvContent(content);
      }
    };
    reader.readAsText(file);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          <FileUp className="h-4 w-4 mr-2" />
          Import
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Import Dataset Items</DialogTitle>
          <DialogDescription>
            Import multiple items from JSON or CSV.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="json">
          <TabsList className="w-full">
            <TabsTrigger value="json" className="flex-1">JSON</TabsTrigger>
            <TabsTrigger value="csv" className="flex-1">CSV</TabsTrigger>
          </TabsList>

          <TabsContent value="json" className="space-y-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>JSON Array</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <Upload className="h-4 w-4 mr-1" />
                  Upload File
                </Button>
              </div>
              <Textarea
                value={jsonContent}
                onChange={(e) => setJsonContent(e.target.value)}
                placeholder={`[
  {"input": "What is 2+2?", "expectedOutput": "4"},
  {"input": "What color is the sky?", "expectedOutput": "Blue"}
]`}
                rows={10}
                className="font-mono text-sm"
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={handleJsonImport}
                disabled={importMutation.isPending || !jsonContent}
              >
                {importMutation.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Import JSON
              </Button>
            </DialogFooter>
          </TabsContent>

          <TabsContent value="csv" className="space-y-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>CSV Content</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <Upload className="h-4 w-4 mr-1" />
                  Upload File
                </Button>
              </div>
              <Textarea
                value={csvContent}
                onChange={(e) => setCsvContent(e.target.value)}
                placeholder={`input,expectedOutput
"What is 2+2?","4"
"What color is the sky?","Blue"`}
                rows={10}
                className="font-mono text-sm"
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <Button
                onClick={handleCsvImport}
                disabled={importMutation.isPending || !csvContent}
              >
                {importMutation.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Import CSV
              </Button>
            </DialogFooter>
          </TabsContent>
        </Tabs>

        <input
          ref={fileInputRef}
          type="file"
          accept=".json,.csv"
          onChange={handleFileUpload}
          className="hidden"
        />
      </DialogContent>
    </Dialog>
  );
}
