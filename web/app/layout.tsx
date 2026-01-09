import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { QueryProvider } from "@/components/providers/query-provider";
import { AuthProvider } from "@/components/providers/auth-provider";
import { Toaster } from "@/components/ui/sonner";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: {
    default: "AgentTrace",
    template: "%s | AgentTrace",
  },
  description: "Observability platform for AI coding agents",
  keywords: ["AI", "observability", "LLM", "tracing", "monitoring", "agents"],
  authors: [{ name: "AgentTrace" }],
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://agenttrace.dev",
    title: "AgentTrace",
    description: "Observability platform for AI coding agents",
    siteName: "AgentTrace",
  },
  twitter: {
    card: "summary_large_image",
    title: "AgentTrace",
    description: "Observability platform for AI coding agents",
  },
  icons: {
    icon: "/favicon.ico",
    shortcut: "/favicon-16x16.png",
    apple: "/apple-touch-icon.png",
  },
  manifest: "/site.webmanifest",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <AuthProvider>
            <QueryProvider>
              {children}
              <Toaster />
            </QueryProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
