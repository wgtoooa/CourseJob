package http

import (
	attendanceService "CourseJob/internal/service/attendance"
	"CourseJob/internal/transport/http/dto"
	"errors"
	nethttp "net/http"
	"strings"
	"time"
)

// GetMyStudentAttendance returns attendance report for the authenticated student.
// @Summary Get my attendance
// @Description Returns attendance records and summary for the current authenticated student.
// @Tags Students
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param from query string false "Start date in YYYY-MM-DD"
// @Param to query string false "End date in YYYY-MM-DD"
// @Success 200 {object} dto.StudentAttendanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/student/me/attendance [get]
func (h *Handler) GetMyStudentAttendance(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{
			Status: "error",
			Error:  "unauthorized",
		})
		return
	}

	filter, err := parseStudentAttendanceFilter(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	report, err := h.attendanceService.GetMyAttendance(r.Context(), principal.UserID, principal.Role, filter)
	if err != nil {
		switch {
		case errors.Is(err, attendanceService.ErrStudentAccessDenied):
			writeJSON(w, nethttp.StatusForbidden, ErrorResponse{
				Status: "error",
				Error:  err.Error(),
			})
		case errors.Is(err, attendanceService.ErrStudentProfileNotFound):
			writeJSON(w, nethttp.StatusNotFound, ErrorResponse{
				Status: "error",
				Error:  err.Error(),
			})
		default:
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{
				Status: "error",
				Error:  "internal server error",
			})
		}
		return
	}

	records := make([]dto.StudentAttendanceRecordResponse, 0, len(report.Records))
	lessonBySessionID := make(map[int64]*dto.StudentAttendanceRecordResponse, len(report.Records))
	for _, rec := range report.Records {
		var scannedAt *string
		if rec.ScannedAt != nil {
			formatted := rec.ScannedAt.UTC().Format(time.RFC3339)
			scannedAt = &formatted
		}

		var (
			subject    *string
			group      *string
			pair       *int
			lessonType *string
			teacher    *string
		)
		if h.scheduleService != nil {
			cached, exists := lessonBySessionID[rec.SessionID]
			if exists && cached != nil {
				subject = cached.Subject
				group = cached.Group
				pair = cached.Pair
				lessonType = cached.LessonType
				teacher = cached.Teacher
			} else {
				lesson, lessonErr := h.scheduleService.FindLessonBySession(r.Context(), rec.SessionID)
				if lessonErr == nil && lesson != nil {
					if lesson.Subject != "" {
						value := lesson.Subject
						subject = &value
					}
					if lesson.Group != "" {
						value := lesson.Group
						group = &value
					}
					if lesson.Pair > 0 {
						value := lesson.Pair
						pair = &value
					}
					if lesson.Type != "" {
						value := lesson.Type
						lessonType = &value
					}
					if lesson.Teacher != nil && *lesson.Teacher != "" {
						value := *lesson.Teacher
						teacher = &value
					}
				}
			}
		}

		record := dto.StudentAttendanceRecordResponse{
			SessionID:  rec.SessionID,
			Date:       rec.Date.Format("2006-01-02"),
			Room:       rec.Room,
			Source:     rec.Source,
			StartedAt:  rec.StartedAt.UTC().Format(time.RFC3339),
			FinishedAt: rec.FinishedAt.UTC().Format(time.RFC3339),
			Present:    rec.Present,
			ScannedAt:  scannedAt,
			Subject:    subject,
			Group:      group,
			Pair:       pair,
			LessonType: lessonType,
			Teacher:    teacher,
		}

		records = append(records, record)
		lessonBySessionID[rec.SessionID] = &record
	}

	response := dto.StudentAttendanceResponse{
		Status:    "success",
		StudentID: report.StudentID,
		Summary: dto.StudentAttendanceSummary{
			TotalSessions:     report.TotalSessions,
			PresentCount:      report.PresentCount,
			AbsentCount:       report.AbsentCount,
			AttendancePercent: report.AttendancePercent,
		},
		Records: records,
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func parseStudentAttendanceFilter(r *nethttp.Request) (attendanceService.StudentAttendanceFilter, error) {
	rawFrom := strings.TrimSpace(r.URL.Query().Get("from"))
	rawTo := strings.TrimSpace(r.URL.Query().Get("to"))

	var (
		from *time.Time
		to   *time.Time
	)

	if rawFrom != "" {
		parsedFrom, err := time.Parse("2006-01-02", rawFrom)
		if err != nil {
			return attendanceService.StudentAttendanceFilter{}, errors.New("from must be YYYY-MM-DD")
		}
		value := parsedFrom.UTC()
		from = &value
	}
	if rawTo != "" {
		parsedTo, err := time.Parse("2006-01-02", rawTo)
		if err != nil {
			return attendanceService.StudentAttendanceFilter{}, errors.New("to must be YYYY-MM-DD")
		}
		value := parsedTo.UTC()
		to = &value
	}

	if from != nil && to != nil && from.After(*to) {
		return attendanceService.StudentAttendanceFilter{}, errors.New("from must be before or equal to to")
	}

	return attendanceService.StudentAttendanceFilter{
		FromDate: from,
		ToDate:   to,
	}, nil
}
