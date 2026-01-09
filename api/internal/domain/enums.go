package domain

// Level represents the severity level
type Level string

const (
	LevelDebug   Level = "DEBUG"
	LevelDefault Level = "DEFAULT"
	LevelWarning Level = "WARNING"
	LevelError   Level = "ERROR"
)

// IsValid checks if the level is valid
func (l Level) IsValid() bool {
	switch l {
	case LevelDebug, LevelDefault, LevelWarning, LevelError:
		return true
	}
	return false
}

// ObservationType represents the type of observation
type ObservationType string

const (
	ObservationTypeSpan       ObservationType = "SPAN"
	ObservationTypeGeneration ObservationType = "GENERATION"
	ObservationTypeEvent      ObservationType = "EVENT"
)

// IsValid checks if the observation type is valid
func (t ObservationType) IsValid() bool {
	switch t {
	case ObservationTypeSpan, ObservationTypeGeneration, ObservationTypeEvent:
		return true
	}
	return false
}

// ScoreSource represents the source of a score
type ScoreSource string

const (
	ScoreSourceAPI        ScoreSource = "API"
	ScoreSourceEval       ScoreSource = "EVAL"
	ScoreSourceAnnotation ScoreSource = "ANNOTATION"
)

// IsValid checks if the score source is valid
func (s ScoreSource) IsValid() bool {
	switch s {
	case ScoreSourceAPI, ScoreSourceEval, ScoreSourceAnnotation:
		return true
	}
	return false
}

// ScoreDataType represents the data type of a score
type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
)

// IsValid checks if the score data type is valid
func (t ScoreDataType) IsValid() bool {
	switch t {
	case ScoreDataTypeNumeric, ScoreDataTypeBoolean, ScoreDataTypeCategorical:
		return true
	}
	return false
}

// PromptType represents the type of prompt
type PromptType string

const (
	PromptTypeText PromptType = "text"
	PromptTypeChat PromptType = "chat"
)

// IsValid checks if the prompt type is valid
func (t PromptType) IsValid() bool {
	switch t {
	case PromptTypeText, PromptTypeChat:
		return true
	}
	return false
}

// EvaluatorType represents the type of evaluator
type EvaluatorType string

const (
	EvaluatorTypeLLM        EvaluatorType = "llm"
	EvaluatorTypeLLMAsJudge EvaluatorType = "llm_as_judge"
	EvaluatorTypeRule       EvaluatorType = "rule"
	EvaluatorTypeCustom     EvaluatorType = "custom"
)

// IsValid checks if the evaluator type is valid
func (t EvaluatorType) IsValid() bool {
	switch t {
	case EvaluatorTypeLLM, EvaluatorTypeLLMAsJudge, EvaluatorTypeRule, EvaluatorTypeCustom:
		return true
	}
	return false
}

// DatasetItemStatus represents the status of a dataset item
type DatasetItemStatus string

const (
	DatasetItemStatusActive   DatasetItemStatus = "active"
	DatasetItemStatusArchived DatasetItemStatus = "archived"
)

// IsValid checks if the dataset item status is valid
func (s DatasetItemStatus) IsValid() bool {
	switch s {
	case DatasetItemStatusActive, DatasetItemStatusArchived:
		return true
	}
	return false
}

// OrgRole represents the role of a user in an organization
type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
	OrgRoleViewer OrgRole = "viewer"
)

// IsValid checks if the org role is valid
func (r OrgRole) IsValid() bool {
	switch r {
	case OrgRoleOwner, OrgRoleAdmin, OrgRoleMember, OrgRoleViewer:
		return true
	}
	return false
}

// CanManageMembers checks if the role can manage members
func (r OrgRole) CanManageMembers() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanManageProject checks if the role can manage projects
func (r OrgRole) CanManageProject() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanWrite checks if the role can write data
func (r OrgRole) CanWrite() bool {
	return r != OrgRoleViewer
}

// CanRead checks if the role can read data
func (r OrgRole) CanRead() bool {
	return true
}

// ExportFormat represents the format for data export
type ExportFormat string

const (
	ExportFormatJSON          ExportFormat = "json"
	ExportFormatCSV           ExportFormat = "csv"
	ExportFormatOpenAIFinetune ExportFormat = "openai_finetune"
)

// IsValid checks if the export format is valid
func (f ExportFormat) IsValid() bool {
	switch f {
	case ExportFormatJSON, ExportFormatCSV, ExportFormatOpenAIFinetune:
		return true
	}
	return false
}

// DestinationType represents the destination type for exports
type DestinationType string

const (
	DestinationTypeS3        DestinationType = "s3"
	DestinationTypeGCS       DestinationType = "gcs"
	DestinationTypeAzureBlob DestinationType = "azure_blob"
)

