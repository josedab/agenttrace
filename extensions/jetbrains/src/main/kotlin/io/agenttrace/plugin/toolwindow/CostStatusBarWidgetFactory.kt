package io.agenttrace.plugin.toolwindow

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.wm.StatusBar
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.openapi.wm.StatusBarWidgetFactory
import io.agenttrace.plugin.api.CostSummary
import io.agenttrace.plugin.services.AgentTraceListener
import io.agenttrace.plugin.services.AgentTraceService
import io.agenttrace.plugin.settings.AgentTraceSettings
import java.awt.event.MouseEvent
import javax.swing.SwingUtilities

class CostStatusBarWidgetFactory : StatusBarWidgetFactory {
    override fun getId(): String = "AgentTraceCostWidget"

    override fun getDisplayName(): String = "AgentTrace Cost"

    override fun isAvailable(project: Project): Boolean =
        AgentTraceSettings.getInstance().showStatusBarWidget

    override fun createWidget(project: Project): StatusBarWidget =
        CostStatusBarWidget(project)

    override fun disposeWidget(widget: StatusBarWidget) {
        Disposer.dispose(widget)
    }

    override fun canBeEnabledOn(statusBar: StatusBar): Boolean = true
}

class CostStatusBarWidget(private val project: Project) : StatusBarWidget, StatusBarWidget.TextPresentation, AgentTraceListener {
    private var statusBar: StatusBar? = null
    private var costText = "AgentTrace"
    private var tooltip = "Click to view cost summary"

    init {
        val service = AgentTraceService.getInstance(project)
        service.addListener(this)
    }

    override fun ID(): String = "AgentTraceCostWidget"

    override fun getPresentation(): StatusBarWidget.WidgetPresentation = this

    override fun install(statusBar: StatusBar) {
        this.statusBar = statusBar
    }

    override fun dispose() {
        statusBar = null
    }

    override fun getText(): String = costText

    override fun getTooltipText(): String = tooltip

    override fun getAlignment(): Float = 0f

    override fun getClickConsumer(): com.intellij.util.Consumer<MouseEvent> {
        return com.intellij.util.Consumer {
            val settings = AgentTraceSettings.getInstance()
            BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}/analytics/cost")
        }
    }

    override fun onCostUpdated(cost: CostSummary) {
        SwingUtilities.invokeLater {
            costText = "AgentTrace: \$${String.format("%.2f", cost.today)} today"
            tooltip = buildTooltip(cost)
            statusBar?.updateWidget(ID())
        }
    }

    private fun buildTooltip(cost: CostSummary): String {
        return buildString {
            appendLine("AgentTrace Cost Summary")
            appendLine()
            appendLine("Today: \$${String.format("%.4f", cost.today)}")
            appendLine("This Week: \$${String.format("%.4f", cost.thisWeek)}")
            appendLine("This Month: \$${String.format("%.4f", cost.thisMonth)}")

            cost.byModel?.let { models ->
                if (models.isNotEmpty()) {
                    appendLine()
                    appendLine("By Model:")
                    models.forEach { (model, modelCost) ->
                        appendLine("  $model: \$${String.format("%.4f", modelCost)}")
                    }
                }
            }

            appendLine()
            append("Click to view full breakdown")
        }
    }
}
