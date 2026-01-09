"use client";

import * as React from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Plus, Loader2, Copy, Check, AlertTriangle } from "lucide-react";

import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";

const createKeySchema = z.object({
  name: z.string().min(1, "Name is required"),
  expiresIn: z.string(),
  scopes: z.array(z.string()).min(1, "At least one scope is required"),
});

type CreateKeyFormData = z.infer<typeof createKeySchema>;

const availableScopes = [
  { id: "traces:read", label: "Read Traces" },
  { id: "traces:write", label: "Write Traces" },
  { id: "prompts:read", label: "Read Prompts" },
  { id: "prompts:write", label: "Write Prompts" },
  { id: "datasets:read", label: "Read Datasets" },
  { id: "datasets:write", label: "Write Datasets" },
  { id: "scores:read", label: "Read Scores" },
  { id: "scores:write", label: "Write Scores" },
];

export function CreateApiKeyDialog() {
  const queryClient = useQueryClient();
  const [open, setOpen] = React.useState(false);
  const [createdKey, setCreatedKey] = React.useState<string | null>(null);
  const [copied, setCopied] = React.useState(false);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateKeyFormData>({
    resolver: zodResolver(createKeySchema),
    defaultValues: {
      expiresIn: "never",
      scopes: ["traces:read", "traces:write"],
    },
  });

  const selectedScopes = watch("scopes");

  const createMutation = useMutation({
    mutationFn: (data: CreateKeyFormData) =>
      api.apiKeys.create({
        name: data.name,
        expiresIn: data.expiresIn === "never" ? undefined : data.expiresIn,
        scopes: data.scopes,
      }),
    onSuccess: (result) => {
      toast.success("API key created");
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      setCreatedKey(result.key);
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create API key");
    },
  });

  const onSubmit = (data: CreateKeyFormData) => {
    createMutation.mutate(data);
  };

  const copyKey = () => {
    if (createdKey) {
      navigator.clipboard.writeText(createdKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleClose = () => {
    setOpen(false);
    setCreatedKey(null);
    setCopied(false);
    reset();
  };

  const toggleScope = (scopeId: string) => {
    const current = selectedScopes || [];
    const updated = current.includes(scopeId)
      ? current.filter((s) => s !== scopeId)
      : [...current, scopeId];
    setValue("scopes", updated, { shouldValidate: true });
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogTrigger asChild>
        <Button onClick={() => setOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create API Key
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {createdKey ? "API Key Created" : "Create API Key"}
          </DialogTitle>
          <DialogDescription>
            {createdKey
              ? "Copy your API key now. You won't be able to see it again."
              : "Create a new API key for SDK access."}
          </DialogDescription>
        </DialogHeader>

        {createdKey ? (
          <div className="space-y-4 py-4">
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Important</AlertTitle>
              <AlertDescription>
                This is the only time you'll see this key. Copy it now and store
                it securely.
              </AlertDescription>
            </Alert>

            <div className="flex items-center gap-2">
              <code className="flex-1 p-3 bg-muted rounded-md text-sm font-mono break-all">
                {createdKey}
              </code>
              <Button variant="outline" size="icon" onClick={copyKey}>
                {copied ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
            </div>

            <DialogFooter>
              <Button onClick={handleClose}>Done</Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={handleSubmit(onSubmit)}>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="My API Key"
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-sm text-destructive">
                    {errors.name.message}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  A friendly name to identify this key
                </p>
              </div>

              <div className="space-y-2">
                <Label>Expiration</Label>
                <Select
                  value={watch("expiresIn")}
                  onValueChange={(value) => setValue("expiresIn", value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="never">Never expires</SelectItem>
                    <SelectItem value="7d">7 days</SelectItem>
                    <SelectItem value="30d">30 days</SelectItem>
                    <SelectItem value="90d">90 days</SelectItem>
                    <SelectItem value="365d">1 year</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Scopes</Label>
                <div className="grid grid-cols-2 gap-2">
                  {availableScopes.map((scope) => (
                    <div
                      key={scope.id}
                      className="flex items-center gap-2 p-2 border rounded-md cursor-pointer hover:bg-muted"
                      onClick={() => toggleScope(scope.id)}
                    >
                      <Checkbox
                        checked={selectedScopes?.includes(scope.id)}
                        onCheckedChange={() => toggleScope(scope.id)}
                      />
                      <span className="text-sm">{scope.label}</span>
                    </div>
                  ))}
                </div>
                {errors.scopes && (
                  <p className="text-sm text-destructive">
                    {errors.scopes.message}
                  </p>
                )}
              </div>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Create Key
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
