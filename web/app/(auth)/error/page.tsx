"use client";

import * as React from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { AlertCircle } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";

const errorMessages: Record<string, string> = {
  Configuration: "There is a problem with the server configuration.",
  AccessDenied: "You do not have permission to sign in.",
  Verification: "The verification link has expired or has already been used.",
  OAuthSignin: "Error in constructing an authorization URL.",
  OAuthCallback: "Error in handling the response from the OAuth provider.",
  OAuthCreateAccount: "Could not create OAuth provider user in the database.",
  EmailCreateAccount: "Could not create email provider user in the database.",
  Callback: "Error in the OAuth callback handler route.",
  OAuthAccountNotLinked: "Email is already associated with another account. Sign in with the same account you used originally.",
  EmailSignin: "Check your email inbox for a sign in link.",
  CredentialsSignin: "Sign in failed. Check the details you provided are correct.",
  SessionRequired: "Please sign in to access this page.",
  Default: "An unexpected error occurred. Please try again.",
};

export default function AuthErrorPage() {
  const searchParams = useSearchParams();
  const error = searchParams.get("error") || "Default";
  const errorMessage = errorMessages[error] || errorMessages.Default;

  return (
    <Card>
      <CardHeader className="space-y-1 text-center">
        <div className="flex justify-center mb-4">
          <div className="h-12 w-12 rounded-lg bg-destructive flex items-center justify-center">
            <AlertCircle className="h-6 w-6 text-destructive-foreground" />
          </div>
        </div>
        <CardTitle className="text-2xl font-bold">Authentication Error</CardTitle>
        <CardDescription>
          There was a problem signing you in
        </CardDescription>
      </CardHeader>
      <CardContent className="text-center">
        <div className="rounded-lg bg-destructive/10 p-4 text-destructive">
          <p className="text-sm">{errorMessage}</p>
        </div>
        {error === "OAuthAccountNotLinked" && (
          <p className="mt-4 text-sm text-muted-foreground">
            If you previously signed in with a different method, please use that same method to sign in.
          </p>
        )}
      </CardContent>
      <CardFooter className="flex flex-col space-y-4">
        <Button asChild className="w-full">
          <Link href="/sign-in">Try again</Link>
        </Button>
        <p className="text-sm text-muted-foreground text-center">
          Need help?{" "}
          <Link href="/support" className="text-primary hover:underline">
            Contact support
          </Link>
        </p>
      </CardFooter>
    </Card>
  );
}
