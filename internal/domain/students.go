package domain

import "time"

type Student struct {
	ID        int64
	FullName  string
	Course    int
	GroupName string
	CardUID   string
	CreatedAt time.Time
}
