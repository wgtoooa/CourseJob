package validator

import (
	"CourseJob/internal/transport/http/dto"
	"errors"
	"strings"
)

func NormalizeStudentRequest(st *dto.StudentRequest) {
	st.FullName = strings.TrimSpace(st.FullName)
	st.CardUID = strings.ToUpper(strings.TrimSpace(st.CardUID))
	st.GroupName = strings.TrimSpace(st.GroupName)
	st.Email = strings.ToLower(strings.TrimSpace(st.Email))
}

func ValidatorStudent(st *dto.StudentRequest) error {
	if st == nil {
		return errors.New("student is nil")
	}
	if !validUID(st.CardUID) {
		return errors.New("invalid uid card")
	}
	if st.GroupName == "" {
		return errors.New("group name empty")
	}
	if st.Course <= 0 || st.Course > 4 {
		return errors.New("invalid course")
	}
	if st.FullName == "" {
		return errors.New("full name empty")
	}
	if !validEmail(st.Email) {
		return errors.New("invalid email")
	}
	return nil
}
