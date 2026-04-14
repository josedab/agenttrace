package domain

// PrivacyCapabilities describes actual behavior rather than a marketing claim.
type PrivacyCapabilities struct {
	Mode             string                       `json:"mode"`
	NoEgress         bool                         `json:"noEgress"`
	RedactionEnabled bool                         `json:"redactionEnabled"`
	Capabilities     map[string]PrivacyCapability `json:"capabilities"`
}

// PrivacyCapability describes whether a potentially external capability is available.
type PrivacyCapability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
