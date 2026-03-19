"use client";

import * as React from "react";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { ReplayEngine } from "@/components/replay/replay-engine";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Clock, Play } from "lucide-react";

export function ReplayDashboard() {
  const searchParams = useSearchParams();
  const sessionId = searchParams.get("sessionId");
  const [selectedSessionId, setSelectedSessionId] = React.useState<string | null>(sessionId);

  // Fetch available sessions
  const { data: sessionsData } = useQuery({
    queryKey: ["replay-sessions"],
    queryFn: () => api.replaySessions.list(),
    enabled: !selectedSessionId,
  });

  if (selectedSessionId) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setSelectedSessionId(null)}>
            ← Back to sessions
          </Button>
          <span className="text-sm text-muted-foreground font-mono">{selectedSessionId}</span>
        </div>
        <ReplayEngine sessionId={selectedSessionId} />
      </div>
    );
  }

  const sessions = sessionsData?.sessions ?? [];

  return (
    <div className="space-y-4">
      {sessions.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p className="text-lg font-medium">No Replay Sessions</p>
          <p className="text-sm mt-2">Record agent sessions to enable time-travel debugging and replay</p>
        </div>
      ) : (
        <div className="grid gap-3">
          {sessions.map((session: any) => (
            <Card key={session.id} className="cursor-pointer hover:border-primary/50 transition-colors" onClick={() => setSelectedSessionId(session.id)}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">{session.name || "Untitled Session"}</CardTitle>
                  <Badge variant="outline">{session.status}</Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {session.totalEvents ?? 0} events
                  </span>
                  <span>{new Date(session.createdAt).toLocaleDateString()}</span>
                  <Button size="sm" variant="ghost" className="ml-auto">
                    <Play className="h-4 w-4 mr-1" /> Replay
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
