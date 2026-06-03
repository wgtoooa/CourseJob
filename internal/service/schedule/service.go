package schedule

import "CourseJob/internal/storage/postgres"

type Service struct {
	transactor postgres.Transactor
}

func NewService(transactor postgres.Transactor) *Service {
	return &Service{transactor: transactor}
}
