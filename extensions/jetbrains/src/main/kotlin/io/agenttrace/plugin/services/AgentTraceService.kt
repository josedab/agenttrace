package io.agenttrace.plugin.services

import com.intellij.openapi.Disposable
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.Service
import com.intellij.openapi.project.Project
import io.agenttrace.plugin.api.AgentTraceClient
import io.agenttrace.plugin.api.CostSummary
import io.agenttrace.plugin.api.Trace
import io.agenttrace.plugin.settings.AgentTraceSettings
import kotlinx.coroutines.*
import java.util.concurrent.CopyOnWriteArrayList

@Service(Service.Level.PROJECT)
class AgentTraceService(private val project: Project) : Disposable {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var refreshJob: Job? = null

    private val listeners = CopyOnWriteArrayList<AgentTraceListener>()

    val client: AgentTraceClient
        get() {
            val settings = AgentTraceSettings.getInstance()
            return AgentTraceClient(
                settings.apiUrl,
                settings.apiKey,
                settings.projectId
            )
        }

    var traces: List<Trace> = emptyList()
        private set

    var costSummary: CostSummary? = null
        private set

    var isLoading: Boolean = false
        private set

    var error: String? = null
        private set

    fun addListener(listener: AgentTraceListener) {
        listeners.add(listener)
    }

    fun removeListener(listener: AgentTraceListener) {
        listeners.remove(listener)
    }

    fun startAutoRefresh() {
        val settings = AgentTraceSettings.getInstance()
        if (!settings.autoRefresh) return

        refreshJob?.cancel()
        refreshJob = scope.launch {
            while (isActive) {
                refresh()
                delay(settings.refreshIntervalSeconds * 1000L)
            }
        }
    }

    fun stopAutoRefresh() {
        refreshJob?.cancel()
        refreshJob = null
    }

    fun refresh() {
        scope.launch {
            loadTraces()
            loadCostSummary()
        }
    }

    private suspend fun loadTraces() {
        val settings = AgentTraceSettings.getInstance()
        if (!settings.isConfigured()) {
            error = "Not configured - set API key and project ID"
            notifyListeners()
            return
        }

        isLoading = true
        error = null
        notifyListeners()

        try {
            val response = withContext(Dispatchers.IO) {
                client.listTraces(limit = settings.maxTraces)
            }

            if (response != null) {
                traces = response.data
                error = null
            } else {
                error = "Failed to load traces"
            }
        } catch (e: Exception) {
            error = "Error: ${e.message}"
        }

        isLoading = false
        notifyListeners()
    }

    private suspend fun loadCostSummary() {
        try {
            costSummary = withContext(Dispatchers.IO) {
                client.getCostSummary()
            }
            notifyCostUpdate()
        } catch (e: Exception) {
            // Silently fail for cost summary
        }
    }

    private fun notifyListeners() {
        ApplicationManager.getApplication().invokeLater {
            listeners.forEach { it.onTracesUpdated(traces, error) }
        }
    }

    private fun notifyCostUpdate() {
        ApplicationManager.getApplication().invokeLater {
            costSummary?.let { cost ->
                listeners.forEach { it.onCostUpdated(cost) }
            }
        }
    }

    override fun dispose() {
        scope.cancel()
        listeners.clear()
    }

    companion object {
        fun getInstance(project: Project): AgentTraceService {
            return project.getService(AgentTraceService::class.java)
        }
    }
}

interface AgentTraceListener {
    fun onTracesUpdated(traces: List<Trace>, error: String?) {}
    fun onCostUpdated(cost: CostSummary) {}
}
