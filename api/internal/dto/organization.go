package dto

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name string `json:"name" validate:"omitempty,min=1,max=100"`
}
