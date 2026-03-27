package dto

import "time"

type StudentRequest struct {
	FullName  string    `json:"full_name"`
	Course    int       `json:"course"`
	GroupName string    `json:"group_name"`
	CardUID   string    `json:"card_uid"`
	CreatedAt time.Time `json:"created_at"`
}
