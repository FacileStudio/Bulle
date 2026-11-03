package spaces

import "time"

type SpaceResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Icon            string    `json:"icon"`
	MyceliumSpaceID string    `json:"mycelium_space_id"`
	OpusWorkspace   string    `json:"opus_workspace"`
	AgendaCalendar  int64     `json:"agenda_calendar"`
	CourrierAccount int64     `json:"courrier_account"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SpaceRequest struct {
	Name            *string `json:"name,omitempty"`
	Icon            *string `json:"icon,omitempty"`
	MyceliumSpaceID *string `json:"mycelium_space_id,omitempty"`
	OpusWorkspace   *string `json:"opus_workspace,omitempty"`
	AgendaCalendar  *int64  `json:"agenda_calendar,omitempty"`
	CourrierAccount *int64  `json:"courrier_account,omitempty"`
}

type EmptyResponse struct{}
