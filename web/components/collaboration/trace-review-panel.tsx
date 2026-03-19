"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  CheckCircle2,
  XCircle,
  Clock,
  MessageSquare,
  Users,
  Bell,
  Send,
  ThumbsUp,
  AlertTriangle,
  Settings,
  Plus,
} from "lucide-react";

// --- Types ---

interface ReviewComment {
  id: string;
  reviewId: string;
  parentId?: string;
  authorId: string;
  authorName: string;
  content: string;
  mentions?: string[];
  spanId?: string;
  resolved: boolean;
  reactions?: Record<string, number>;
  createdAt: string;
  updatedAt: string;
  replies?: ReviewComment[];
}

interface TraceReview {
  id: string;
  projectId: string;
  traceId: string;
  title: string;
  description?: string;
  requestedBy: string;
  assignedTo: string[];
  status: "pending" | "in_review" | "approved" | "rejected" | "closed";
  priority: "low" | "medium" | "high" | "urgent";
  labels?: string[];
  dueAt?: string;
  comments?: ReviewComment[];
  approvalCount: number;
  requiredApprovals: number;
  createdAt: string;
  updatedAt: string;
}

interface ReviewQueue {
  id: string;
  projectId: string;
  name: string;
  assignmentRule: string;
  reviewers: string[];
  autoAssign: boolean;
  slaHours: number;
  escalationHours: number;
  pendingCount: number;
  avgReviewTimeHours: number;
  createdAt: string;
}

interface NotificationIntegration {
  id: string;
  projectId: string;
  type: "slack" | "teams" | "github" | "email";
  name: string;
  webhookUrl?: string;
  channelId?: string;
  enabled: boolean;
  events: string[];
  createdAt: string;
}

// --- Status helpers ---

const statusConfig: Record<
  TraceReview["status"],
  { label: string; variant: "default" | "secondary" | "destructive" | "outline"; icon: React.ReactNode }
> = {
  pending: { label: "Pending", variant: "secondary", icon: <Clock className="h-3 w-3" /> },
  in_review: { label: "In Review", variant: "default", icon: <MessageSquare className="h-3 w-3" /> },
  approved: { label: "Approved", variant: "outline", icon: <CheckCircle2 className="h-3 w-3 text-green-500" /> },
  rejected: { label: "Rejected", variant: "destructive", icon: <XCircle className="h-3 w-3" /> },
  closed: { label: "Closed", variant: "secondary", icon: <XCircle className="h-3 w-3" /> },
};

const priorityColors: Record<string, string> = {
  low: "text-gray-500",
  medium: "text-blue-500",
  high: "text-orange-500",
  urgent: "text-red-600 font-semibold",
};

// --- Sample data ---

const sampleReviews: TraceReview[] = [
  {
    id: "r1",
    projectId: "p1",
    traceId: "trace-abc-123",
    title: "Review: Agent hallucination in customer support flow",
    description: "The agent generated incorrect product pricing in step 3. Needs human verification.",
    requestedBy: "user-1",
    assignedTo: ["user-2", "user-3"],
    status: "in_review",
    priority: "high",
    labels: ["hallucination", "pricing"],
    dueAt: new Date(Date.now() + 4 * 3600 * 1000).toISOString(),
    approvalCount: 1,
    requiredApprovals: 2,
    createdAt: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 1800 * 1000).toISOString(),
  },
  {
    id: "r2",
    projectId: "p1",
    traceId: "trace-def-456",
    title: "Review: Multi-step reasoning chain for order processing",
    requestedBy: "user-2",
    assignedTo: ["user-1"],
    status: "pending",
    priority: "medium",
    labels: ["reasoning"],
    approvalCount: 0,
    requiredApprovals: 1,
    createdAt: new Date(Date.now() - 3600 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 3600 * 1000).toISOString(),
  },
  {
    id: "r3",
    projectId: "p1",
    traceId: "trace-ghi-789",
    title: "Review: Tool selection accuracy in data pipeline agent",
    requestedBy: "user-3",
    assignedTo: ["user-1", "user-2"],
    status: "approved",
    priority: "low",
    approvalCount: 2,
    requiredApprovals: 2,
    createdAt: new Date(Date.now() - 86400 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 43200 * 1000).toISOString(),
  },
];

