package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}
type StudentRepo interface {
	GetByCardUID(ctx context.Context, cardUID string) (*domain.Student, error)
	CreateStudent(ctx context.Context, st *domain.Student) error
}

type SessionRepo interface {
	Create(ctx context.Context, session *domain.AttendanceSession) error
}

type EventRepo interface {
	Create(ctx context.Context, event *domain.AttendanceEvent) error
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

type Repository interface {
	Students() StudentRepo
	Sessions() SessionRepo
	Events() EventRepo
}

func NewRepositories(db DBTX) Repository {
	return &repositories{
		studentRepo: NewStudentRepository(db),
		sessionRepo: NewAttendanceSessionRepository(db),
		eventRepo:   NewAttendanceEventRepository(db),
	}
}

type repositories struct {
	studentRepo StudentRepo
	sessionRepo SessionRepo
	eventRepo   EventRepo
}

func (re *repositories) Students() StudentRepo {
	return re.studentRepo
}

func (re *repositories) Sessions() SessionRepo {
	return re.sessionRepo
}
func (re *repositories) Events() EventRepo {
	return re.eventRepo
}
