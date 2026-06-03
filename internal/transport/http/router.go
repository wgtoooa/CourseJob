package http

import (
	"CourseJob/internal/domain"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"log"
	nethttp "net/http"
)

func NewRouter(h *Handler, log *log.Logger) nethttp.Handler {
	r := chi.NewRouter()

	//r.Post("api/v1/login", h.Login)
	//r.Post("api/v1/register", h.Register)
	//r.Use(CORS())
	logging := Logging(log)
	r.Get("/health/live", h.Leave)
	r.Get("/health/ready", h.Ready)

	r.Post("/auth/register/student", h.RegisterStudent)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/revoke", h.Revoke)

	r.Get("/api/v1/teachers", h.GetTeachers)
	r.Get("/api/v1/subjects", h.GetSubjects)
	r.Get("/api/v1/rooms", h.GetRooms)
	r.Get("/api/v1/groups", h.GetGroups)
	r.Get("/api/v1/schedule", h.GetScheduleByCourse)
	r.Get("/api/v1/sse/schedule", h.ScheduleSSE)
	r.Get("/api/v1/plan", h.GetPlanByCourse)
	r.Post("/api/v1/schedule", h.ScheduleWebHook)
	r.Put("/api/v1/plan", h.UpsertPlanItem)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Post("/api/v1/student", h.AddStudent)

	r.Route("/api/v1/attendance", func(ar chi.Router) {
		ar.Use(h.AuthRequired)
		ar.Use(h.RequireRoles(domain.RoleTeacher, domain.RoleAdmin))
		ar.Post("/sessions", h.CreateAttendanceSession)
	})

	r.Route("/api/v1/student/me", func(sr chi.Router) {
		sr.Use(h.AuthRequired)
		sr.Use(h.RequireRoles(domain.RoleStudent))
		sr.Get("/attendance", h.GetMyStudentAttendance)
	})

	r.Route("/admin", func(ar chi.Router) {
		ar.Use(h.AuthRequired)
		ar.Use(h.RequireRoles(domain.RoleAdmin))
		ar.Get("/users", h.ListUsersByAdmin)
		ar.Post("/users", h.CreateUserByAdmin)
		ar.Patch("/users/{id}/role", h.UpdateUserRoleByAdmin)
		ar.Patch("/users/{id}/active", h.UpdateUserActiveByAdmin)
	})

	return logging(r)
}
