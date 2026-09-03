package propagation

import (
	"fmt"
	"sort"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
)

// Graph represents the service dependency graph as adjacency lists.
type Graph struct {
	// downstream maps a service ID to the services it calls (dependencies)
	downstream map[uuid.UUID][]uuid.UUID
	// upstream maps a service ID to the services that call it (callers)
	upstream map[uuid.UUID][]uuid.UUID
	// nameByID maps service IDs to display names
	nameByID map[uuid.UUID]string
	// idByName maps service names and display names to IDs
	idByName map[string]uuid.UUID
	// serviceByID maps IDs to full Service structs
	serviceByID map[uuid.UUID]domain.Service
}

// NewGraph creates a dependency graph from services and their dependencies.
func NewGraph(services []domain.Service, deps []domain.Dependency) *Graph {
	g := &Graph{
		downstream:  make(map[uuid.UUID][]uuid.UUID),
		upstream:    make(map[uuid.UUID][]uuid.UUID),
		nameByID:    make(map[uuid.UUID]string),
		idByName:    make(map[string]uuid.UUID),
		serviceByID: make(map[uuid.UUID]domain.Service),
	}

	for _, s := range services {
		g.nameByID[s.ID] = s.DisplayName
		g.idByName[s.Name] = s.ID
		g.idByName[s.DisplayName] = s.ID
		g.serviceByID[s.ID] = s
	}

	for _, d := range deps {
		// source depends on target (source calls target)
		g.downstream[d.SourceID] = append(g.downstream[d.SourceID], d.TargetID)
		g.upstream[d.TargetID] = append(g.upstream[d.TargetID], d.SourceID)
	}

	return g
}

type rootCandidateInternal struct {
	id         uuid.UUID
	name       string
	riskScore  int
	distance   int
	confidence float64
	reason     string
}

// AnalyzeImpact performs full impact analysis for GET /api/services/:id/dependency-impact.
func (g *Graph) AnalyzeImpact(serviceID uuid.UUID, riskScores map[uuid.UUID]int) domain.DependencyImpactResult {
	svc := g.serviceByID[serviceID]

	// 1. Identify downstream services (services that this service depends on)
	downstreamIDs := g.findDependencies(serviceID)
	var downstreamNames []string
	for _, id := range downstreamIDs {
		if name, ok := g.nameByID[id]; ok {
			downstreamNames = append(downstreamNames, name)
		}
	}

	// 2. Identify upstream services (callers that depend on this service)
	upstreamIDs := g.findCallers(serviceID)
	var upstreamNames []string
	for _, id := range upstreamIDs {
		if name, ok := g.nameByID[id]; ok {
			upstreamNames = append(upstreamNames, name)
		}
	}

	// 3. Identify and rank possible root causes
	rankedCandidates := g.rankRootCausesInternal(serviceID, downstreamIDs, riskScores)

	// 4. Prime root cause ID is the top candidate
	primeRootID := serviceID
	if len(rankedCandidates) > 0 {
		primeRootID = rankedCandidates[0].id
	}

	// 5. Affected services: all callers of the prime root cause (who are impacted by its degradation)
	affectedIDs := g.findCallers(primeRootID)
	var affectedNames []string
	for _, id := range affectedIDs {
		if name, ok := g.nameByID[id]; ok {
			affectedNames = append(affectedNames, name)
		}
	}

	// 6. Calculate potential propagation paths from root cause up through callers
	paths := g.calculatePropagationPaths(primeRootID)

	var publicCandidates []domain.RootCauseCandidate
	for _, c := range rankedCandidates {
		publicCandidates = append(publicCandidates, domain.RootCauseCandidate{
			ServiceName: c.name,
			RiskScore:   c.riskScore,
			Distance:    c.distance,
			Confidence:  c.confidence,
			Reason:      c.reason,
		})
	}

	return domain.DependencyImpactResult{
		Service:            svc,
		UpstreamServices:   upstreamNames,
		DownstreamServices: downstreamNames,
		PossibleRootCauses: publicCandidates,
		AffectedServices:   affectedNames,
		PropagationPaths:   paths,
	}
}

