package service

import (
	"fmt"

	"github.com/agenttrace/agenttrace/api/internal/domain"
	apperrors "github.com/agenttrace/agenttrace/api/internal/pkg/errors"
)

// ExternalCapability identifies behavior that can leave the deployment.
type ExternalCapability string

// Outbound capabilities controlled by no-egress mode.
const (
	EgressWebhooks      ExternalCapability = "webhookDelivery"
	EgressGitHub        ExternalCapability = "githubReporting"
	EgressRemoteImport  ExternalCapability = "remoteImport"
	EgressRemoteExport  ExternalCapability = "remoteExport"
	EgressExternalModel ExternalCapability = "externalModelProviders"
	EgressOTelExport    ExternalCapability = "otelExport"
	EgressFederation    ExternalCapability = "federation"
	EgressWarehouseSync ExternalCapability = "warehouseSync"
	EgressSentry        ExternalCapability = "sentry"
)

// outboundCapabilities is the canonical list reported by the privacy endpoint.
// Every capability listed here must be enforced by a call to OutboundGuard.Require.
var outboundCapabilities = []ExternalCapability{
	EgressWebhooks,
	EgressGitHub,
	EgressRemoteImport,
	EgressRemoteExport,
	EgressExternalModel,
	EgressOTelExport,
	EgressFederation,
	EgressWarehouseSync,
	EgressSentry,
}

// OutboundGuard authorizes runtime-created outbound behavior.
// Services depend on this narrow interface instead of reading privacy config directly.
type OutboundGuard interface {
	Require(capability ExternalCapability) error
}

// EgressPolicy centrally enforces local/private mode.
type EgressPolicy struct {
	noEgress         bool
	redactionEnabled bool
}

// NewEgressPolicy creates a privacy policy.
func NewEgressPolicy(noEgress, redactionEnabled bool) *EgressPolicy {
	return &EgressPolicy{
		noEgress:         noEgress,
		redactionEnabled: redactionEnabled,
	}
}

// AllowAllOutbound returns a guard that permits every outbound capability.
func AllowAllOutbound() *EgressPolicy {
	return NewEgressPolicy(false, true)
}

// Require rejects outbound behavior while no-egress mode is active.
func (p *EgressPolicy) Require(capability ExternalCapability) error {
	if p != nil && p.noEgress {
		return apperrors.Unprocessable(
			fmt.Sprintf("%s is disabled by privacy no-egress mode", capability),
		)
	}
	return nil
}

// RequireOutbound applies an optional guard; an absent guard leaves behavior unchanged.
func RequireOutbound(guard OutboundGuard, capability ExternalCapability) error {
	if guard == nil {
		return nil
	}
	return guard.Require(capability)
}

// Capabilities returns an honest runtime capability report.
func (p *EgressPolicy) Capabilities() domain.PrivacyCapabilities {
	noEgress := p != nil && p.noEgress
	redactionEnabled := p == nil || p.redactionEnabled
	mode := "standard"
	if noEgress {
		mode = "local_private"
	}

	capabilities := make(map[string]domain.PrivacyCapability)
	for _, capability := range outboundCapabilities {
		item := domain.PrivacyCapability{Available: !noEgress}
		if noEgress {
			item.Reason = "Disabled by privacy no-egress mode"
		}
		capabilities[string(capability)] = item
	}
	capabilities["localTraceStorage"] = domain.PrivacyCapability{Available: true}
	capabilities["redactedShareLinks"] = domain.PrivacyCapability{
		Available: redactionEnabled,
		Reason: func() string {
			if redactionEnabled {
				return ""
			}
			return "Deterministic redaction is disabled"
		}(),
	}

	return domain.PrivacyCapabilities{
		Mode:             mode,
		NoEgress:         noEgress,
		RedactionEnabled: redactionEnabled,
		Capabilities:     capabilities,
	}
}
