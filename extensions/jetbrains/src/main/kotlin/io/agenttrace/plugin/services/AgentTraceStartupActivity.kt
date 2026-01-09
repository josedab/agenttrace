package io.agenttrace.plugin.services

import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.ProjectActivity
import io.agenttrace.plugin.settings.AgentTraceSettings

class AgentTraceStartupActivity : ProjectActivity {
    override suspend fun execute(project: Project) {
        val settings = AgentTraceSettings.getInstance()

        if (!settings.isConfigured()) {
            NotificationGroupManager.getInstance()
                .getNotificationGroup("AgentTrace Notifications")
                .createNotification(
                    "AgentTrace",
                    "Configure AgentTrace to start tracking traces. Go to Settings → Tools → AgentTrace.",
                    NotificationType.INFORMATION
                )
                .notify(project)
            return
        }

        // Start the service
        val service = AgentTraceService.getInstance(project)
        service.startAutoRefresh()
    }
}
