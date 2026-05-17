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
type ScheduleRepo interface {
	UpSetWeekSchedule(ctx context.Context, weekSchedule *domain.WeekSchedule) error
	UpSetLessons(ctx context.Context, lessons *domain.ScheduleLesson) error
	UpSetGroup(ctx context.Context, group *domain.ScheduleGroup) error
}

type TeacherRepo interface {
	UpSetTeacher(ctx context.Context, teacher *domain.Teacher) error
	GetTeachers(ctx context.Context) ([]domain.Teacher, error)
}

type SubjectRepo interface {
	UpSetSubject(ctx context.Context, subject *domain.Subject) error
	GetSubjects(ctx context.Context) ([]domain.Subject, error)
}

type RoomRepo interface {
	UpSetRoom(ctx context.Context, room *domain.Room) error
	GetRooms(ctx context.Context) ([]domain.Room, error)
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
	Schedule() ScheduleRepo
	Teachers() TeacherRepo
	Subjects() SubjectRepo
	Rooms() RoomRepo
}

func NewRepositories(db DBTX) Repository {
	return &repositories{
		studentRepo:  NewStudentRepository(db),
		sessionRepo:  NewAttendanceSessionRepository(db),
		eventRepo:    NewAttendanceEventRepository(db),
		scheduleRepo: NewScheduleRepository(db),
		teacherRepo:  NewTeacherRepository(db),
		subjectRepo:  NewSubjectRepository(db),
		roomRepo:     NewRoomRepository(db),
	}
}

type repositories struct {
	studentRepo  StudentRepo
	sessionRepo  SessionRepo
	eventRepo    EventRepo
	scheduleRepo ScheduleRepo
	teacherRepo  TeacherRepo
	subjectRepo  SubjectRepo
	roomRepo     RoomRepo
}

func (re *repositories) Students() StudentRepo {
	return re.studentRepo
}
func (re *repositories) Teachers() TeacherRepo { return re.teacherRepo }
func (re *repositories) Subjects() SubjectRepo { return re.subjectRepo }
func (re *repositories) Rooms() RoomRepo       { return re.roomRepo }
func (re *repositories) Sessions() SessionRepo {
	return re.sessionRepo
}
func (re *repositories) Events() EventRepo {
	return re.eventRepo
}
func (re *repositories) Schedule() ScheduleRepo { return re.scheduleRepo }
