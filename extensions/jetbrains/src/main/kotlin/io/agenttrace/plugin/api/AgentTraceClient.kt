package io.agenttrace.plugin.api

import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

data class Trace(
    val id: String,
    val name: String?,
    val projectId: String,
    val sessionId: String?,
    val status: String,
    val startTime: String,
    val endTime: String?,
    val duration: Long?,
    val inputTokens: Int,
    val outputTokens: Int,
    val totalCost: Double,
    val level: String?,
    val metadata: Map<String, Any>?,
    val tags: List<String>?,
    val gitCommitSha: String?,
    val gitBranch: String?
)

data class Observation(
    val id: String,
    val traceId: String,
    val parentObservationId: String?,
    val name: String,
    val type: String,
    val startTime: String,
    val endTime: String?,
    val status: String,
    val model: String?,
    val inputTokens: Int?,
    val outputTokens: Int?,
    val cost: Double?,
    val input: Any?,
    val output: Any?,
    val metadata: Map<String, Any>?
)

data class CostSummary(
    val today: Double,
    val thisWeek: Double,
    val thisMonth: Double,
    val byModel: Map<String, Double>?
)

data class TracesResponse(
    val data: List<Trace>,
    val totalCount: Int,
    val hasMore: Boolean
)

data class GitLinkInput(
    val traceId: String,
    val commitSha: String,
    val branch: String?,
    val commitMessage: String?,
    val commitAuthor: String?
)

data class CheckpointInput(
    val traceId: String,
    val name: String,
    val description: String?,
    val type: String = "manual"
)

class AgentTraceClient(
    private var apiUrl: String,
    private var apiKey: String,
    private var projectId: String
) {
    private val gson = Gson()
    private val client = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .build()

    private val jsonMediaType = "application/json".toMediaType()

    fun updateConfig(apiUrl: String, apiKey: String, projectId: String) {
        this.apiUrl = apiUrl
        this.apiKey = apiKey
        this.projectId = projectId
    }

    fun isConfigured(): Boolean = apiKey.isNotBlank() && projectId.isNotBlank()

    fun getProjectId(): String = projectId

    fun getDashboardUrl(): String = apiUrl.replace("/api", "").replace("api.", "app.")

    fun listTraces(
        limit: Int = 50,
        offset: Int = 0,
        status: String? = null,
        search: String? = null,
        sessionId: String? = null
    ): TracesResponse? {
        val params = mutableListOf("limit=$limit", "offset=$offset")
        status?.let { params.add("status=$it") }
        search?.let { params.add("search=$it") }
        sessionId?.let { params.add("sessionId=$it") }

        val request = Request.Builder()
            .url("$apiUrl/v1/traces?${params.joinToString("&")}")
            .header("Authorization", "Bearer $apiKey")
            .get()
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    val body = response.body?.string()
                    gson.fromJson(body, TracesResponse::class.java)
                } else {
                    null
                }
            }
        } catch (e: Exception) {
            null
        }
    }

    fun getTrace(traceId: String): Trace? {
        val request = Request.Builder()
            .url("$apiUrl/v1/traces/$traceId")
            .header("Authorization", "Bearer $apiKey")
            .get()
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    val body = response.body?.string()
                    gson.fromJson(body, Trace::class.java)
                } else {
                    null
                }
            }
        } catch (e: Exception) {
            null
        }
    }

    fun getTraceObservations(traceId: String): List<Observation> {
        val request = Request.Builder()
            .url("$apiUrl/v1/traces/$traceId/observations")
            .header("Authorization", "Bearer $apiKey")
            .get()
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    val body = response.body?.string()
                    val type = object : TypeToken<Map<String, Any>>() {}.type
                    val result: Map<String, Any> = gson.fromJson(body, type)
                    val data = result["data"] as? List<*>
                    data?.mapNotNull {
                        try {
                            gson.fromJson(gson.toJson(it), Observation::class.java)
                        } catch (e: Exception) {
                            null
                        }
                    } ?: emptyList()
                } else {
                    emptyList()
                }
            }
        } catch (e: Exception) {
            emptyList()
        }
    }

    fun getCostSummary(): CostSummary? {
        val request = Request.Builder()
            .url("$apiUrl/v1/metrics/costs")
            .header("Authorization", "Bearer $apiKey")
            .get()
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    val body = response.body?.string()
                    gson.fromJson(body, CostSummary::class.java)
                } else {
                    null
                }
            }
        } catch (e: Exception) {
            null
        }
    }

    fun createGitLink(input: GitLinkInput): Boolean {
        val json = gson.toJson(input)
        val request = Request.Builder()
            .url("$apiUrl/v1/git-links")
            .header("Authorization", "Bearer $apiKey")
            .post(json.toRequestBody(jsonMediaType))
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                response.isSuccessful
            }
        } catch (e: Exception) {
            false
        }
    }

    fun createCheckpoint(input: CheckpointInput): Boolean {
        val json = gson.toJson(input)
        val request = Request.Builder()
            .url("$apiUrl/v1/checkpoints")
            .header("Authorization", "Bearer $apiKey")
            .post(json.toRequestBody(jsonMediaType))
            .build()

        return try {
            client.newCall(request).execute().use { response ->
                response.isSuccessful
            }
        } catch (e: Exception) {
            false
        }
    }

    fun searchTracesByFile(filePath: String): List<Trace> {
        val response = listTraces(limit = 20, search = filePath)
        return response?.data ?: emptyList()
    }
}
