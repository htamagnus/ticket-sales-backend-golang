package http

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"

	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/usecase"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/domain"
)

// EventsHandler handles HTTP requests for events.
type EventsHandler struct {
	ListEventsUseCase  *usecase.ListEventsUseCase
	GetEventUseCase    *usecase.GetEventUseCase
	CreateEventUseCase *usecase.CreateEventUseCase
	BuyTicketsUseCase  *usecase.BuyTicketsUseCase
	CreateSpotsUseCase *usecase.CreateSpotsUseCase
	ListSpotsUseCase   *usecase.ListSpotsUseCase
	Data               *domain.Data
}

func NewEventsHandler(
	listEventsUseCase *usecase.ListEventsUseCase,
	getEventUseCase *usecase.GetEventUseCase,
	createEventUseCase *usecase.CreateEventUseCase,
	buyTicketsUseCase *usecase.BuyTicketsUseCase,
	createSpotsUseCase *usecase.CreateSpotsUseCase,
	listSpotsUseCase *usecase.ListSpotsUseCase,
	data *domain.Data,
) *EventsHandler {
	return &EventsHandler{
		ListEventsUseCase:  listEventsUseCase,
		GetEventUseCase:    getEventUseCase,
		CreateEventUseCase: createEventUseCase,
		BuyTicketsUseCase:  buyTicketsUseCase,
		CreateSpotsUseCase: createSpotsUseCase,
		ListSpotsUseCase:   listSpotsUseCase,
		Data:               data,
	}
}

// ListEvents handles the request to list all events.
// @Summary List all events
// @Description Get all events with their details
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} usecase.ListEventsOutputDTO
// @Failure 500 {object} string
// @Router /events [get]

func (h *EventsHandler) ListEvents(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.Data.Events)
}

// GetEvent handles the request to get details of a specific event.
// @Summary Get event details
// @Description Get details of an event by ID
// @Tags Events
// @Accept json
// @Produce json
// @Param eventID path string true "Event ID"
// @Success 200 {object} usecase.GetEventOutputDTO
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /events/{eventID} [get]
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventID")

	var event *domain.Event
	for _, e := range h.Data.Events {
		if e.ID == eventID {
			event = &e
			break
		}
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// writeErrorResponse writes an error response in JSON format
func (h *EventsHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

// ErrorResponse represents the structure of an error response
type ErrorResponse struct {
	Message string `json:"message"`
}

// CreateSpotsRequest represents the input for creating spots.
type CreateSpotsRequest struct {
	NumberOfSpots int `json:"number_of_spots"`
}

// ListSpots lists spots for an event.
// @Summary List spots for an event
// @Description List all spots for a specific event
// @Tags Events
// @Accept json
// @Produce json
// @Param eventID path string true "Event ID"
// @Success 200 {object} usecase.ListSpotsOutputDTO
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /events/{eventID}/spots [get]
func (h *EventsHandler) ListSpots(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventID")

	var spots []domain.Spot
	for _, s := range h.Data.Spots {
		if s.EventID == eventID {
			spots = append(spots, s)
		}
	}

	if len(spots) == 0 {
		http.Error(w, "No spots found for this event", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spots)
}

func (h *EventsHandler) ReserveSpot(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventID")
	spotID := ps.ByName("spotID")

	var reservedSpot *domain.Spot
	for i, s := range h.Data.Spots {
		if s.EventID == eventID && s.ID == spotID {
			if s.Status == "reserved" {
				http.Error(w, "Spot already reserved", http.StatusConflict)
				return
			}
			h.Data.Spots[i].Status = "reserved"
			reservedSpot = &h.Data.Spots[i]
			break
		}
	}

	if reservedSpot == nil {
		http.Error(w, "Spot not found for this event", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reservedSpot)
}
