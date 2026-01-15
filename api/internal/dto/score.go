package dto

// CreateScoreRequest represents the request to create a score
type CreateScoreRequest struct {
	TraceID       string   `json:"traceId" validate:"required"`
	ObservationID *string  `json:"observationId,omitempty"`
	Name          string   `json:"name" validate:"required"`
	Value         *float64 `json:"value,omitempty"`
	StringValue   *string  `json:"stringValue,omitempty"`
	Comment       *string  `json:"comment,omitempty"`
	DataType      *string  `json:"dataType,omitempty" validate:"omitempty,oneof=NUMERIC BOOLEAN CATEGORICAL"`
	Source        *string  `json:"source,omitempty" validate:"omitempty,oneof=API SDK ANNOTATION EVAL"`
}

// UpdateScoreRequest represents the request to update a score
type UpdateScoreRequest struct {
	Value       *float64 `json:"value,omitempty"`
	StringValue *string  `json:"stringValue,omitempty"`
	Comment     *string  `json:"comment,omitempty"`
}

// SubmitFeedbackRequest represents user feedback submission
type SubmitFeedbackRequest struct {
	TraceID       string   `json:"traceId" validate:"required"`
	ObservationID *string  `json:"observationId,omitempty"`
	Name          string   `json:"name" validate:"required"`
	Value         *float64 `json:"value,omitempty"`
	StringValue   *string  `json:"stringValue,omitempty"`
	Comment       *string  `json:"comment,omitempty"`
	UserID        *string  `json:"userId,omitempty"`
}

// BatchCreateScoresRequest represents a batch score creation request
type BatchCreateScoresRequest struct {
	Scores []CreateScoreRequest `json:"scores" validate:"required,min=1,max=100,dive"`
}
