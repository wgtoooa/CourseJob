package postgres

import (
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

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

type Repositories interface {
	Students() *StudentRepository
	Sessions() *AttendanceSessionRepository
	Events() *AttendanceEventRepository
}

func NewRepository(db DBTX) Repositories {
	return &repositories{
		studentRepo: NewStudentRepository(db),
		sessionRepo: NewAttendanceSessionRepository(db),
		eventRepo:   NewAttendanceEventRepository(db),
	}
}

type repositories struct {
	studentRepo *StudentRepository
	sessionRepo *AttendanceSessionRepository
	eventRepo   *AttendanceEventRepository
}

func (re *repositories) Students() *StudentRepository {
	return re.studentRepo
}

func (re *repositories) Sessions() *AttendanceSessionRepository {
	return re.sessionRepo
}
func (re *repositories) Events() *AttendanceEventRepository {
	return re.eventRepo
}
