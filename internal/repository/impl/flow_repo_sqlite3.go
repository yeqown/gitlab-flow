package impl

import (
	"errors"
	"math/rand"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/yeqown/gitlab-flow/internal/repository"
	"github.com/yeqown/log"
	gorm2 "gorm.io/gorm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// isDatabaseLocked 检查是不是数据库 busy [并发写导致写操作失败]
func isDatabaseLocked(err error) bool {
	v, ok := err.(sqlite3.Error)
	if !ok {
		return false
	}

	return v.Code == sqlite3.ErrBusy
}

type sqliteFlowRepositoryImpl struct {
	db *gorm2.DB
}

func NewBasedSqlite3(db *gorm2.DB) repository.IFlowRepository {
	return sqliteFlowRepositoryImpl{
		db: db,
	}
}

// 保存项目记录
func (repo sqliteFlowRepositoryImpl) SaveProject(m *repository.ProjectDO) (err error) {
	return repo.insertRecordWithCheck(m, new(repository.ProjectDO))
}

// 根据项目名查询项目
func (repo sqliteFlowRepositoryImpl) QueryProject(filter *repository.ProjectDO) (*repository.ProjectDO, error) {
	out := new(repository.ProjectDO)
	if err := repo.db.Model(out).
		Where(filter).First(out).Error; err != nil {
		return nil, err
	}

	return out, nil
}

func (repo sqliteFlowRepositoryImpl) SaveMilestone(m *repository.MilestoneDO) (err error) {
	return repo.insertRecordWithCheck(m, new(repository.MilestoneDO))
}

func (repo sqliteFlowRepositoryImpl) QueryMilestone(filter *repository.MilestoneDO) (*repository.MilestoneDO, error) {
	out := new(repository.MilestoneDO)
	err := repo.db.Model(filter).Where(filter).First(out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) QueryMilestones(
	filter *repository.MilestoneDO) ([]*repository.MilestoneDO, error) {

	out := make([]*repository.MilestoneDO, 0, 10)
	err := repo.db.Model(filter).Where(filter).Find(&out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) SaveBranch(m *repository.BranchDO) (err error) {
	return repo.insertRecordWithCheck(m, new(repository.BranchDO))
}

func (repo sqliteFlowRepositoryImpl) QueryBranch(filter *repository.BranchDO) (*repository.BranchDO, error) {
	out := new(repository.BranchDO)
	err := repo.db.Model(filter).Where(filter).First(out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) SaveIssue(m *repository.IssueDO) (err error) {
	return repo.insertRecordWithCheck(m, new(repository.IssueDO))
}

func (repo sqliteFlowRepositoryImpl) QueryIssue(filter *repository.IssueDO) (*repository.IssueDO, error) {
	out := new(repository.IssueDO)
	err := repo.db.Model(filter).Where(filter).First(out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) QueryIssues(filter *repository.IssueDO) ([]*repository.IssueDO, error) {
	out := make([]*repository.IssueDO, 0, 10)
	err := repo.db.Model(filter).Where(filter).Find(&out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) SaveMergeRequest(m *repository.MergeRequestDO) error {
	return repo.insertRecordWithCheck(m, new(repository.MergeRequestDO))
}

func (repo sqliteFlowRepositoryImpl) QueryMergeRequest(
	filter *repository.MergeRequestDO) (*repository.MergeRequestDO, error) {
	out := new(repository.MergeRequestDO)
	err := repo.db.Model(filter).Where(filter).First(out).Error
	return out, err
}

func (repo sqliteFlowRepositoryImpl) QueryMergeRequests(
	filter *repository.MergeRequestDO) ([]*repository.MergeRequestDO, error) {

	out := make([]*repository.MergeRequestDO, 0, 10)
	err := repo.db.Model(filter).Where(filter).Find(&out).Error
	return out, err
}

const _maxRetryTimes = 5

// FIXME: resolve "Database is locked"
// Solution1: 重试机制，随机回退避让，不能保证100%解决问题，并发量大时候也会有问题
func (repo sqliteFlowRepositoryImpl) insertRecordWithCheck(m interface{}, out interface{}) (err error) {
	insertFunc := func() error {
		tx := repo.db.Begin()
		if err = tx.Model(m).Where(m).First(out).Error; err == nil {
			// 已经存在则直接返回
			return nil
		}

		// 数据库异常
		if !errors.Is(err, gorm2.ErrRecordNotFound) {
			log.WithFields(log.Fields{"filter": m, "out": out, "error": err}).Warn("数据检查异常")
			return err
		}

		// 不存在则创建
		if err = tx.Model(m).Create(m).Error; err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit().Error
	}

	var retry int32 // 重试计数
	for err = insertFunc(); err != nil && isDatabaseLocked(err) && retry < _maxRetryTimes; retry++ {
		// 如果因为 "database is locked" 则重试，最多重试 _maxRetryTimes 次
		log.WithFields(log.Fields{"err": err}).Warnf("Database is locked, retrying....%d", retry)

		// 随机时间回避
		backoff := rand.Int31() % 2000
		if backoff < 0 {
			backoff -= 0
		}
		time.Sleep(time.Duration(backoff)*time.Millisecond + time.Duration(retry/2)*time.Second)
	}

	return err
}

// ClearMilestoneAndRelated 清理跟里程碑相关联
func (repo sqliteFlowRepositoryImpl) ClearMilestoneAndRelated(milestoneID int) error {
	if milestoneID == 0 {
		log.Warn("milestoneID = 0")
		return nil
	}
	tx := repo.db.Begin()

	// 删除里程碑
	filter0 := &repository.MilestoneDO{MilestoneID: milestoneID}
	result := tx.Where(filter0).Delete(repository.MilestoneDO{})
	log.WithFields(log.Fields{
		"error":    result.Error,
		"affected": result.RowsAffected,
	}).Debugf("删除里程碑")
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 删除分支
	filter1 := repository.BranchDO{MilestoneID: milestoneID}
	result = tx.Where(filter1).Delete(repository.BranchDO{})
	log.WithFields(log.Fields{
		"error":    result.Error,
		"affected": result.RowsAffected,
	}).Debugf("删除分支")
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 删除 ISSUE
	filter2 := repository.IssueDO{MilestoneID: milestoneID}
	result = tx.Where(filter2).Delete(repository.IssueDO{})
	log.WithFields(log.Fields{
		"error":    result.Error,
		"affected": result.RowsAffected,
	}).Debugf("删除Issue")
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 删除 MR
	filter3 := repository.MergeRequestDO{MilestoneID: milestoneID}
	result = tx.Where(filter3).Delete(repository.MergeRequestDO{})
	log.WithFields(log.Fields{
		"error":    result.Error,
		"affected": result.RowsAffected,
	}).Debugf("删除MR")
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	tx.Commit()

	return nil
}

// SaveSyncFeaturesData
// 一次事务中插入，本次同步所获取到的数据, 插入过程中会忽略已经存在的数据
// 批量需要拆分
func (repo sqliteFlowRepositoryImpl) SaveSyncFeaturesData(
	m *repository.MilestoneDO,
	b []*repository.BranchDO,
	i []*repository.IssueDO,
	mrs []*repository.MergeRequestDO,
) (err error) {
	tx := repo.db.Begin()

	// milestone
	milestoneRecords := make([]*repository.MilestoneDO, 0, 10)
	if _, err = repo.QueryMilestone(m); err != nil {
		if !errors.Is(err, gorm2.ErrRecordNotFound) {
			return err
		}
		milestoneRecords = append(milestoneRecords, m)
	}

	if err = tx.CreateInBatches(milestoneRecords, len(milestoneRecords)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// b
	branchRecords := make([]*repository.BranchDO, 0, 10)
	for idx, v := range b {
		if _, err = repo.QueryBranch(v); err != nil {
			if !errors.Is(err, gorm2.ErrRecordNotFound) {
				return err
			}
			branchRecords = append(branchRecords, b[idx])
		}
	}
	if err = tx.CreateInBatches(branchRecords, len(branchRecords)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// i
	issueRecords := make([]*repository.IssueDO, 0, 10)
	for idx, v := range i {
		if _, err := repo.QueryIssue(v); err != nil {
			if !errors.Is(err, gorm2.ErrRecordNotFound) {
				return err
			}
			issueRecords = append(issueRecords, i[idx])
		}
	}
	if err = tx.CreateInBatches(issueRecords, len(issueRecords)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// mrs
	mrRecords := make([]*repository.MergeRequestDO, 0, 10)
	for idx, v := range mrs {
		if _, err := repo.QueryMergeRequest(v); err != nil {
			if !errors.Is(err, gorm2.ErrRecordNotFound) {
				return err
			}
			mrRecords = append(mrRecords, mrs[idx])
		}
	}
	if err = tx.CreateInBatches(mrRecords, len(mrRecords)).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