// IsValid checks if the destination type is valid
func (t DestinationType) IsValid() bool {
	switch t {
	case DestinationTypeS3, DestinationTypeGCS, DestinationTypeAzureBlob:
		return true
	}
	return false
}

// JobStatus represents the status of a background job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// IsValid checks if the job status is valid
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusPending, JobStatusRunning, JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	}
	return false
}

// IsTerminal checks if the job status is terminal
func (s JobStatus) IsTerminal() bool {
	return s == JobStatusCompleted || s == JobStatusFailed || s == JobStatusCancelled
}

// SortOrder represents the sort order for queries
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// IsValid checks if the sort order is valid
func (o SortOrder) IsValid() bool {
	switch o {
	case SortOrderAsc, SortOrderDesc:
		return true
	}
	return false
}

// CheckpointType represents the type of checkpoint
type CheckpointType string

const (
	CheckpointTypeManual   CheckpointType = "manual"
	CheckpointTypeAuto     CheckpointType = "auto"
	CheckpointTypePreEdit  CheckpointType = "pre_edit"
	CheckpointTypePostEdit CheckpointType = "post_edit"
	CheckpointTypeRollback CheckpointType = "rollback"
)

// IsValid checks if the checkpoint type is valid
func (t CheckpointType) IsValid() bool {
	switch t {
	case CheckpointTypeManual, CheckpointTypeAuto, CheckpointTypePreEdit, CheckpointTypePostEdit, CheckpointTypeRollback:
		return true
	}
	return false
}

// GitLinkType represents the type of git link
type GitLinkType string

const (
	GitLinkTypeCurrent    GitLinkType = "current"
	GitLinkTypeStart      GitLinkType = "start"
	GitLinkTypeEnd        GitLinkType = "end"
	GitLinkTypeReferenced GitLinkType = "referenced"
)

// IsValid checks if the git link type is valid
func (t GitLinkType) IsValid() bool {
	switch t {
	case GitLinkTypeCurrent, GitLinkTypeStart, GitLinkTypeEnd, GitLinkTypeReferenced:
		return true
	}
	return false
}

// FileOperationType represents the type of file operation
type FileOperationType string

const (
	FileOperationCreate FileOperationType = "create"
	FileOperationRead   FileOperationType = "read"
	FileOperationUpdate FileOperationType = "update"
	FileOperationDelete FileOperationType = "delete"
	FileOperationRename FileOperationType = "rename"
	FileOperationMove   FileOperationType = "move"
	FileOperationCopy   FileOperationType = "copy"
)

// IsValid checks if the file operation type is valid
func (t FileOperationType) IsValid() bool {
	switch t {
	case FileOperationCreate, FileOperationRead, FileOperationUpdate, FileOperationDelete, FileOperationRename, FileOperationMove, FileOperationCopy:
		return true
	}
	return false
}

// CIProvider represents the CI/CD provider
type CIProvider string

const (
	CIProviderGitHubActions CIProvider = "github_actions"
	CIProviderGitLabCI      CIProvider = "gitlab_ci"
	CIProviderJenkins       CIProvider = "jenkins"
	CIProviderCircleCI      CIProvider = "circleci"
	CIProviderAzureDevOps   CIProvider = "azure_devops"
	CIProviderBitbucket     CIProvider = "bitbucket"
	CIProviderOther         CIProvider = "other"
)

// IsValid checks if the CI provider is valid
func (p CIProvider) IsValid() bool {
	switch p {
	case CIProviderGitHubActions, CIProviderGitLabCI, CIProviderJenkins, CIProviderCircleCI, CIProviderAzureDevOps, CIProviderBitbucket, CIProviderOther:
		return true
	}
	return false
}

// CIRunStatus represents the status of a CI run
type CIRunStatus string

const (
	CIRunStatusPending   CIRunStatus = "pending"
	CIRunStatusRunning   CIRunStatus = "running"
	CIRunStatusSuccess   CIRunStatus = "success"
	CIRunStatusFailure   CIRunStatus = "failure"
	CIRunStatusCancelled CIRunStatus = "cancelled"
	CIRunStatusSkipped   CIRunStatus = "skipped"
)

// IsValid checks if the CI run status is valid
func (s CIRunStatus) IsValid() bool {
	switch s {
	case CIRunStatusPending, CIRunStatusRunning, CIRunStatusSuccess, CIRunStatusFailure, CIRunStatusCancelled, CIRunStatusSkipped:
		return true
	}
	return false
}

// IsTerminal checks if the CI run status is terminal
func (s CIRunStatus) IsTerminal() bool {
	return s == CIRunStatusSuccess || s == CIRunStatusFailure || s == CIRunStatusCancelled || s == CIRunStatusSkipped
}
