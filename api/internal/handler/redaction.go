package handler

import (
	"net/url"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

func redactEndpoint(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func redactWebhook(webhook *domain.Webhook) *domain.Webhook {
	redacted := *webhook
	redacted.URL = redactEndpoint(redacted.URL)
	redacted.Secret = ""
	redacted.Headers = nil
	return &redacted
}

func redactFederationPeer(peer *domain.OTelFederationPeer) *domain.OTelFederationPeer {
	redacted := *peer
	redacted.URL = redactEndpoint(redacted.URL)
	redacted.APIKey = ""
	return &redacted
}

func redactExportDestination(destination *domain.ExportDestination) *domain.ExportDestination {
	redacted := *destination
	redacted.Endpoint = redactEndpoint(redacted.Endpoint)
	redacted.Headers = nil
	return &redacted
}

func redactOTelExportDestination(destination *domain.OTelExportDestination) *domain.OTelExportDestination {
	redacted := *destination
	redacted.Endpoint = redactEndpoint(redacted.Endpoint)
	redacted.Headers = nil
	return &redacted
}

func redactOTelDestinationRef(destination domain.ExportDestinationRef) domain.ExportDestinationRef {
	destination.Endpoint = redactEndpoint(destination.Endpoint)
	destination.Headers = nil
	return destination
}

func redactOTelBridgeConfig(config *domain.OTelBridgeConfig) *domain.OTelBridgeConfig {
	redacted := *config
	redacted.ExportDestinations = append(
		[]domain.ExportDestinationRef(nil),
		config.ExportDestinations...,
	)
	for i := range redacted.ExportDestinations {
		redacted.ExportDestinations[i] = redactOTelDestinationRef(
			redacted.ExportDestinations[i],
		)
	}
	return &redacted
}
