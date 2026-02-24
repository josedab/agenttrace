package io.agenttrace.plugin.api

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.MediaType.Companion.toMediaType

/**
 * IDE Trace View API client for fetching file-level trace annotations
 * from the AgentTrace backend. Used by the inline annotation provider
 * to display trace data directly in the editor gutter.
 */
class IDETraceViewClient(
    private val httpClient: OkHttpClient,
    private val baseUrl: String,
    private val apiKey: String,
    private val projectId: String
) {
    private val gson = Gson()
    private val jsonMediaType = "application/json".toMediaType()

    /**
     * Fetch trace annotations for a single file
     */
    fun getFileMapping(filePath: String): FileTraceMapping? {
        val request = Request.Builder()
            .url("$baseUrl/api/public/ide/file-mapping?filePath=${java.net.URLEncoder.encode(filePath, "UTF-8")}")
            .header("Authorization", "Bearer $apiKey")
            .header("X-Project-ID", projectId)
            .get()
            .build()

        return try {
            val response = httpClient.newCall(request).execute()
            if (response.isSuccessful) {
                response.body?.string()?.let {
                    gson.fromJson(it, FileTraceMapping::class.java)
                }
            } else null
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Fetch trace annotations for multiple files in a single request
     */
    fun getBatchMappings(filePaths: List<String>): List<FileTraceMapping> {
        val body = gson.toJson(mapOf("filePaths" to filePaths))
            .toRequestBody(jsonMediaType)

        val request = Request.Builder()
            .url("$baseUrl/api/public/ide/batch-mappings")
            .header("Authorization", "Bearer $apiKey")
            .header("X-Project-ID", projectId)
            .post(body)
            .build()

        return try {
            val response = httpClient.newCall(request).execute()
            if (response.isSuccessful) {
                response.body?.string()?.let { responseBody ->
                    val type = object : TypeToken<BatchMappingsResponse>() {}.type
                    val result: BatchMappingsResponse = gson.fromJson(responseBody, type)
                    result.mappings
                } ?: emptyList()
            } else emptyList()
        } catch (e: Exception) {
            emptyList()
        }
    }

    /**
     * Get detailed trace context for displaying in a tool window
     */
    fun getTraceContext(traceId: String): IDETraceContext? {
        val request = Request.Builder()
            .url("$baseUrl/api/public/ide/trace-context/$traceId")
            .header("Authorization", "Bearer $apiKey")
            .header("X-Project-ID", projectId)
            .get()
            .build()

        return try {
            val response = httpClient.newCall(request).execute()
            if (response.isSuccessful) {
                response.body?.string()?.let {
                    gson.fromJson(it, IDETraceContext::class.java)
                }
            } else null
        } catch (e: Exception) {
            null
        }
    }
}

data class FileTraceMapping(
    val filePath: String,
    val projectId: String,
    val annotations: List<LineAnnotation>,
    val summary: FileTraceSummary
)

data class LineAnnotation(
    val line: Int,
    val traceId: String,
    val traceName: String,
    val type: String,       // created, modified, read
    val agentName: String,
    val cost: Double,
    val latencyMs: Long,
    val timestamp: String,
    val confidence: Double
)

data class FileTraceSummary(
    val totalTraces: Int,
    val totalModifications: Int,
    val topAgents: List<String>,
    val totalCost: Double,
    val avgLatencyMs: Double,
    val lastModified: String?
)

data class IDETraceContext(
    val traceId: String,
    val traceName: String,
    val agentSession: String,
    val reasoning: String,
    val cost: Double,
    val latencyMs: Double,
    val fileChanges: List<IDEFileChange>
)

data class IDEFileChange(
    val path: String,
    val operation: String,
    val diffSummary: String
)

data class BatchMappingsResponse(
    val mappings: List<FileTraceMapping>
)
