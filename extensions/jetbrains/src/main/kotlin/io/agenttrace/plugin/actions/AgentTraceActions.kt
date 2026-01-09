package io.agenttrace.plugin.actions

import com.intellij.ide.BrowserUtil
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.options.ShowSettingsUtil
import com.intellij.openapi.ui.Messages
import git4idea.GitUtil
import git4idea.repo.GitRepository
import io.agenttrace.plugin.api.CheckpointInput
import io.agenttrace.plugin.api.GitLinkInput
import io.agenttrace.plugin.services.AgentTraceService
import io.agenttrace.plugin.settings.AgentTraceConfigurable
import io.agenttrace.plugin.settings.AgentTraceSettings

class RefreshTracesAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        AgentTraceService.getInstance(project).refresh()
    }
}

class OpenDashboardAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val settings = AgentTraceSettings.getInstance()
        BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}")
    }

    override fun update(e: AnActionEvent) {
        e.presentation.isEnabled = AgentTraceSettings.getInstance().isConfigured()
    }
}

class OpenSettingsAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        ShowSettingsUtil.getInstance().showSettingsDialog(project, AgentTraceConfigurable::class.java)
    }
}

class LinkGitCommitAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = AgentTraceService.getInstance(project)

        // Get current git repository
        val gitRepositories = GitUtil.getRepositories(project)
        if (gitRepositories.isEmpty()) {
            showNotification(project, "No Git repository found", NotificationType.WARNING)
            return
        }

        val repo = gitRepositories.first()
        val currentBranch = repo.currentBranch?.name
        val currentRevision = repo.currentRevision

        if (currentRevision == null) {
            showNotification(project, "No commits in repository", NotificationType.WARNING)
            return
        }

        // Get commit message
        val lastCommitMessage = getLastCommitMessage(repo)
        val author = getCommitAuthor(repo)

        // Ask for trace ID
        val traceId = Messages.showInputDialog(
            project,
            "Enter the Trace ID to link with commit ${currentRevision.take(7)}:",
            "Link Git Commit to Trace",
            null
        ) ?: return

        if (traceId.isBlank()) return

        // Create git link
        val input = GitLinkInput(
            traceId = traceId,
            commitSha = currentRevision,
            branch = currentBranch,
            commitMessage = lastCommitMessage,
            commitAuthor = author
        )

        val success = service.client.createGitLink(input)

        if (success) {
            showNotification(
                project,
                "Linked commit ${currentRevision.take(7)} to trace",
                NotificationType.INFORMATION
            )
            service.refresh()
        } else {
            showNotification(project, "Failed to create git link", NotificationType.ERROR)
        }
    }

    private fun getLastCommitMessage(repo: GitRepository): String? {
        return try {
            val git = git4idea.commands.Git.getInstance()
            val result = git.runCommand {
                it.addParameters("log", "-1", "--format=%s")
            }
            result.getOutputOrThrow().trim()
        } catch (e: Exception) {
            null
        }
    }

    private fun getCommitAuthor(repo: GitRepository): String? {
        return try {
            val git = git4idea.commands.Git.getInstance()
            val result = git.runCommand {
                it.addParameters("log", "-1", "--format=%an")
            }
            result.getOutputOrThrow().trim()
        } catch (e: Exception) {
            null
        }
    }

    override fun update(e: AnActionEvent) {
        val project = e.project
        e.presentation.isEnabled = project != null &&
                AgentTraceSettings.getInstance().isConfigured() &&
                GitUtil.getRepositories(project).isNotEmpty()
    }
}

class CreateCheckpointAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = AgentTraceService.getInstance(project)

        val traceId = Messages.showInputDialog(
            project,
            "Enter the Trace ID:",
            "Create Checkpoint",
            null
        ) ?: return

        if (traceId.isBlank()) return

        val name = Messages.showInputDialog(
            project,
            "Enter checkpoint name:",
            "Create Checkpoint",
            null
        ) ?: return

        if (name.isBlank()) return

        val description = Messages.showInputDialog(
            project,
            "Enter description (optional):",
            "Create Checkpoint",
            null
        )

        val input = CheckpointInput(
            traceId = traceId,
            name = name,
            description = description?.takeIf { it.isNotBlank() }
        )

        val success = service.client.createCheckpoint(input)

        if (success) {
            showNotification(project, "Checkpoint '$name' created", NotificationType.INFORMATION)
        } else {
            showNotification(project, "Failed to create checkpoint", NotificationType.ERROR)
        }
    }

    override fun update(e: AnActionEvent) {
        e.presentation.isEnabled = e.project != null && AgentTraceSettings.getInstance().isConfigured()
    }
}

class FilterByFileAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val editor = e.getData(CommonDataKeys.EDITOR) ?: return
        val virtualFile = e.getData(CommonDataKeys.VIRTUAL_FILE) ?: return

        val service = AgentTraceService.getInstance(project)
        val relativePath = project.basePath?.let {
            virtualFile.path.removePrefix(it).removePrefix("/")
        } ?: virtualFile.name

        val traces = service.client.searchTracesByFile(relativePath)

        if (traces.isEmpty()) {
            showNotification(project, "No traces found for $relativePath", NotificationType.INFORMATION)
            return
        }

        // Show list of traces
        val options = traces.map { "${it.name ?: it.id.take(12)} - ${it.status}" }.toTypedArray()
        val selection = Messages.showChooseDialog(
            project,
            "Select a trace for $relativePath:",
            "Traces for File",
            null,
            options,
            options.firstOrNull()
        )

        if (selection >= 0 && selection < traces.size) {
            val trace = traces[selection]
            val settings = AgentTraceSettings.getInstance()
            BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}/traces/${trace.id}")
        }
    }

    override fun update(e: AnActionEvent) {
        e.presentation.isEnabled = e.project != null &&
                e.getData(CommonDataKeys.VIRTUAL_FILE) != null &&
                AgentTraceSettings.getInstance().isConfigured()
    }
}

class OpenTraceInBrowserAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        // This would need access to selected trace from tool window
        // For now, open the traces list
        val settings = AgentTraceSettings.getInstance()
        BrowserUtil.browse("${settings.dashboardUrl}/projects/${settings.projectId}/traces")
    }
}

private fun showNotification(project: com.intellij.openapi.project.Project, message: String, type: NotificationType) {
    NotificationGroupManager.getInstance()
        .getNotificationGroup("AgentTrace Notifications")
        .createNotification("AgentTrace", message, type)
        .notify(project)
}
