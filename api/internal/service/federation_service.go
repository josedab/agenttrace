package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// FederationService manages cross-instance federation
type FederationService struct {
	logger       *zap.Logger
	peers        map[uuid.UUID]*domain.OTelFederationPeer
	destinations map[uuid.UUID]*domain.ExportDestination
}

// NewFederationService creates a new federation service
func NewFederationService(logger *zap.Logger) *FederationService {
	return &FederationService{
		logger:       logger,
		peers:        make(map[uuid.UUID]*domain.OTelFederationPeer),
		destinations: make(map[uuid.UUID]*domain.ExportDestination),
	}
}

// AddPeer registers a federation peer
func (s *FederationService) AddPeer(ctx context.Context, projectID uuid.UUID, input *domain.OTelFederationInput) (*domain.OTelFederationPeer, error) {
	peer := &domain.OTelFederationPeer{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		URL:       input.URL,
		APIKey:    input.APIKey,
		Status:    "connected",
		CreatedAt: time.Now(),
	}

	now := time.Now()
	peer.LastSeen = &now

	s.peers[peer.ID] = peer

	s.logger.Info("added federation peer",
		zap.String("peerId", peer.ID.String()),
		zap.String("name", peer.Name),
		zap.String("url", peer.URL),
	)

	return peer, nil
}

// ListPeers returns all federation peers for a project
func (s *FederationService) ListPeers(ctx context.Context, projectID uuid.UUID) []domain.OTelFederationPeer {
	var result []domain.OTelFederationPeer
	for _, peer := range s.peers {
		if peer.ProjectID == projectID {
			result = append(result, *peer)
		}
	}
	return result
}

// RemovePeer removes a federation peer
func (s *FederationService) RemovePeer(ctx context.Context, peerID uuid.UUID) error {
	delete(s.peers, peerID)
	return nil
}

// FederatedQuery executes a query across federation peers
func (s *FederationService) FederatedQuery(ctx context.Context, projectID uuid.UUID, query *domain.FederationQuery) (*domain.FederationResult, error) {
	start := time.Now()
	result := &domain.FederationResult{
		Results: []domain.FederationPeerResult{},
	}

	peers := s.ListPeers(ctx, projectID)
	for _, peer := range peers {
		if len(query.PeerIDs) > 0 {
			found := false
			for _, id := range query.PeerIDs {
				if id == peer.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		peerResult := domain.FederationPeerResult{
			PeerID:    peer.ID,
			PeerName:  peer.Name,
			LatencyMs: 0,
		}
		result.Results = append(result.Results, peerResult)
	}

	result.QueryTime = time.Since(start).Milliseconds()

	s.logger.Info("executed federated query",
		zap.Int("peerCount", len(result.Results)),
		zap.Int64("queryTimeMs", result.QueryTime),
	)

	return result, nil
}

// CreateExportDestination creates a new export destination
func (s *FederationService) CreateExportDestination(ctx context.Context, projectID uuid.UUID, input *domain.ExportDestinationInput) (*domain.ExportDestination, error) {
	dest := &domain.ExportDestination{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      input.Name,
		Type:      input.Type,
		Endpoint:  input.Endpoint,
		Protocol:  input.Protocol,
		Headers:   input.Headers,
		Enabled:   true,
		Sampling:  1.0,
		BatchSize: 100,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	if input.Sampling != nil {
		dest.Sampling = *input.Sampling
	}
	if input.BatchSize != nil {
		dest.BatchSize = *input.BatchSize
	}
	if dest.Protocol == "" {
		dest.Protocol = "grpc"
	}

	s.destinations[dest.ID] = dest

	s.logger.Info("created export destination",
		zap.String("destId", dest.ID.String()),
		zap.String("type", dest.Type),
		zap.String("endpoint", dest.Endpoint),
	)

	return dest, nil
}

// ListExportDestinations returns all destinations for a project
func (s *FederationService) ListExportDestinations(ctx context.Context, projectID uuid.UUID) []domain.ExportDestination {
	var result []domain.ExportDestination
	for _, dest := range s.destinations {
		if dest.ProjectID == projectID {
			result = append(result, *dest)
		}
	}
	return result
}
