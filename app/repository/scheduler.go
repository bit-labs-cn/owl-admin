package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	schedulerpkg "bit-labs.cn/owl/provider/scheduler"
	"gorm.io/gorm"
)

// SchedulerRepositoryInterface 是定时任务与执行记录的仓储。
//
// 任务的排期、抢占与执行由框架的调度引擎负责，这里只做管理端需要的查询与自定义任务维护，
// 不参与调度语义。
type SchedulerRepositoryInterface interface {
	// RetrieveJobs 分页查询任务
	RetrieveJobs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []schedulerpkg.Job, err error)
	// FindJob 按主键查询任务
	FindJob(id uint) (*schedulerpkg.Job, error)
	// CreateJob 创建自定义任务
	CreateJob(job *schedulerpkg.Job) error
	// SaveJob 保存任务的全部字段
	SaveJob(job *schedulerpkg.Job) error
	// UpdateJobFields 更新任务的指定字段
	UpdateJobFields(id uint, updates map[string]any) error
	// DeleteJob 删除任务
	DeleteJob(id uint) error

	// RetrieveRuns 分页查询执行记录
	RetrieveRuns(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []schedulerpkg.Run, err error)
	// FindRun 按主键查询执行记录
	FindRun(id uint) (*schedulerpkg.Run, error)
	// CountActiveRuns 统计某条任务仍在排队或执行中的记录数
	CountActiveRuns(jobID uint) (int64, error)
	// DeleteTerminalRunsBefore 删除指定时间之前的终态执行记录
	DeleteTerminalRunsBefore(cutoff time.Time) (int64, error)

	// CountJobsByStatus 任务按状态计数
	CountJobsByStatus() (map[int8]int64, error)
	// CountRunsByStatusSince 指定时间之后创建的执行记录按状态计数
	CountRunsByStatusSince(since time.Time) (map[int8]int64, error)
	// RecentFailedRuns 最近的失败记录
	RecentFailedRuns(limit int) ([]schedulerpkg.Run, error)

	contract.WithContext[SchedulerRepositoryInterface]
}

var _ SchedulerRepositoryInterface = (*SchedulerRepository)(nil)

// SchedulerRepository 基于 GORM 的调度仓储实现。
type SchedulerRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewSchedulerRepository(d *gorm.DB) SchedulerRepositoryInterface {
	return &SchedulerRepository{db: d}
}

func (i *SchedulerRepository) WithContext(ctx context.Context) SchedulerRepositoryInterface {
	// 返回副本而不是原地改 db：仓储在容器里是单例，原地改会让并发请求互相覆盖上下文
	cloned := *i
	cloned.db = i.db.WithContext(ctx)
	cloned.ctx = ctx
	return &cloned
}

func (i *SchedulerRepository) RetrieveJobs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []schedulerpkg.Job, err error) {
	tx := i.db.Model(&schedulerpkg.Job{})
	if fn != nil {
		fn(tx)
	}
	if err = tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("source asc, id asc").Find(&list).Error
	return count, list, err
}

func (i *SchedulerRepository) FindJob(id uint) (*schedulerpkg.Job, error) {
	row := new(schedulerpkg.Job)
	if err := i.db.Where("id = ?", id).First(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (i *SchedulerRepository) CreateJob(job *schedulerpkg.Job) error {
	return i.db.Create(job).Error
}

func (i *SchedulerRepository) SaveJob(job *schedulerpkg.Job) error {
	return i.db.Save(job).Error
}

func (i *SchedulerRepository) UpdateJobFields(id uint, updates map[string]any) error {
	return i.db.Model(&schedulerpkg.Job{}).Where("id = ?", id).Updates(updates).Error
}

func (i *SchedulerRepository) DeleteJob(id uint) error {
	return i.db.Where("id = ?", id).Delete(&schedulerpkg.Job{}).Error
}

func (i *SchedulerRepository) RetrieveRuns(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []schedulerpkg.Run, err error) {
	tx := i.db.Model(&schedulerpkg.Run{})
	if fn != nil {
		fn(tx)
	}
	if err = tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("id desc").Find(&list).Error
	return count, list, err
}

func (i *SchedulerRepository) FindRun(id uint) (*schedulerpkg.Run, error) {
	row := new(schedulerpkg.Run)
	if err := i.db.Where("id = ?", id).First(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (i *SchedulerRepository) CountActiveRuns(jobID uint) (int64, error) {
	var count int64
	err := i.db.Model(&schedulerpkg.Run{}).
		Where("job_id = ? AND status IN ?", jobID, schedulerpkg.ActiveRunStatuses()).
		Count(&count).Error
	return count, err
}

func (i *SchedulerRepository) DeleteTerminalRunsBefore(cutoff time.Time) (int64, error) {
	res := i.db.
		Where("status IN ? AND finished_at IS NOT NULL AND finished_at < ?", schedulerpkg.TerminalRunStatuses(), cutoff).
		Delete(&schedulerpkg.Run{})
	return res.RowsAffected, res.Error
}

func (i *SchedulerRepository) CountJobsByStatus() (map[int8]int64, error) {
	type statusCount struct {
		Status int8  `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	var rows []statusCount
	err := i.db.Model(&schedulerpkg.Job{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int8]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (i *SchedulerRepository) CountRunsByStatusSince(since time.Time) (map[int8]int64, error) {
	type statusCount struct {
		Status int8  `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	var rows []statusCount
	err := i.db.Model(&schedulerpkg.Run{}).
		Select("status, count(*) as count").
		Where("created_at >= ?", since).
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int8]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (i *SchedulerRepository) RecentFailedRuns(limit int) ([]schedulerpkg.Run, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []schedulerpkg.Run
	err := i.db.
		Where("status IN ?", []int8{schedulerpkg.RunStatusFailed, schedulerpkg.RunStatusTimeout}).
		Order("id desc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
