package http

import (
	"CourseJob/internal/domain"
	serviceSchedule "CourseJob/internal/service/schedule"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"
)

// ScheduleWebHook imports a schedule chunk.
// @Summary Import schedule
// @Description Imports one course schedule chunk and publishes an SSE update event.
// @Tags Schedule
// @Accept json
// @Produce json
// @Param request body dto.WeekScheduleRequest true "Week schedule payload"
// @Success 200 {object} ScheduleImportSuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedule [post]
func (h *Handler) ScheduleWebHook(w nethttp.ResponseWriter, r *nethttp.Request) {
	h.scheduleImportHandler(w, r, nethttp.MethodPost)
}

func (h *Handler) scheduleImportHandler(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
	expectedMethod string,
) {
	if r.Method != expectedMethod {
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
			"ok":      false,
			"status":  "error",
			"error":   "invalid_payload",
			"message": err.Error(),
		})
		return
	}

	if req.Course <= 0 || req.WeekNumber <= 0 || len(req.Lessons) == 0 {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"ok":      false,
			"status":  "error",
			"error":   "invalid_payload",
			"message": "course, week_number and lessons are required",
		})
		return
	}

	weekSchedule := domain.WeekSchedule{
		Name:        req.Name,
		GeneratedAt: req.GeneratedAt,
		Course:      req.Course,
		Semester:    req.Semester,
		WeekNumber:  req.WeekNumber,
		DateRange:   req.DateRange,
		Groups:      make([]domain.ScheduleGroup, 0, len(req.Groups)),
		Lessons:     make([]domain.ScheduleLesson, 0, len(req.Lessons)),
	}
	for _, group := range req.Groups {
		groupID := strings.TrimSpace(group.ID)
		groupName := strings.TrimSpace(group.Name)
		if groupID == "" {
			groupID = groupName
		}
		if groupName == "" {
			groupName = groupID
		}

		weekSchedule.Groups = append(weekSchedule.Groups, domain.ScheduleGroup{
			ID:         groupID,
			Name:       groupName,
			Specialty:  strings.TrimSpace(group.Specialty),
			Department: strings.TrimSpace(group.Department),
		})
	}
	for _, lesson := range req.Lessons {
		weekSchedule.Lessons = append(weekSchedule.Lessons, domain.ScheduleLesson{
			Day:           lesson.Day,
			DayNumber:     lesson.DayNumber,
			Date:          lesson.Date,
			Pair:          lesson.Pair,
			Duration:      lesson.Duration,
			Time:          lesson.Time,
			Group:         strings.TrimSpace(lesson.Group),
			Type:          lesson.Type,
			Subject:       lesson.Subject,
			Teacher:       lesson.Teacher,
			Room:          lesson.Room,
			Subgroup:      lesson.Subgroup,
			Frequency:     lesson.Frequency,
			PeriodStart:   lesson.PeriodStart,
			PeriodEnd:     lesson.PeriodEnd,
			Comment:       lesson.Comment,
			Cancelled:     lesson.Cancelled,
			GoogleSheetID: lesson.GoogleSheetID,
		})
	}
	input := serviceSchedule.ScheduleImportInput{
		WeekSchedule: weekSchedule,
	}

	err = h.scheduleService.ScheduleImport(r.Context(), input)
	if err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"ok":      false,
			"status":  "error",
			"error":   "import_failed",
			"message": fmt.Sprintf("failed to import schedule: %v", err),
		})

		return
	}

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	h.scheduleEvents.publish(scheduleUpdatedEvent{
		Type:      "schedule_updated",
		UpdatedAt: updatedAt,
		Chunk:     req,
	})

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"ok":            true,
		"status":        "success",
		"course":        req.Course,
		"week_number":   req.WeekNumber,
		"lessons_count": len(req.Lessons),
		"updated_at":    updatedAt,
	})
}

