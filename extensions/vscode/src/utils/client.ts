import axios, { AxiosInstance, AxiosError } from 'axios';

export interface Trace {
    id: string;
    name: string;
    projectId: string;
    sessionId?: string;
    status: 'running' | 'completed' | 'error';
    startTime: string;
    endTime?: string;
    duration?: number;
    inputTokens: number;
    outputTokens: number;
    totalCost: number;
    level: 'DEBUG' | 'INFO' | 'WARNING' | 'ERROR';
    metadata?: Record<string, any>;
    tags?: string[];
    gitCommitSha?: string;
    gitBranch?: string;
}

export interface Observation {
    id: string;
    traceId: string;
    parentObservationId?: string;
    name: string;
    type: 'span' | 'generation' | 'event';
    startTime: string;
    endTime?: string;
    status: 'running' | 'completed' | 'error';
    model?: string;
    inputTokens?: number;
    outputTokens?: number;
    cost?: number;
    input?: any;
    output?: any;
    metadata?: Record<string, any>;
}

export interface Session {
    id: string;
    name?: string;
    projectId: string;
    traceCount: number;
    totalCost: number;
    firstTraceTime: string;
    lastTraceTime: string;
}

export interface GitLink {
    id: string;
    traceId: string;
    commitSha: string;
    branch?: string;
    commitMessage?: string;
    commitAuthor?: string;
    commitTimestamp: string;
    filesChanged: string[];
}

export interface CostSummary {
    today: number;
    thisWeek: number;
    thisMonth: number;
    byModel: Record<string, number>;
}

export interface TracesResponse {
    data: Trace[];
    totalCount: number;
    hasMore: boolean;
}

export interface SessionsResponse {
    data: Session[];
    totalCount: number;
}

export interface GitLinksResponse {
    data: GitLink[];
    totalCount: number;
}

export class AgentTraceClient {
    private client: AxiosInstance;
    private apiUrl: string;
    private apiKey: string;
    private projectId: string;

    constructor(apiUrl: string, apiKey: string, projectId: string) {
        this.apiUrl = apiUrl;
        this.apiKey = apiKey;
        this.projectId = projectId;
        this.client = this.createClient();
    }

    private createClient(): AxiosInstance {
        return axios.create({
            baseURL: this.apiUrl,
            headers: {
                'Authorization': `Bearer ${this.apiKey}`,
                'Content-Type': 'application/json',
            },
            timeout: 10000,
        });
    }

    updateConfig(apiUrl: string, apiKey: string, projectId: string) {
        this.apiUrl = apiUrl;
        this.apiKey = apiKey;
        this.projectId = projectId;
        this.client = this.createClient();
    }

    isConfigured(): boolean {
        return !!this.apiKey && !!this.projectId;
    }

    getProjectId(): string {
        return this.projectId;
    }

    getDashboardUrl(): string {
        return this.apiUrl.replace('/api', '').replace('api.', 'app.');
    }

    // Traces

    async listTraces(options?: {
        limit?: number;
        offset?: number;
        status?: string;
        fromTime?: string;
        toTime?: string;
        sessionId?: string;
        search?: string;
    }): Promise<TracesResponse> {
        try {
            const params = new URLSearchParams();
            if (options?.limit) params.append('limit', options.limit.toString());
            if (options?.offset) params.append('offset', options.offset.toString());
            if (options?.status) params.append('status', options.status);
            if (options?.fromTime) params.append('fromTimestamp', options.fromTime);
            if (options?.toTime) params.append('toTimestamp', options.toTime);
            if (options?.sessionId) params.append('sessionId', options.sessionId);
            if (options?.search) params.append('search', options.search);

            const response = await this.client.get(`/v1/traces?${params.toString()}`);
            return {
                data: response.data.data || [],
                totalCount: response.data.totalCount || 0,
                hasMore: response.data.hasMore || false,
            };
        } catch (error) {
            this.handleError(error);
            return { data: [], totalCount: 0, hasMore: false };
        }
    }

