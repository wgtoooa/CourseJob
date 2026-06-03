package http

import (
	"CourseJob/internal/transport/http/dto"
)

type HealthLiveResponse struct {
	Status string `json:"status" example:"ok"`
}

type HealthReadyResponse struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"connected"`
}

type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Error   string `json:"error" example:"invalid request body"`
	Message string `json:"message,omitempty" example:"course query parameter is required"`
}

type StudentCreateResponse struct {
	Status       string `json:"status" example:"created"`
	CreatedCount int    `json:"created_count,omitempty" example:"2"`
}

type AttendanceCreateResponse struct {
	Status string                 `json:"status" example:"created"`
	Data   dto.AttendanceResponse `json:"data"`
}

type ScheduleImportSuccessResponse struct {
	OK           bool   `json:"ok" example:"true"`
	Status       string `json:"status" example:"success"`
	Course       int    `json:"course" example:"3"`
	WeekNumber   int    `json:"week_number" example:"14"`
	LessonsCount int    `json:"lessons_count" example:"120"`
	UpdatedAt    string `json:"updated_at" example:"2026-05-17T19:00:00Z"`
}

type PlanUpsertSuccessResponse struct {
	Status string `json:"status" example:"success"`
}

type TeachersResponse struct {
	Status   string               `json:"status" example:"success"`
	Teachers []TeacherCatalogItem `json:"teachers"`
}

type SubjectsResponse struct {
	Status   string               `json:"status" example:"success"`
	Subjects []SubjectCatalogItem `json:"subjects"`
}

type RoomsResponse struct {
	Status string            `json:"status" example:"success"`
	Rooms  []RoomCatalogItem `json:"rooms"`
}

type GroupsResponse struct {
	Status string             `json:"status" example:"success"`
	Course int                `json:"course" example:"3"`
	Groups []GroupCatalogItem `json:"groups"`
}

type TeacherCatalogItem struct {
	PostFullName string `json:"post_full_name" example:"доц. Иванов И.И."`
}

type SubjectCatalogItem struct {
	Name string `json:"name" example:"Высшая математика"`
}

type RoomCatalogItem struct {
	Name string `json:"name" example:"115"`
}

type GroupCatalogItem struct {
	ID    string `json:"id" example:"4"`
	Name  string `json:"name" example:"Группа 4"`
	Count int    `json:"count" example:"28"`
}
