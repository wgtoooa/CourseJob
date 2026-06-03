package dto

import "time"

type StudentRequest struct {
	FullName  string     `json:"full_name"`
	Course    int        `json:"course"`
	GroupName string     `json:"group_name"`
	Email     string     `json:"email"`
	UID       string     `json:"uid,omitempty"`
	CardUID   string     `json:"card_uid,omitempty"`
	CreatedAt *time.Time `json:"created_at"`
}
