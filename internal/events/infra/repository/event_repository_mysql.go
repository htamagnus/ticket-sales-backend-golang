package repository

import (
	"database/sql"
	"errors"
	"time"
	"github.com/htamagnus/ticket-sales-backend-golang/internal/events/domain"
	_ "github.com/go-sql-driver/mysql"
)

type mysqlEventRepository struct {
	db *sql.DB
}

// NewMysqlEventRepository creates a new MySQL event repository.
func NewMysqlEventRepository(db *sql.DB) (domain.EventRepository, error) {
	return &mysqlEventRepository{db: db}, nil
}

// ListEvents returns all events with their associated spots and tickets.
func (r *mysqlEventRepository) ListEvents() ([]domain.Event, error) {
	query := `
		SELECT 
			e.id, e.name, e.location, e.organization, e.rating, e.date, e.image_url, e.capacity, e.price, e.partner_id,
			s.id, s.event_id, s.name, s.status, s.ticket_id,
			t.id, t.event_id, t.spot_id, t.ticket_type, t.price
		FROM events e
		LEFT JOIN spots s ON e.id = s.event_id
		LEFT JOIN tickets t ON s.id = t.spot_id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	eventMap := make(map[string]*domain.Event)
	spotMap := make(map[string]*domain.Spot)
	for rows.Next() {
		var eventID, eventName, eventLocation, eventOrganization, eventRating, eventImageURL, spotID, spotEventID, spotName, spotStatus, spotTicketID, ticketID, ticketEventID, ticketSpotID, ticketType sql.NullString
		var eventDate sql.NullString
		var eventCapacity int
		var eventPrice, ticketPrice sql.NullFloat64
		var partnerID sql.NullInt32

		err := rows.Scan(
			&eventID, &eventName, &eventLocation, &eventOrganization, &eventRating, &eventDate, &eventImageURL, &eventCapacity, &eventPrice, &partnerID,
			&spotID, &spotEventID, &spotName, &spotStatus, &spotTicketID,
			&ticketID, &ticketEventID, &ticketSpotID, &ticketType, &ticketPrice,
		)
		if err != nil {
			return nil, err
		}

		if !eventID.Valid || !eventName.Valid || !eventLocation.Valid || !eventOrganization.Valid || !eventRating.Valid || !eventDate.Valid || !eventImageURL.Valid || !eventPrice.Valid || !partnerID.Valid {
			continue
		}

		event, exists := eventMap[eventID.String]
		if !exists {
			eventDateParsed, err := time.Parse("2006-01-02 15:04:05", eventDate.String)
			if err != nil {
				return nil, err
			}
			event = &domain.Event{
				ID:           eventID.String,
				Name:         eventName.String,
				Location:     eventLocation.String,
				Organization: eventOrganization.String,
				Rating:       domain.Rating(eventRating.String),
				Date:         eventDateParsed,
				ImageURL:     eventImageURL.String,
				Capacity:     eventCapacity,
				Price:        eventPrice.Float64,
				PartnerID:    int(partnerID.Int32),
				Spots:        []domain.Spot{},
				Tickets:      []domain.Ticket{},
			}
			eventMap[eventID.String] = event
		}

		if spotID.Valid {
			spot, spotExists := spotMap[spotID.String]
			if !spotExists {
				spot = &domain.Spot{
					ID:       spotID.String,
					EventID:  spotEventID.String,
					Name:     spotName.String,
					Status:   domain.SpotStatus(spotStatus.String),
					TicketID: spotTicketID.String,
				}
				event.Spots = append(event.Spots, *spot)
				spotMap[spotID.String] = spot
			}

			if ticketID.Valid {
				ticket := domain.Ticket{
					ID:         ticketID.String,
					EventID:    ticketEventID.String,
					Spot:       spot,
					TicketType: domain.TicketType(ticketType.String),
					Price:      ticketPrice.Float64,
				}
				event.Tickets = append(event.Tickets, ticket)
			}
		}
	}

	events := make([]domain.Event, 0, len(eventMap))
	for _, event := range eventMap {
		events = append(events, *event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

}