const sampleComments: ReviewComment[] = [
  {
    id: "c1",
    reviewId: "r1",
    authorId: "user-2",
    authorName: "Alice Chen",
    content: "The pricing lookup in span `tool-call-3` returned $99 instead of $149. This seems to be a retrieval issue with the product catalog tool.",
    spanId: "tool-call-3",
    resolved: false,
    reactions: { "👍": 2 },
    createdAt: new Date(Date.now() - 3600 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 3600 * 1000).toISOString(),
    replies: [
      {
        id: "c1-r1",
        reviewId: "r1",
        parentId: "c1",
        authorId: "user-3",
        authorName: "Bob Kim",
        content: "@alice I confirmed the catalog was updated yesterday. The agent may be using a cached version. cc @dave",
        mentions: ["alice", "dave"],
        resolved: false,
        createdAt: new Date(Date.now() - 1800 * 1000).toISOString(),
        updatedAt: new Date(Date.now() - 1800 * 1000).toISOString(),
      },
    ],
  },
  {
    id: "c2",
    reviewId: "r1",
    authorId: "user-1",
    authorName: "Dave Park",
    content: "I've checked the retrieval span. The embedding similarity score was 0.72 which is below our 0.85 threshold. We should flag traces with low retrieval scores automatically.",
    resolved: true,
    reactions: { "✅": 1 },
    createdAt: new Date(Date.now() - 900 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 900 * 1000).toISOString(),
  },
];

const sampleQueues: ReviewQueue[] = [
  {
    id: "q1",
    projectId: "p1",
    name: "Critical Issues",
    assignmentRule: "round_robin",
    reviewers: ["user-1", "user-2", "user-3"],
    autoAssign: true,
    slaHours: 4,
    escalationHours: 8,
    pendingCount: 3,
    avgReviewTimeHours: 2.5,
    createdAt: new Date(Date.now() - 7 * 86400 * 1000).toISOString(),
  },
  {
    id: "q2",
    projectId: "p1",
    name: "Quality Audits",
    assignmentRule: "load_balanced",
    reviewers: ["user-1", "user-2"],
    autoAssign: false,
    slaHours: 24,
    escalationHours: 48,
    pendingCount: 7,
    avgReviewTimeHours: 6.2,
    createdAt: new Date(Date.now() - 14 * 86400 * 1000).toISOString(),
  },
];

const sampleIntegrations: NotificationIntegration[] = [
  {
    id: "i1",
    projectId: "p1",
    type: "slack",
    name: "Agent Alerts Channel",
    channelId: "#agent-alerts",
    enabled: true,
    events: ["review_created", "approved", "mentioned"],
    createdAt: new Date(Date.now() - 30 * 86400 * 1000).toISOString(),
  },
  {
    id: "i2",
    projectId: "p1",
    type: "github",
    name: "GitHub Issues Sync",
    enabled: true,
    events: ["review_created", "comment_added"],
    createdAt: new Date(Date.now() - 14 * 86400 * 1000).toISOString(),
  },
];

// --- Subcomponents ---

function StatusBadge({ status }: { status: TraceReview["status"] }) {
  const config = statusConfig[status];
  return (
    <Badge variant={config.variant} className="gap-1">
      {config.icon}
      {config.label}
    </Badge>
  );
}

function SLACountdown({ dueAt }: { dueAt?: string }) {
  if (!dueAt) return null;
  const remaining = new Date(dueAt).getTime() - Date.now();
  if (remaining <= 0) {
    return <span className="text-xs text-red-600 font-medium flex items-center gap-1"><AlertTriangle className="h-3 w-3" /> SLA overdue</span>;
  }
  const hours = Math.floor(remaining / 3600000);
  const minutes = Math.floor((remaining % 3600000) / 60000);
  const isUrgent = hours < 2;
  return (
    <span className={`text-xs flex items-center gap-1 ${isUrgent ? "text-orange-500" : "text-muted-foreground"}`}>
      <Clock className="h-3 w-3" />
      {hours}h {minutes}m remaining
    </span>
  );
}

