package propagation

import (
	"testing"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
)

func TestMultiLevelDependencyChain(t *testing.T) {
	gatewayID := uuid.New()
	authID := uuid.New()
	orderID := uuid.New()
	inventoryID := uuid.New()
	paymentID := uuid.New()

	services := []domain.Service{
		{ID: gatewayID, Name: "api-gateway", DisplayName: "API Gateway"},
		{ID: authID, Name: "auth-service", DisplayName: "Auth Service"},
		{ID: orderID, Name: "order-service", DisplayName: "Order Service"},
		{ID: inventoryID, Name: "inventory-service", DisplayName: "Inventory Service"},
		{ID: paymentID, Name: "payment-service", DisplayName: "Payment Service"},
	}

	deps := []domain.Dependency{
		{SourceID: gatewayID, TargetID: authID, DependencyType: "sync"},
		{SourceID: gatewayID, TargetID: orderID, DependencyType: "sync"},
		{SourceID: orderID, TargetID: inventoryID, DependencyType: "sync"},
		{SourceID: orderID, TargetID: paymentID, DependencyType: "sync"},
	}

	graph := NewGraph(services, deps)

	// Test 1: Downstream dependencies from Gateway
	depsFromGateway := graph.findDependencies(gatewayID)
	if len(depsFromGateway) != 4 {
		t.Errorf("expected 4 downstream dependencies for Gateway, got %d", len(depsFromGateway))
	}

	// Test 2: Upstream callers of Payment Service
	callersOfPayment := graph.findCallers(paymentID)
	if len(callersOfPayment) != 2 { // Order and Gateway
		t.Errorf("expected 2 upstream callers for Payment, got %d", len(callersOfPayment))
	}

	// Test 3: Failure propagation when Payment has CRITICAL risk score
	riskScores := map[uuid.UUID]int{
		gatewayID:   45,
		authID:      10,
		orderID:     65,
		inventoryID: 15,
		paymentID:   92,
	}

	impact := graph.AnalyzeImpact(gatewayID, riskScores)

	// Root cause should be Payment Service
	if len(impact.PossibleRootCauses) == 0 {
		t.Fatalf("expected possible root causes, got none")
	}
	topRoot := impact.PossibleRootCauses[0]
	if topRoot.ServiceName != "Payment Service" {
		t.Errorf("expected Payment Service to be ranked #1 root cause, got %s", topRoot.ServiceName)
	}
	if topRoot.RiskScore != 92 {
		t.Errorf("expected top root risk score 92, got %d", topRoot.RiskScore)
	}

	// Affected services should include Order Service and API Gateway
	hasOrder := false
	hasGateway := false
	for _, s := range impact.AffectedServices {
		if s == "Order Service" {
			hasOrder = true
		}
		if s == "API Gateway" {
			hasGateway = true
		}
	}
	if !hasOrder || !hasGateway {
		t.Errorf("expected affected services to include Order and Gateway, got %+v", impact.AffectedServices)
	}

	// Propagation paths should trace Payment -> Order -> Gateway
	foundPath := false
	for _, p := range impact.PropagationPaths {
		if len(p) == 3 && p[0] == "Payment Service" && p[1] == "Order Service" && p[2] == "API Gateway" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Errorf("expected propagation path [Payment Service, Order Service, API Gateway], got %+v", impact.PropagationPaths)
	}
}
