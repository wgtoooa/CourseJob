package dto

type AttendanceResponse struct {
	SessionID    int      `json:"session_id"`
	SavedEvents  int      `json:"saved_events"`
	NotFoundCard []string `json:"not_found_card"`
}
