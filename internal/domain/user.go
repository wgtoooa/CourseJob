package domain

import "time"

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	StudentID    *int64
	TeacherID    *int64
	IsActive     bool
	CreatedAt    time.Time
}
