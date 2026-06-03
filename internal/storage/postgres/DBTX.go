package postgres

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/attendance"
	authRepo "CourseJob/internal/storage/postgres/auth"
	"CourseJob/internal/storage/postgres/contract"
	"CourseJob/internal/storage/postgres/schedule"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DBTX = contract.DBTX
type StudentRepo interface {
	GetByCardUID(ctx context.Context, cardUID string) (*domain.Student, error)
	CreateStudent(ctx context.Context, st *domain.Student) error
	CreateStudents(ctx context.Context, students []domain.Student) ([]string, error)
}

type SessionRepo interface {
	Create(ctx context.Context, session *domain.AttendanceSession) error
}

type ReportRepo interface {
	BuildSessionReport(ctx context.Context, sessionID int64) error
	ListStudentAttendance(ctx context.Context, studentID int64, fromDate *time.Time, toDate *time.Time) ([]domain.StudentAttendanceRecord, error)
}

type EventRepo interface {
	Create(ctx context.Context, event *domain.AttendanceEvent) (bool, error)
}
type ScheduleRepo interface {
	UpSetWeekSchedule(ctx context.Context, weekSchedule *domain.WeekSchedule) error
	UpSetLessons(ctx context.Context, lessons *domain.ScheduleLesson) error
	UpSetGroup(ctx context.Context, group *domain.ScheduleGroup) error
	GetCourseSchedule(ctx context.Context, course int, filters domain.ScheduleFilters) (*domain.ScheduleCourse, error)
	FindLessonBySession(ctx context.Context, sessionID int64) (*domain.ScheduleLesson, error)
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

type PlanRepo interface {
	GetPlanByCourse(ctx context.Context, course int) ([]domain.PlanItem, error)
	UpsertPlanItem(ctx context.Context, item *domain.PlanItem) error
}

type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	CountByRole(ctx context.Context, role string) (int64, error)
	List(ctx context.Context, role string, emailQuery string, limit int, offset int) ([]domain.User, error)
	UpdateRole(ctx context.Context, id int64, role string) (*domain.User, error)
	UpdateActive(ctx context.Context, id int64, isActive bool) (*domain.User, error)
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, session *domain.RefreshTokenSession) error
	RotateAndRevoke(ctx context.Context, tokenID, tokenHash string, userID int64, usedAt time.Time, replacedByTokenID string) (bool, error)
	Revoke(ctx context.Context, tokenID, tokenHash string, userID int64, usedAt time.Time) (bool, error)
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
	Plan() PlanRepo
	Report() ReportRepo
	Users() UserRepo
	RefreshTokens() RefreshTokenRepo
}

func NewRepositories(db DBTX) Repository {
	return &repositories{
		studentRepo:      schedule.NewStudentRepository(db),
		sessionRepo:      attendance.NewAttendanceSessionRepository(db),
		eventRepo:        attendance.NewAttendanceEventRepository(db),
		scheduleRepo:     schedule.NewScheduleRepository(db),
		teacherRepo:      schedule.NewTeacherRepository(db),
		subjectRepo:      schedule.NewSubjectRepository(db),
		roomRepo:         schedule.NewRoomRepository(db),
		planRepo:         schedule.NewPlanRepository(db),
		reportRepo:       attendance.NewAttendanceReportRepository(db),
		userRepo:         authRepo.NewUserRepository(db),
		refreshTokenRepo: authRepo.NewRefreshTokenRepository(db),
	}
}

type repositories struct {
	studentRepo      StudentRepo
	sessionRepo      SessionRepo
	eventRepo        EventRepo
	scheduleRepo     ScheduleRepo
	teacherRepo      TeacherRepo
	subjectRepo      SubjectRepo
	roomRepo         RoomRepo
	planRepo         PlanRepo
	reportRepo       ReportRepo
	userRepo         UserRepo
	refreshTokenRepo RefreshTokenRepo
}

func (re *repositories) Students() StudentRepo {
	return re.studentRepo
}
func (re *repositories) Teachers() TeacherRepo { return re.teacherRepo }
func (re *repositories) Subjects() SubjectRepo { return re.subjectRepo }
func (re *repositories) Rooms() RoomRepo       { return re.roomRepo }
func (re *repositories) Plan() PlanRepo        { return re.planRepo }
func (re *repositories) Sessions() SessionRepo {
	return re.sessionRepo
}
func (re *repositories) Events() EventRepo {
	return re.eventRepo
}
func (re *repositories) Schedule() ScheduleRepo { return re.scheduleRepo }
func (re *repositories) Report() ReportRepo     { return re.reportRepo }
func (re *repositories) Users() UserRepo        { return re.userRepo }
func (re *repositories) RefreshTokens() RefreshTokenRepo {
	return re.refreshTokenRepo
}
