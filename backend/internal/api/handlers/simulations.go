package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SimulationHandler struct {
	simSvc service.SimulationService
}

func NewSimulationHandler(simSvc service.SimulationService) *SimulationHandler {
	return &SimulationHandler{simSvc: simSvc}
}

type CreateSimulationRequest struct {
	Scenario string `json:"scenario"`
}

type ProgrammaticDegradationRequest struct {
	TargetService string             `json:"target_service"`
	Modifiers     map[string]float64 `json:"modifiers"`
	DurationSec   int                `json:"duration_seconds"`
}

// Create triggers a predefined failure simulation scenario.
func (h *SimulationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	sim, err := h.simSvc.StartSimulation(r.Context(), req.Scenario)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Created(w, sim)
}

// Stop terminates an active simulation and clears its active degradation modifiers immediately.
func (h *SimulationHandler) Stop(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid simulation ID format")
		return
	}

	sim, err := h.simSvc.StopSimulation(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, sim)
}

// Degrade injects custom programmatic degradation modifiers.
func (h *SimulationHandler) Degrade(w http.ResponseWriter, r *http.Request) {
	var req ProgrammaticDegradationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	duration := time.Duration(req.DurationSec) * time.Second
	if req.DurationSec <= 0 {
		duration = 60 * time.Second
	}

	err := h.simSvc.DegradeService(r.Context(), req.TargetService, req.Modifiers, duration)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"message":        "programmatic degradation applied",
		"target_service": req.TargetService,
		"duration_sec":   duration.Seconds(),
	})
}

// List returns all simulations.
func (h *SimulationHandler) List(w http.ResponseWriter, r *http.Request) {
	sims, err := h.simSvc.GetAll(r.Context(), 50)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, sims)
}

// Scenarios returns available failure scenarios.
func (h *SimulationHandler) Scenarios(w http.ResponseWriter, r *http.Request) {
	response.Success(w, h.simSvc.GetScenarios())
}
