package http

import (
	"github.com/go-chi/chi/v5"
	"log"
	nethttp "net/http"
)

func NewRouter(h *Handler, log *log.Logger) nethttp.Handler {
	r := chi.NewRouter()

	//r.Post("api/v1/login", h.Login)
	//r.Post("api/v1/register", h.Register)
	logging := Logging(log)
	r.Get("/health/live", h.Leave)
	r.Get("/health/ready", h.Ready)

	r.Get("/api/v1/teachers", h.GetTeachers)
	r.Get("/api/v1/subjects", h.GetSubjects)
	r.Get("/api/v1/rooms", h.GetRooms)
	r.Post("/api/v1/attendance/sessions", h.CreateAttendanceSession)
	r.Post("/api/v1/student", h.AddStudent)
	r.Post("/api/v1/schedule", h.ScheduleWebHook)
	return logging(r)
}