// GetScheduleByCourse returns schedule data for a specific course.
// @Summary Get schedule by course
// @Description Returns schedule grouped by weeks with optional filters.
// @Tags Schedule
// @Produce json
// @Param course query int true "Course number (1-4)"
// @Param week query int false "Week number"
// @Param group query string false "Group filter"
// @Param day query string false "Day filter"
// @Param type query string false "Lesson type filter"
// @Param teacher query string false "Teacher filter"
// @Param subject query string false "Subject filter"
// @Success 200 {object} dto.CourseScheduleResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedule [get]
func (h *Handler) GetScheduleByCourse(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	courseRaw := strings.TrimSpace(r.URL.Query().Get("course"))
	if courseRaw == "" {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "course query parameter is required",
		})
		return
	}

	course, err := strconv.Atoi(courseRaw)
	if err != nil || course < 1 || course > 4 {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "course must be an integer from 1 to 4",
		})
		return
	}

	filters, err := parseScheduleFilters(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	data, err := h.scheduleService.GetScheduleByCourse(r.Context(), course, filters)
	if err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("failed to get schedule: %v", err),
		})
		return
	}

	response := dto.CourseScheduleResponse{
		Course:      data.Course,
		GeneratedAt: data.GeneratedAt,
		Groups:      make([]dto.ScheduleGroupSummaryDTO, 0, len(data.Groups)),
		Weeks:       make([]dto.WeekScheduleResponse, 0, len(data.Weeks)),
	}

	for _, group := range data.Groups {
		response.Groups = append(response.Groups, dto.ScheduleGroupSummaryDTO{
			ID:    group.ID,
			Name:  group.Name,
			Count: group.Count,
		})
	}

	for _, week := range data.Weeks {
		weekResp := dto.WeekScheduleResponse{
			Name:        week.Name,
			GeneratedAt: week.GeneratedAt,
			Course:      week.Course,
			Semester:    week.Semester,
			WeekNumber:  week.WeekNumber,
			DateRange:   week.DateRange,
			Groups:      make([]dto.ScheduleGroupSummaryDTO, 0, len(week.Groups)),
			Lessons:     make([]dto.ScheduleLessonResponse, 0, len(week.Lessons)),
		}

		for _, group := range week.Groups {
			weekResp.Groups = append(weekResp.Groups, dto.ScheduleGroupSummaryDTO{
				ID:    group.ID,
				Name:  group.Name,
				Count: group.Count,
			})
		}

		for _, lesson := range week.Lessons {
			weekResp.Lessons = append(weekResp.Lessons, dto.ScheduleLessonResponse{
				Day:           lesson.Day,
				DayNumber:     lesson.DayNumber,
				Date:          lesson.Date,
				Pair:          lesson.Pair,
				Duration:      lesson.Duration,
				Time:          lesson.Time,
				Group:         lesson.Group,
				Type:          lesson.Type,
				Subject:       lesson.Subject,
				Teacher:       lesson.Teacher,
				Room:          lesson.Room,
				Subgroup:      lesson.Subgroup,
				Frequency:     lesson.Frequency,
				PeriodStart:   lesson.PeriodStart,
				PeriodEnd:     lesson.PeriodEnd,
				Comment:       lesson.Comment,
				Cancelled:     lesson.Cancelled,
				WeekNumber:    lesson.WeekNumber,
				GoogleSheetID: lesson.GoogleSheetID,
			})
		}

		response.Weeks = append(response.Weeks, weekResp)
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func parseScheduleFilters(r *nethttp.Request) (domain.ScheduleFilters, error) {
	f := domain.ScheduleFilters{
		Group:      strings.TrimSpace(r.URL.Query().Get("group")),
		Day:        strings.TrimSpace(r.URL.Query().Get("day")),
		LessonType: strings.TrimSpace(r.URL.Query().Get("type")),
		Teacher:    strings.TrimSpace(r.URL.Query().Get("teacher")),
		Subject:    strings.TrimSpace(r.URL.Query().Get("subject")),
	}

	weekRaw := strings.TrimSpace(r.URL.Query().Get("week"))
	if weekRaw == "" {

		return f, nil
	}

	week, err := strconv.Atoi(weekRaw)
	if err != nil || week < 1 || week > 52 {
		return domain.ScheduleFilters{}, fmt.Errorf("week must be an integer from 1 to 14")
	}
	f.Week = week
	return f, nil
}