function CommentThread({ comment }: { comment: ReviewComment }) {
  return (
    <div className="space-y-3">
      <div className={`flex gap-3 ${comment.resolved ? "opacity-60" : ""}`}>
        <Avatar className="h-8 w-8">
          <AvatarFallback className="text-xs">
            {comment.authorName.split(" ").map((n) => n[0]).join("")}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">{comment.authorName}</span>
            <span className="text-xs text-muted-foreground">
              {new Date(comment.createdAt).toLocaleTimeString()}
            </span>
            {comment.spanId && (
              <Badge variant="outline" className="text-xs">span: {comment.spanId}</Badge>
            )}
            {comment.resolved && (
              <Badge variant="outline" className="text-xs text-green-600">Resolved</Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground whitespace-pre-wrap">
            {comment.content.split(/(@\w+)/g).map((part, i) =>
              part.startsWith("@") ? (
                <span key={i} className="text-blue-500 font-medium">{part}</span>
              ) : (
                <span key={i}>{part}</span>
              )
            )}
          </p>
          {comment.reactions && Object.keys(comment.reactions).length > 0 && (
            <div className="flex gap-1 mt-1">
              {Object.entries(comment.reactions).map(([emoji, count]) => (
                <button
                  key={emoji}
                  className="inline-flex items-center gap-1 text-xs border rounded-full px-2 py-0.5 hover:bg-muted"
                >
                  {emoji} {count}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
      {comment.replies && comment.replies.length > 0 && (
        <div className="ml-11 border-l-2 pl-4 space-y-3">
          {comment.replies.map((reply) => (
            <CommentThread key={reply.id} comment={reply} />
          ))}
        </div>
      )}
    </div>
  );
}

function CommentInput() {
  const [content, setContent] = useState("");

  return (
    <div className="flex gap-2 items-end">
      <Textarea
        placeholder="Add a comment... Use @username to mention someone"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="min-h-[80px] text-sm"
      />
      <Button size="sm" disabled={!content.trim()}>
        <Send className="h-4 w-4" />
      </Button>
    </div>
  );
}

// --- Review List ---

function ReviewList({
  reviews,
  onSelect,
  selectedId,
}: {
  reviews: TraceReview[];
  onSelect: (review: TraceReview) => void;
  selectedId?: string;
}) {
  return (
    <div className="space-y-2">
      {reviews.map((review) => (
        <Card
          key={review.id}
          className={`cursor-pointer transition-colors hover:bg-muted/50 ${
            selectedId === review.id ? "ring-2 ring-primary" : ""
          }`}
          onClick={() => onSelect(review)}
        >
          <CardContent className="p-4">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <StatusBadge status={review.status} />
                  <span className={`text-xs ${priorityColors[review.priority]}`}>
                    {review.priority}
                  </span>
                </div>
                <h4 className="text-sm font-medium truncate">{review.title}</h4>
                <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                  <span>Trace: {review.traceId.slice(0, 16)}...</span>
                  <span>
                    {review.approvalCount}/{review.requiredApprovals} approvals
                  </span>
                  <SLACountdown dueAt={review.dueAt} />
                </div>
              </div>
              {review.labels && review.labels.length > 0 && (
                <div className="flex gap-1 flex-shrink-0">
                  {review.labels.map((label) => (
                    <Badge key={label} variant="outline" className="text-xs">
                      {label}
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      ))}
      {reviews.length === 0 && (
        <div className="text-center py-8 text-sm text-muted-foreground">
          No reviews found.
        </div>
      )}
    </div>
  );
}

// --- Review Detail ---

function ReviewDetail({ review }: { review: TraceReview }) {
  const comments = sampleComments.filter((c) => c.reviewId === review.id);

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2 mb-2">
            <StatusBadge status={review.status} />
            <span className={`text-xs ${priorityColors[review.priority]}`}>
              {review.priority} priority
            </span>
            <SLACountdown dueAt={review.dueAt} />
          </div>
          <CardTitle className="text-lg">{review.title}</CardTitle>
          {review.description && (
            <CardDescription>{review.description}</CardDescription>
          )}
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4 text-sm">
            <span className="text-muted-foreground">Trace:</span>
            <code className="text-xs bg-muted px-2 py-1 rounded">{review.traceId}</code>
          </div>
          <div className="flex items-center gap-4 text-sm">
            <span className="text-muted-foreground">Approvals:</span>
            <span>
              {review.approvalCount} / {review.requiredApprovals}
            </span>
          </div>
          <div className="flex gap-2">
            {review.status !== "approved" && review.status !== "closed" && (
              <>
                <Button size="sm" variant="default" className="gap-1">
                  <CheckCircle2 className="h-4 w-4" /> Approve
                </Button>
                <Button size="sm" variant="destructive" className="gap-1">
                  <XCircle className="h-4 w-4" /> Reject
                </Button>
                <Button size="sm" variant="outline" className="gap-1">
                  <AlertTriangle className="h-4 w-4" /> Request Changes
                </Button>
              </>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Comments ({comments.length})
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {comments.map((comment) => (
            <CommentThread key={comment.id} comment={comment} />
          ))}
          {comments.length === 0 && (
            <p className="text-sm text-muted-foreground text-center py-4">
              No comments yet. Start the discussion.
            </p>
          )}
          <div className="border-t pt-4">
            <CommentInput />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

// --- Queue Management ---

function QueueManagement() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-base font-medium">Review Queues</h3>
        <Button size="sm" variant="outline" className="gap-1">
          <Plus className="h-4 w-4" /> New Queue
        </Button>
      </div>
      {sampleQueues.map((queue) => (
        <Card key={queue.id}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-medium">{queue.name}</h4>
              <Badge variant="outline">{queue.assignmentRule.replace("_", " ")}</Badge>
            </div>
            <div className="grid grid-cols-4 gap-4 text-xs">
              <div>
                <p className="text-muted-foreground">Pending</p>
                <p className="text-lg font-bold">{queue.pendingCount}</p>
              </div>
              <div>
                <p className="text-muted-foreground">SLA</p>
                <p className="text-lg font-bold">{queue.slaHours}h</p>
              </div>
              <div>
                <p className="text-muted-foreground">Avg Review</p>
                <p className="text-lg font-bold">{queue.avgReviewTimeHours}h</p>
              </div>
              <div>
                <p className="text-muted-foreground">Reviewers</p>
                <p className="text-lg font-bold">{queue.reviewers.length}</p>
              </div>
            </div>
            <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
              <Users className="h-3 w-3" />
              {queue.autoAssign ? "Auto-assign enabled" : "Manual assignment"}
              {queue.escalationHours > 0 && ` • Escalation after ${queue.escalationHours}h`}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

// --- Integration Settings ---

function IntegrationSettings() {
  const typeIcons: Record<string, string> = {
    slack: "💬",
    teams: "👥",
    github: "🐙",
    email: "📧",
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-base font-medium">Notification Integrations</h3>
        <Button size="sm" variant="outline" className="gap-1">
          <Plus className="h-4 w-4" /> Add Integration
        </Button>
      </div>
      {sampleIntegrations.map((integration) => (
        <Card key={integration.id}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-lg">{typeIcons[integration.type] || "🔔"}</span>
                <h4 className="text-sm font-medium">{integration.name}</h4>
                <Badge variant="outline" className="text-xs capitalize">
                  {integration.type}
                </Badge>
              </div>
              <Badge variant={integration.enabled ? "default" : "secondary"}>
                {integration.enabled ? "Active" : "Disabled"}
              </Badge>
            </div>
            {integration.channelId && (
              <p className="text-xs text-muted-foreground mb-2">
                Channel: {integration.channelId}
              </p>
            )}
            <div className="flex gap-1 flex-wrap">
              {integration.events.map((event) => (
                <Badge key={event} variant="outline" className="text-xs">
                  {event.replace("_", " ")}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

// --- Main Panel ---

export function TraceReviewPanel() {
  const [selectedReview, setSelectedReview] = useState<TraceReview | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const filteredReviews =
    statusFilter === "all"
      ? sampleReviews
      : sampleReviews.filter((r) => r.status === statusFilter);

  const stats = {
    pending: sampleReviews.filter((r) => r.status === "pending").length,
    inReview: sampleReviews.filter((r) => r.status === "in_review").length,
    approved: sampleReviews.filter((r) => r.status === "approved").length,
    overdue: sampleReviews.filter(
      (r) => r.dueAt && new Date(r.dueAt).getTime() < Date.now() && r.status !== "approved" && r.status !== "closed"
    ).length,
  };

  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card className="cursor-pointer hover:bg-muted/50" onClick={() => setStatusFilter("pending")}>
          <CardContent className="p-4">
            <p className="text-sm text-muted-foreground">Open Reviews</p>
            <p className="text-2xl font-bold">{stats.pending}</p>
            <p className="text-xs text-muted-foreground">Awaiting review</p>
          </CardContent>
        </Card>
        <Card className="cursor-pointer hover:bg-muted/50" onClick={() => setStatusFilter("in_review")}>
          <CardContent className="p-4">
            <p className="text-sm text-muted-foreground">In Review</p>
            <p className="text-2xl font-bold">{stats.inReview}</p>
            <p className="text-xs text-muted-foreground">Being reviewed</p>
          </CardContent>
        </Card>
        <Card className="cursor-pointer hover:bg-muted/50" onClick={() => setStatusFilter("approved")}>
          <CardContent className="p-4">
            <p className="text-sm text-muted-foreground">Approved</p>
            <p className="text-2xl font-bold">{stats.approved}</p>
            <p className="text-xs text-muted-foreground">Total approved</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-sm text-muted-foreground">Overdue SLA</p>
            <p className="text-2xl font-bold">{stats.overdue}</p>
            <p className="text-xs text-muted-foreground">Past due date</p>
          </CardContent>
        </Card>
      </div>

      {/* Main content */}
      <Tabs defaultValue="reviews" className="space-y-4">
        <TabsList>
          <TabsTrigger value="reviews" className="gap-1">
            <MessageSquare className="h-4 w-4" /> Reviews
          </TabsTrigger>
          <TabsTrigger value="queues" className="gap-1">
            <Users className="h-4 w-4" /> Queues
          </TabsTrigger>
          <TabsTrigger value="integrations" className="gap-1">
            <Bell className="h-4 w-4" /> Integrations
          </TabsTrigger>
          <TabsTrigger value="settings" className="gap-1">
            <Settings className="h-4 w-4" /> Settings
          </TabsTrigger>
        </TabsList>

        <TabsContent value="reviews">
          <div className="flex items-center gap-2 mb-4">
            <Button
              size="sm"
              variant={statusFilter === "all" ? "default" : "outline"}
              onClick={() => setStatusFilter("all")}
            >
              All
            </Button>
            {(["pending", "in_review", "approved", "rejected"] as const).map((status) => (
              <Button
                key={status}
                size="sm"
                variant={statusFilter === status ? "default" : "outline"}
                onClick={() => setStatusFilter(status)}
              >
                {statusConfig[status].label}
              </Button>
            ))}
            <div className="flex-1" />
            <Button size="sm" className="gap-1">
              <Plus className="h-4 w-4" /> New Review
            </Button>
          </div>

          <div className="grid gap-6 lg:grid-cols-2">
            <div>
              <ReviewList
                reviews={filteredReviews}
                onSelect={setSelectedReview}
                selectedId={selectedReview?.id}
              />
            </div>
            <div>
              {selectedReview ? (
                <ReviewDetail review={selectedReview} />
              ) : (
                <Card>
                  <CardContent className="flex items-center justify-center h-64 text-sm text-muted-foreground">
                    Select a review to see details and comments
                  </CardContent>
                </Card>
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="queues">
          <QueueManagement />
        </TabsContent>

        <TabsContent value="integrations">
          <IntegrationSettings />
        </TabsContent>

        <TabsContent value="settings">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Review Settings</CardTitle>
              <CardDescription>
                Configure default review requirements and notification preferences.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Default Required Approvals</label>
                  <input
                    type="number"
                    min={1}
                    max={10}
                    defaultValue={1}
                    className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Default SLA (hours)</label>
                  <input
                    type="number"
                    min={1}
                    defaultValue={24}
                    className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm"
                  />
                </div>
              </div>
              <Button size="sm">
                <ThumbsUp className="h-4 w-4 mr-1" /> Save Settings
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