    async getTrace(traceId: string): Promise<Trace | null> {
        try {
            const response = await this.client.get(`/v1/traces/${traceId}`);
            return response.data;
        } catch (error) {
            this.handleError(error);
            return null;
        }
    }

    async getTraceObservations(traceId: string): Promise<Observation[]> {
        try {
            const response = await this.client.get(`/v1/traces/${traceId}/observations`);
            return response.data.data || [];
        } catch (error) {
            this.handleError(error);
            return [];
        }
    }

    // Sessions

    async listSessions(options?: {
        limit?: number;
        offset?: number;
    }): Promise<SessionsResponse> {
        try {
            const params = new URLSearchParams();
            if (options?.limit) params.append('limit', options.limit.toString());
            if (options?.offset) params.append('offset', options.offset.toString());

            const response = await this.client.get(`/v1/sessions?${params.toString()}`);
            return {
                data: response.data.data || [],
                totalCount: response.data.totalCount || 0,
            };
        } catch (error) {
            this.handleError(error);
            return { data: [], totalCount: 0 };
        }
    }

    // Git Links

    async listGitLinks(options?: {
        limit?: number;
        offset?: number;
        branch?: string;
        commitSha?: string;
    }): Promise<GitLinksResponse> {
        try {
            const params = new URLSearchParams();
            if (options?.limit) params.append('limit', options.limit.toString());
            if (options?.offset) params.append('offset', options.offset.toString());
            if (options?.branch) params.append('branch', options.branch);
            if (options?.commitSha) params.append('commitSha', options.commitSha);

            const response = await this.client.get(`/v1/git-links?${params.toString()}`);
            return {
                data: response.data.data || [],
                totalCount: response.data.totalCount || 0,
            };
        } catch (error) {
            this.handleError(error);
            return { data: [], totalCount: 0 };
        }
    }

    async getGitTimeline(branch?: string, limit?: number): Promise<any[]> {
        try {
            const params = new URLSearchParams();
            if (branch) params.append('branch', branch);
            if (limit) params.append('limit', limit.toString());

            const response = await this.client.get(`/v1/git-links/timeline?${params.toString()}`);
            return response.data.commits || [];
        } catch (error) {
            this.handleError(error);
            return [];
        }
    }

    async createGitLink(input: {
        traceId: string;
        commitSha: string;
        branch?: string;
        commitMessage?: string;
        commitAuthor?: string;
    }): Promise<GitLink | null> {
        try {
            const response = await this.client.post('/v1/git-links', input);
            return response.data;
        } catch (error) {
            this.handleError(error);
            return null;
        }
    }

    // Checkpoints

    async createCheckpoint(input: {
        traceId: string;
        name: string;
        description?: string;
        type?: string;
    }): Promise<any> {
        try {
            const response = await this.client.post('/v1/checkpoints', input);
            return response.data;
        } catch (error) {
            this.handleError(error);
            return null;
        }
    }

    // Metrics

    async getCostSummary(): Promise<CostSummary | null> {
        try {
            const response = await this.client.get('/v1/metrics/costs');
            return response.data;
        } catch (error) {
            this.handleError(error);
            return null;
        }
    }

    async getTracesByFile(filePath: string): Promise<Trace[]> {
        try {
            // Search for traces that modified this file
            const response = await this.client.get(`/v1/traces?search=${encodeURIComponent(filePath)}&limit=20`);
            return response.data.data || [];
        } catch (error) {
            this.handleError(error);
            return [];
        }
    }

    private handleError(error: unknown) {
        if (axios.isAxiosError(error)) {
            const axiosError = error as AxiosError;
            if (axiosError.response?.status === 401) {
                console.error('AgentTrace: Unauthorized - check your API key');
            } else if (axiosError.response?.status === 403) {
                console.error('AgentTrace: Forbidden - check your project permissions');
            } else {
                console.error('AgentTrace API error:', axiosError.message);
            }
        } else {
            console.error('AgentTrace error:', error);
        }
    }
}
