package io.agenttrace.plugin.settings

import com.intellij.openapi.options.Configurable
import com.intellij.ui.components.JBCheckBox
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBPasswordField
import com.intellij.ui.components.JBTextField
import com.intellij.util.ui.FormBuilder
import javax.swing.JComponent
import javax.swing.JPanel
import javax.swing.JSpinner
import javax.swing.SpinnerNumberModel

class AgentTraceConfigurable : Configurable {
    private var panel: JPanel? = null
    private val apiUrlField = JBTextField()
    private val dashboardUrlField = JBTextField()
    private val apiKeyField = JBPasswordField()
    private val projectIdField = JBTextField()
    private val autoRefreshCheckbox = JBCheckBox("Auto-refresh traces")
    private val refreshIntervalSpinner = JSpinner(SpinnerNumberModel(30, 5, 300, 5))
    private val showStatusBarCheckbox = JBCheckBox("Show cost in status bar")
    private val maxTracesSpinner = JSpinner(SpinnerNumberModel(50, 10, 200, 10))

    override fun getDisplayName(): String = "AgentTrace"

    override fun createComponent(): JComponent {
        panel = FormBuilder.createFormBuilder()
            .addLabeledComponent(JBLabel("API URL:"), apiUrlField, 1, false)
            .addLabeledComponent(JBLabel("Dashboard URL:"), dashboardUrlField, 1, false)
            .addLabeledComponent(JBLabel("API Key:"), apiKeyField, 1, false)
            .addLabeledComponent(JBLabel("Project ID:"), projectIdField, 1, false)
            .addSeparator()
            .addComponent(autoRefreshCheckbox)
            .addLabeledComponent(JBLabel("Refresh interval (seconds):"), refreshIntervalSpinner, 1, false)
            .addComponent(showStatusBarCheckbox)
            .addLabeledComponent(JBLabel("Max traces to display:"), maxTracesSpinner, 1, false)
            .addComponentFillVertically(JPanel(), 0)
            .panel

        reset()
        return panel!!
    }

    override fun isModified(): Boolean {
        val settings = AgentTraceSettings.getInstance()
        return apiUrlField.text != settings.apiUrl ||
                dashboardUrlField.text != settings.dashboardUrl ||
                String(apiKeyField.password) != settings.apiKey ||
                projectIdField.text != settings.projectId ||
                autoRefreshCheckbox.isSelected != settings.autoRefresh ||
                refreshIntervalSpinner.value as Int != settings.refreshIntervalSeconds ||
                showStatusBarCheckbox.isSelected != settings.showStatusBarWidget ||
                maxTracesSpinner.value as Int != settings.maxTraces
    }

    override fun apply() {
        val settings = AgentTraceSettings.getInstance()
        settings.apiUrl = apiUrlField.text
        settings.dashboardUrl = dashboardUrlField.text
        settings.apiKey = String(apiKeyField.password)
        settings.projectId = projectIdField.text
        settings.autoRefresh = autoRefreshCheckbox.isSelected
        settings.refreshIntervalSeconds = refreshIntervalSpinner.value as Int
        settings.showStatusBarWidget = showStatusBarCheckbox.isSelected
        settings.maxTraces = maxTracesSpinner.value as Int
    }

    override fun reset() {
        val settings = AgentTraceSettings.getInstance()
        apiUrlField.text = settings.apiUrl
        dashboardUrlField.text = settings.dashboardUrl
        apiKeyField.text = settings.apiKey
        projectIdField.text = settings.projectId
        autoRefreshCheckbox.isSelected = settings.autoRefresh
        refreshIntervalSpinner.value = settings.refreshIntervalSeconds
        showStatusBarCheckbox.isSelected = settings.showStatusBarWidget
        maxTracesSpinner.value = settings.maxTraces
    }

    override fun disposeUIResources() {
        panel = null
    }
}
