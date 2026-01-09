"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { Plus, X, Info } from "lucide-react";

import { useCreateSSOConfiguration, CreateSSOConfigurationInput } from "@/hooks/use-sso";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const samlSchema = z.object({
  provider: z.literal("saml"),
  enabled: z.boolean().default(true),
  issuer: z.string().min(1, "Issuer is required"),
  ssoUrl: z.string().url("Must be a valid URL"),
  certificate: z.string().min(1, "Certificate is required"),
  allowedDomains: z.array(z.string()).min(1, "At least one domain is required"),
  defaultRole: z.string().optional(),
});

const oidcSchema = z.object({
  provider: z.literal("oidc"),
  enabled: z.boolean().default(true),
  clientId: z.string().min(1, "Client ID is required"),
  clientSecret: z.string().min(1, "Client Secret is required"),
  discoveryUrl: z.string().url("Must be a valid URL"),
  allowedDomains: z.array(z.string()).min(1, "At least one domain is required"),
  defaultRole: z.string().optional(),
});

const formSchema = z.discriminatedUnion("provider", [samlSchema, oidcSchema]);

type FormValues = z.infer<typeof formSchema>;

interface SSOConfigurationFormProps {
  organizationId: string;
}

export function SSOConfigurationForm({ organizationId }: SSOConfigurationFormProps) {
  const [provider, setProvider] = React.useState<"saml" | "oidc">("oidc");
  const [domainInput, setDomainInput] = React.useState("");
  const createMutation = useCreateSSOConfiguration(organizationId);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      provider: "oidc",
      enabled: true,
      allowedDomains: [],
      defaultRole: "member",
    } as FormValues,
  });

  const allowedDomains = form.watch("allowedDomains") || [];

  const addDomain = () => {
    const domain = domainInput.trim().toLowerCase();
    if (domain && !allowedDomains.includes(domain)) {
      form.setValue("allowedDomains", [...allowedDomains, domain]);
      setDomainInput("");
    }
  };

  const removeDomain = (domain: string) => {
    form.setValue(
      "allowedDomains",
      allowedDomains.filter((d) => d !== domain)
    );
  };

  const onSubmit = (data: FormValues) => {
    createMutation.mutate(data as CreateSSOConfigurationInput, {
      onSuccess: () => {
        toast.success("SSO configuration created");
        form.reset();
      },
      onError: (error: Error) => {
        toast.error(error.message || "Failed to create SSO configuration");
      },
    });
  };

  const handleProviderChange = (newProvider: "saml" | "oidc") => {
    setProvider(newProvider);
    form.reset({
      provider: newProvider,
      enabled: true,
      allowedDomains: [],
      defaultRole: "member",
    } as FormValues);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Add SSO Configuration</CardTitle>
        <CardDescription>
          Set up SAML 2.0 or OpenID Connect authentication for your organization.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <Tabs value={provider} onValueChange={(v) => handleProviderChange(v as "saml" | "oidc")}>
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="oidc">OpenID Connect (OIDC)</TabsTrigger>
                <TabsTrigger value="saml">SAML 2.0</TabsTrigger>
              </TabsList>

              <TabsContent value="oidc" className="space-y-4 mt-4">
                <FormField
                  control={form.control}
                  name="clientId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Client ID</FormLabel>
                      <FormControl>
                        <Input placeholder="your-client-id" {...field} value={field.value || ""} />
                      </FormControl>
                      <FormDescription>
                        The client ID from your identity provider.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="clientSecret"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Client Secret</FormLabel>
                      <FormControl>
                        <Input type="password" placeholder="••••••••" {...field} value={field.value || ""} />
                      </FormControl>
                      <FormDescription>
                        The client secret from your identity provider.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="discoveryUrl"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Discovery URL
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Info className="h-4 w-4 text-muted-foreground" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p>Usually ends with /.well-known/openid-configuration</p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </FormLabel>
                      <FormControl>
                        <Input
                          placeholder="https://accounts.google.com/.well-known/openid-configuration"
                          {...field}
                          value={field.value || ""}
                        />
                      </FormControl>
                      <FormDescription>
                        The OpenID Connect discovery endpoint URL.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </TabsContent>

              <TabsContent value="saml" className="space-y-4 mt-4">
                <FormField
                  control={form.control}
                  name="issuer"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Entity ID / Issuer</FormLabel>
                      <FormControl>
                        <Input placeholder="https://idp.example.com" {...field} value={field.value || ""} />
                      </FormControl>
                      <FormDescription>
                        The unique identifier for your identity provider.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="ssoUrl"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>SSO URL</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="https://idp.example.com/sso/saml"
                          {...field}
                          value={field.value || ""}
                        />
                      </FormControl>
                      <FormDescription>
                        The SAML Single Sign-On service URL.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="certificate"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>X.509 Certificate</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                          className="font-mono text-sm"
                          rows={6}
                          {...field}
                          value={field.value || ""}
                        />
                      </FormControl>
                      <FormDescription>
                        The public certificate from your identity provider.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </TabsContent>
            </Tabs>

            {/* Common fields */}
            <div className="space-y-4 pt-4 border-t">
              <FormItem>
                <FormLabel>Allowed Domains</FormLabel>
                <div className="flex gap-2">
                  <Input
                    placeholder="example.com"
                    value={domainInput}
                    onChange={(e) => setDomainInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        e.preventDefault();
                        addDomain();
                      }
                    }}
                  />
                  <Button type="button" variant="outline" onClick={addDomain}>
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
                <FormDescription>
                  Only users with email addresses from these domains can use SSO.
                </FormDescription>
                {allowedDomains.length > 0 && (
                  <div className="flex flex-wrap gap-2 mt-2">
                    {allowedDomains.map((domain) => (
                      <Badge key={domain} variant="secondary" className="gap-1">
                        {domain}
                        <button
                          type="button"
                          onClick={() => removeDomain(domain)}
                          className="ml-1 hover:text-destructive"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </Badge>
                    ))}
                  </div>
                )}
                {form.formState.errors.allowedDomains && (
                  <p className="text-sm text-destructive mt-1">
                    {form.formState.errors.allowedDomains.message}
                  </p>
                )}
              </FormItem>

              <FormField
                control={form.control}
                name="defaultRole"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Default Role</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select default role" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="viewer">Viewer</SelectItem>
                        <SelectItem value="member">Member</SelectItem>
                        <SelectItem value="admin">Admin</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      The default role assigned to new users who sign up via SSO.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="enabled"
                render={({ field }) => (
                  <FormItem className="flex items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>Enable SSO</FormLabel>
                      <FormDescription>
                        Allow users to sign in using this SSO configuration.
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <div className="flex justify-end">
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create Configuration"}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
