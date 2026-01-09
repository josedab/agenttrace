package io.agenttrace.plugin.settings

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.util.xmlb.XmlSerializerUtil

@State(
    name = "AgentTraceSettings",
    storages = [Storage("AgentTraceSettings.xml")]
)
class AgentTraceSettings : PersistentStateComponent<AgentTraceSettings> {
    var apiUrl: String = "https://api.agenttrace.io"
    var dashboardUrl: String = "https://app.agenttrace.io"
    var apiKey: String = ""
    var projectId: String = ""
    var autoRefresh: Boolean = true
    var refreshIntervalSeconds: Int = 30
    var showStatusBarWidget: Boolean = true
    var maxTraces: Int = 50

    override fun getState(): AgentTraceSettings = this

    override fun loadState(state: AgentTraceSettings) {
        XmlSerializerUtil.copyBean(state, this)
    }

    fun isConfigured(): Boolean = apiKey.isNotBlank() && projectId.isNotBlank()

    companion object {
        fun getInstance(): AgentTraceSettings {
            return ApplicationManager.getApplication().getService(AgentTraceSettings::class.java)
        }
    }
}
