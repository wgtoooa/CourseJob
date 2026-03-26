package dto

type AttendanceResponse struct {
	SessionID     int64    `json:"session_id"`
	SavedEvents   int      `json:"saved_events"`
	NotFoundCards []string `json:"not_found_cards"`
}
