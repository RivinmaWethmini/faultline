package handlers

import (
	"net/http"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ServiceHandler struct {
	svc      service.ServiceService
	riskRepo repository.RiskRepository
}

func NewServiceHandler(svc service.ServiceService, riskRepo repository.RiskRepository) *ServiceHandler {
	return &ServiceHandler{svc: svc, riskRepo: riskRepo}
}

type ServiceResponse struct {
	domain.Service
	RiskScore *domain.RiskScore `json:"risk_score,omitempty"`
}

// List returns all services with their latest risk scores.
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.svc.GetAll(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	var riskMap map[uuid.UUID]*domain.RiskScore
	if h.riskRepo != nil {
		riskScores, _ := h.riskRepo.GetAllLatest(r.Context())
		riskMap = make(map[uuid.UUID]*domain.RiskScore)
		for i := range riskScores {
			riskMap[riskScores[i].ServiceID] = &riskScores[i]
		}
	}

	result := make([]ServiceResponse, len(services))
	for i, s := range services {
		result[i] = ServiceResponse{
			Service: s,
		}
		if riskMap != nil {
			result[i].RiskScore = riskMap[s.ID]
		}
	}

	response.Success(w, result)
}

// Get returns a single service by ID.
func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid service ID format")
		return
	}

	s, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	res := ServiceResponse{Service: *s}
	if h.riskRepo != nil {
		rs, _ := h.riskRepo.GetLatestByService(r.Context(), s.ID)
		res.RiskScore = rs
	}

	response.Success(w, res)
}
