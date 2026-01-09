package io.agenttrace.plugin.toolwindow

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.actionSystem.*
import com.intellij.openapi.project.Project
import com.intellij.ui.JBColor
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBList
import com.intellij.ui.components.JBScrollPane
import com.intellij.util.ui.JBUI
import io.agenttrace.plugin.api.CostSummary
import io.agenttrace.plugin.api.Trace
import io.agenttrace.plugin.services.AgentTraceListener
import io.agenttrace.plugin.services.AgentTraceService
import io.agenttrace.plugin.settings.AgentTraceSettings
import java.awt.BorderLayout
import java.awt.Component
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.text.SimpleDateFormat
import java.util.*
import javax.swing.*

class TracesToolWindowPanel(private val project: Project) : JPanel(BorderLayout()), AgentTraceListener {
    private val service = AgentTraceService.getInstance(project)
    private val listModel = DefaultListModel<Trace>()
    private val traceList = JBList(listModel)
    private val statusLabel = JBLabel("Loading...")

    init {
        setupUI()
        setupListeners()
        service.addListener(this)
        service.startAutoRefresh()
    }

    private fun setupUI() {
        // Toolbar
        val toolbar = createToolbar()
        add(toolbar, BorderLayout.NORTH)

        // List
        traceList.cellRenderer = TraceListCellRenderer()
        traceList.selectionMode = ListSelectionModel.SINGLE_SELECTION

        traceList.addMouseListener(object : MouseAdapter() {
            override fun mouseClicked(e: MouseEvent) {
                if (e.clickCount == 2) {
                    val trace = traceList.selectedValue
                    if (trace != null) {
                        openTraceInBrowser(trace)
                    }
                }
            }
        })

        val scrollPane = JBScrollPane(traceList)
        add(scrollPane, BorderLayout.CENTER)

        // Status bar
        statusLabel.border = JBUI.Borders.empty(4, 8)
        add(statusLabel, BorderLayout.SOUTH)
    }

    private fun createToolbar(): JComponent {
        val actionGroup = DefaultActionGroup().apply {
            add(object : AnAction("Refresh", "Refresh traces", com.intellij.icons.AllIcons.Actions.Refresh) {
                override fun actionPerformed(e: AnActionEvent) {
                    service.refresh()
                }
            })
            addSeparator()
            add(object : AnAction("Open Dashboard", "Open in browser", com.intellij.icons.AllIcons.General.Web) {
                override fun actionPerformed(e: AnActionEvent) {
                    val settings = AgentTraceSettings.getInstance()
                    BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}")
                }
            })
        }

        val toolbar = ActionManager.getInstance().createActionToolbar("AgentTraceToolbar", actionGroup, true)
        toolbar.targetComponent = this
        return toolbar.component
    }

    private fun setupListeners() {
        traceList.addListSelectionListener { e ->
            if (!e.valueIsAdjusting) {
                val trace = traceList.selectedValue
                // Could show detail panel here
            }
        }
    }

    override fun onTracesUpdated(traces: List<Trace>, error: String?) {
        SwingUtilities.invokeLater {
            listModel.clear()
            if (error != null) {
                statusLabel.text = error
                statusLabel.foreground = JBColor.RED
            } else {
                traces.forEach { listModel.addElement(it) }
                statusLabel.text = "${traces.size} traces"
                statusLabel.foreground = JBColor.foreground()
            }
        }
    }

    override fun onCostUpdated(cost: CostSummary) {
        // Update status bar with cost info
        SwingUtilities.invokeLater {
            val count = listModel.size()
            statusLabel.text = "$count traces | Today: \$${String.format("%.4f", cost.today)}"
        }
    }

    private fun openTraceInBrowser(trace: Trace) {
        val settings = AgentTraceSettings.getInstance()
        BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}/traces/${trace.id}")
    }
}

class TraceListCellRenderer : ListCellRenderer<Trace> {
    private val dateFormat = SimpleDateFormat("HH:mm:ss")

    override fun getListCellRendererComponent(
        list: JList<out Trace>,
        value: Trace,
        index: Int,
        isSelected: Boolean,
        cellHasFocus: Boolean
    ): Component {
        val panel = JPanel(BorderLayout())
        panel.border = JBUI.Borders.empty(4, 8)

        // Status icon and name
        val statusIcon = when (value.status) {
            "completed" -> "\u2714" // Check mark
            "running" -> "\u25B6" // Play
            "error" -> "\u2716" // X
            else -> "\u2022" // Bullet
        }

        val statusColor = when (value.status) {
            "completed" -> JBColor.GREEN
            "running" -> JBColor.BLUE
            "error" -> JBColor.RED
            else -> JBColor.GRAY
        }

        val nameLabel = JBLabel("$statusIcon ${value.name ?: value.id.take(12)}")
        nameLabel.foreground = if (isSelected) list.selectionForeground else statusColor
        panel.add(nameLabel, BorderLayout.WEST)

        // Duration and cost
        val duration = value.duration?.let { "${it / 1000.0}s" } ?: "running"
        val cost = String.format("\$%.4f", value.totalCost)
        val detailLabel = JBLabel("$duration | $cost")
        detailLabel.foreground = if (isSelected) list.selectionForeground else JBColor.GRAY
        panel.add(detailLabel, BorderLayout.EAST)

        // Time
        try {
            val date = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss").parse(value.startTime.take(19))
            val timeLabel = JBLabel(dateFormat.format(date))
            timeLabel.foreground = if (isSelected) list.selectionForeground else JBColor.GRAY
            timeLabel.font = timeLabel.font.deriveFont(10f)
            panel.add(timeLabel, BorderLayout.SOUTH)
        } catch (e: Exception) {
            // Ignore parse errors
        }

        if (isSelected) {
            panel.background = list.selectionBackground
        } else {
            panel.background = list.background
        }

        return panel
    }
}
