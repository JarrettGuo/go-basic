package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64, version int) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{
		db: db,
	}
}

func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id =?", id).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).Updates(map[string]any{
		"next_time": next.UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"status": JobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) Release(ctx context.Context, id int64, version int) error {
	now := time.Now().UnixMilli()
	// 释放任务，检查任务ID、状态和版本号是否匹配
	result := g.db.WithContext(ctx).Model(&Job{}).Where("id = ? AND status = ? AND version = ?",
		id, JobStatusRunning, version).Updates(map[string]any{
		"status":  JobStatusWaiting,
		"utime":   now,
		"version": gorm.Expr("version + 1"),
	})
	if result.Error != nil {
		return result.Error
	}
	// 检查是否有行被更新，如果没有，说明任务不存在或状态/版本不匹配
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Preempt 抢占任务
func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := g.db.WithContext(ctx).Model(&Job{})
	for {
		now := time.Now().UnixMilli()
		var j Job
		// 查询下一个需要执行的任务
		err := db.Where("status = ? AND next_time <= ?", JobStatusWaiting, now).First(&j).Error
		if err != nil {
			return Job{}, err
		}
		// 抢占任务
		res := db.Where("id = ? AND version = ?", j.Id, j.Version).Updates(map[string]any{
			"status":  JobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 说明任务已经被抢占, 继续下一轮
			continue
		}
		return j, nil
	}
}

type Job struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 用于乐观锁，防止并发问题
	Version int
	Cfg     string
	// 用状态来标志哪些任务可以抢，那些任务已经被抢，哪些任务永远不会执行
	Status int
	Cron   string
	Name   string `gorm:"unique"`
	// 定时任务，下一次被调度的时间
	NextTime int64 `gorm:"index"`
	Executor string
	Ctime    int64
	Utime    int64
}

const (
	JobStatusWaiting = iota
	// 任务被抢占
	JobStatusRunning
	// 任务暂停调度
	JobStatusPaused
)
