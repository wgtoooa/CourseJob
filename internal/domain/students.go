package domain

import "time"

type Students struct {
	ID        int64
	FullName  string
	Course    int
	GroupName string
	CardUID   string
	CreateAt  time.Time
}
