package http

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	"fmt"
	nethttp "net/http"
)

func (h *Handler) ScheduleWebHook(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "Method Not Allowed",
		})
		return
	}
	r.Body = nethttp.MaxBytesReader(w, r.Body, 1<<20) //~ 1 MB
	defer r.Body.Close()

	var req dto.WeekScheduleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	weekSchedule := domain.WeekSchedule{
		Name:        req.Name,
		GeneratedAt: req.GeneratedAt,
		Course:      req.Course,
		Semester:    req.Semester,
		WeekNumber:  req.WeekNumber,
		Groups:      make([]domain.ScheduleGroup, 0, len(req.Groups)),
		Lessons:     make([]domain.ScheduleLesson, 0, len(req.Lessons)),
	}
	for _, group := range req.Groups {
		weekSchedule.Groups = append(weekSchedule.Groups, domain.ScheduleGroup{
			ID:         group.ID,
			Name:       group.Name,
			Specialty:  group.Specialty,
			Department: group.Department,
		})
	}
	for _, lesson := range req.Lessons {
		weekSchedule.Lessons = append(weekSchedule.Lessons, domain.ScheduleLesson{
			Day:         lesson.Day,
			DayNumber:   lesson.DayNumber,
			Pair:        lesson.Pair,
			Duration:    lesson.Duration,
			Time:        lesson.Time,
			Group:       lesson.Group,
			Type:        lesson.Type,
			Subject:     lesson.Subject,
			Teacher:     lesson.Teacher,
			Room:        lesson.Room,
			Subgroup:    lesson.Subgroup,
			Frequency:   lesson.Frequency,
			PeriodStart: lesson.PeriodStart,
			PeriodEnd:   lesson.PeriodEnd,
			Comment:     lesson.Comment,
			Cancelled:   lesson.Cancelled,
		})
	}
	input := service.ScheduleImportInput{
		WeekSchedule: weekSchedule,
	}

	if err = h.attendanceService.ScheduleImport(r.Context(), input); err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("failed to import schedule: %v", err),
		})

		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"status": "success",
	})
}
