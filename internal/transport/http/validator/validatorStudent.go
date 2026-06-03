package validator

import (
	authA "CourseJob/internal/auth"
	"CourseJob/internal/transport/http/dto"
	"errors"
	"strings"
)

func NormalizeStudentRequest(st *dto.StudentRequest) {
	st.FullName = strings.TrimSpace(st.FullName)
	st.GroupName = strings.TrimSpace(st.GroupName)
	st.Email = strings.ToLower(strings.TrimSpace(st.Email))

	uid := authA.NormalizeCardUID(st.UID)
	cardUID := authA.NormalizeCardUID(st.CardUID)
	if uid == "" {
		uid = cardUID
	}
	if cardUID == "" {
		cardUID = uid
	}

	st.UID = uid
	st.CardUID = cardUID
}

func ValidatorStudent(st *dto.StudentRequest) error {
	if st == nil {
		return errors.New("student is nil")
	}
	if st.UID != "" && st.CardUID != "" && st.UID != st.CardUID {
		return errors.New("uid and card_uid must match")
	}
	if !validUID(st.CardUID) {
		return errors.New("invalid uid")
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
