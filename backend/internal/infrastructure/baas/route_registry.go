package baas

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// RouteRegistry maps versioned paths to their generated HTTP handlers.
type RouteRegistry struct {
	mu           sync.RWMutex
	registry     map[string]http.HandlerFunc // key: "{version}/{path}"
	endpointKey  map[string]uuid.UUID        // key -> endpoint ID
	generator    RESTGenerator
	endpointRepo repository.EndpointRepository
	logger       repository.Logger
}

// NewRouteRegistry creates a new route registry.
func NewRouteRegistry(gen RESTGenerator, epRepo repository.EndpointRepository, logger repository.Logger) *RouteRegistry {
	return &RouteRegistry{
		registry:     make(map[string]http.HandlerFunc),
		endpointKey:  make(map[string]uuid.UUID),
		generator:    gen,
		endpointRepo: epRepo,
		logger:       logger,
	}
}

// RegisterEndpoint registers a handler for an endpoint.
func (rr *RouteRegistry) RegisterEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	handler, err := rr.generator.GenerateHandler(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("generate handler for endpoint %s: %w", endpoint.ID, err)
	}

	key := routeKey(endpoint.Version, endpoint.Path)
	rr.mu.Lock()
	rr.registry[key] = handler
	rr.endpointKey[key] = endpoint.ID
	rr.mu.Unlock()

	rr.logger.Info(ctx, "endpoint registered",
		"key", key,
		"endpoint_id", endpoint.ID,
		"db_type", endpoint.DBType,
	)
	return nil
}

// UnregisterEndpoint removes a handler registration by endpoint ID.
func (rr *RouteRegistry) UnregisterEndpoint(ctx context.Context, endpointID uuid.UUID) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for key, id := range rr.endpointKey {
		if id == endpointID {
			delete(rr.registry, key)
			delete(rr.endpointKey, key)
			rr.logger.Info(ctx, "endpoint unregistered", "key", key, "endpoint_id", id)
			return
		}
	}
}

// GetHandler returns the handler for a versioned path.
func (rr *RouteRegistry) GetHandler(version, path string) (http.HandlerFunc, bool) {
	key := routeKey(version, path)
	rr.mu.RLock()
	handler, ok := rr.registry[key]
	rr.mu.RUnlock()
	return handler, ok
}

// GetEndpointID returns the endpoint ID for a versioned path if registered.
func (rr *RouteRegistry) GetEndpointID(version, path string) (uuid.UUID, bool) {
	key := routeKey(version, path)
	rr.mu.RLock()
	id, ok := rr.endpointKey[key]
	rr.mu.RUnlock()
	return id, ok
}

// Endpoints returns a snapshot of all registered routes.
func (rr *RouteRegistry) Endpoints() map[string]uuid.UUID {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	result := make(map[string]uuid.UUID, len(rr.endpointKey))
	for k, v := range rr.endpointKey {
		result[k] = v
	}
	return result
}

// RefreshAll re-registers all active endpoints from the database.
func (rr *RouteRegistry) RefreshAll(ctx context.Context) error {
	endpoints, err := rr.endpointRepo.ListAllActive(ctx)
	if err != nil {
		return fmt.Errorf("refresh route registry: %w", err)
	}

	// Clear existing registrations
	rr.mu.Lock()
	rr.registry = make(map[string]http.HandlerFunc)
	rr.endpointKey = make(map[string]uuid.UUID)
	rr.mu.Unlock()

	// Register all active endpoints
	for _, ep := range endpoints {
		if err := rr.RegisterEndpoint(ctx, ep); err != nil {
			rr.logger.Error(ctx, "failed to register endpoint",
				"endpoint_id", ep.ID,
				"error", err,
			)
		}
	}

	rr.logger.Info(ctx, "route registry refreshed", "total", len(endpoints))
	return nil
}

// routeKey generates a consistent route key from version and path.
func routeKey(version, path string) string {
	return version + path
}