// Analyze returns simple PropagationResult for backward compatibility and incident generation.
func (g *Graph) Analyze(serviceID uuid.UUID, riskScores map[uuid.UUID]int) domain.PropagationResult {
	impact := g.AnalyzeImpact(serviceID, riskScores)

	var rootCauseName string
	if len(impact.PossibleRootCauses) > 0 {
		rootCauseName = impact.PossibleRootCauses[0].ServiceName
	} else {
		rootCauseName = g.nameByID[serviceID]
	}

	var primaryPath []string
	if len(impact.PropagationPaths) > 0 {
		primaryPath = impact.PropagationPaths[0]
	} else {
		primaryPath = []string{g.nameByID[serviceID]}
	}

	return domain.PropagationResult{
		RootCause:        rootCauseName,
		PropagationPath:  primaryPath,
		AffectedServices: impact.AffectedServices,
		UpstreamImpact:   impact.UpstreamServices,
	}
}

// rankRootCausesInternal ranks possible failure origins based on anomaly scores and topological distance.
func (g *Graph) rankRootCausesInternal(start uuid.UUID, depIDs []uuid.UUID, riskScores map[uuid.UUID]int) []rootCandidateInternal {
	type candidateSeed struct {
		id       uuid.UUID
		distance int
	}

	candidates := []candidateSeed{{id: start, distance: 0}}
	distances := g.computeDepDistances(start)
	for _, depID := range depIDs {
		dist := distances[depID]
		candidates = append(candidates, candidateSeed{id: depID, distance: dist})
	}

	var results []rootCandidateInternal
	for _, c := range candidates {
		score := riskScores[c.id]
		name := g.nameByID[c.id]

		confidence := float64(score) / 100.0
		if c.distance > 0 {
			confidence *= 1.0 / (1.0 + 0.15*float64(c.distance))
		}

		var reason string
		switch {
		case c.distance == 0:
			reason = fmt.Sprintf("Target service has risk score %d (%s)", score, domain.ClassifyRisk(score))
		case c.distance == 1:
			reason = fmt.Sprintf("Direct downstream dependency with risk score %d (%s)", score, domain.ClassifyRisk(score))
		default:
			reason = fmt.Sprintf("Transitive downstream dependency %d hops away with risk score %d (%s)", c.distance, score, domain.ClassifyRisk(score))
		}

		results = append(results, rootCandidateInternal{
			id:         c.id,
			name:       name,
			riskScore:  score,
			distance:   c.distance,
			confidence: confidence,
			reason:     reason,
		})
	}

	// Rank candidates by risk score descending, then distance ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].riskScore != results[j].riskScore {
			return results[i].riskScore > results[j].riskScore
		}
		return results[i].distance < results[j].distance
	})

	return results
}

// findDependencies (BFS) finds all services that 'start' calls directly or transitively.
func (g *Graph) findDependencies(start uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	visited[start] = true
	queue := []uuid.UUID{start}
	var result []uuid.UUID

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, depID := range g.downstream[curr] {
			if !visited[depID] {
				visited[depID] = true
				result = append(result, depID)
				queue = append(queue, depID)
			}
		}
	}
	return result
}

// findCallers (BFS) finds all services that call 'start' directly or transitively.
func (g *Graph) findCallers(start uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	visited[start] = true
	queue := []uuid.UUID{start}
	var result []uuid.UUID

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, callerID := range g.upstream[curr] {
			if !visited[callerID] {
				visited[callerID] = true
				result = append(result, callerID)
				queue = append(queue, callerID)
			}
		}
	}
	return result
}

// computeDepDistances computes BFS shortest distances to all downstream dependencies.
func (g *Graph) computeDepDistances(start uuid.UUID) map[uuid.UUID]int {
	dist := make(map[uuid.UUID]int)
	dist[start] = 0
	queue := []uuid.UUID{start}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		d := dist[curr]

		for _, depID := range g.downstream[curr] {
			if _, seen := dist[depID]; !seen {
				dist[depID] = d + 1
				queue = append(queue, depID)
			}
		}
	}
	return dist
}

// calculatePropagationPaths returns ordered chains from root up to leaf callers (e.g. [Payment -> Order -> Gateway]).
func (g *Graph) calculatePropagationPaths(root uuid.UUID) [][]string {
	var allPaths [][]string

	var dfs func(curr uuid.UUID, currentPath []string)
	dfs = func(curr uuid.UUID, currentPath []string) {
		name := g.nameByID[curr]
		path := append(currentPath, name)

		callers := g.upstream[curr]
		if len(callers) == 0 {
			allPaths = append(allPaths, path)
			return
		}

		for _, caller := range callers {
			dfs(caller, path)
		}
	}

	dfs(root, []string{})
	if len(allPaths) == 0 && g.nameByID[root] != "" {
		allPaths = append(allPaths, []string{g.nameByID[root]})
	}

	return allPaths
}
