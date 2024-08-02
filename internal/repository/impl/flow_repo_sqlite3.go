package impl

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/yeqown/log"
	"gorm.io/driver/sqlite"
	gorm2 "gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yeqown/gitlab-flow/internal/repository"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// isDatabaseLocked check err is sqlite3.ErrBusy or not.
func isDatabaseLocked(err error) bool {
	v, ok := err.(sqlite3.Error)
	if !ok {
		return false
	}

	return v.Code == sqlite3.ErrBusy
}

type sqliteFlowRepositoryImpl struct {
	// connectFunc provide the way to connect database.MasterBranch
	// also helps resolve "database is locked".
	connectFunc func() *gorm2.DB

	db *gorm2.DB

	// txCounter atomic.Value
}

func ConnectDB(path string, debug bool) func() *gorm2.DB {
	dbName := "gitlab-flow.db"
	init := false
	// if debug {
	//	dbName = "gitlab-flow.debug.db"
	// }
	path = filepath.Join(path, dbName)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			init = true
		} else {
			log.Fatal(err)
			panic("could not reach")
		}
	}

	return func() *gorm2.DB {
		cfg := gorm2.Config{}
		db, err := gorm2.Open(sqlite.Open(path), &cfg)
		if err != nil {
			log.Fatalf("loading db failed: %v", err)
		}

		log.
			WithFields(log.Fields{
				"path": path,
				"init": init,
			}).
			Debug("ConnectDB() called")

		// @yeqown auto migrate enabled in each connection.
		if err = db.AutoMigrate(
			&repository.ProjectDO{},
			&repository.MilestoneDO{},
			&repository.BranchDO{},
			&repository.IssueDO{},
			&repository.MergeRequestDO{},
		); err != nil {
			log.Warnf("auto migrate database failed: %v", err)
		}

		// db logger SetLogLevel
		db.Logger = db.Logger.LogMode(logger.Silent)
		if debug {
			db.Logger = db.Logger.LogMode(logger.Info)
		}

		// DONE(@yeqown): init or load database file.
		return db
	}
}

// NewBasedSqlite3
// DONE(@yeqown): load or create sqlite3 database.
func NewBasedSqlite3(connectFunc func() *gorm2.DB) repository.IFlowRepository {
	repo := sqliteFlowRepositoryImpl{
		connectFunc: connectFunc,
		db:          connectFunc(),
	}

	return &repo
}

func (repo *sqliteFlowRepositoryImpl) reinit() {
	if repo.db != nil {
		sqlDB, err := repo.db.DB()
		if err != nil {
			log.Debugf("get sql.DB failed: %v", err)
		}
		if err = sqlDB.Close(); err != nil {
			log.Debugf("sqlDB.Close() failed: %v", err)
		}
	}

	repo.db = repo.connectFunc()
}

// txIn get tx from txs, if txs is nil we return repo.db instead of tx.
// txs must be got from repo.StartTransaction().
func (repo *sqliteFlowRepositoryImpl) txIn(txs ...*gorm2.DB) (tx *gorm2.DB) {
	if len(txs) != 0 {
		tx = txs[0]
	}

	return
}

func (repo *sqliteFlowRepositoryImpl) StartTransaction() *gorm2.DB {
	return repo.db.Begin()
}

func (repo *sqliteFlowRepositoryImpl) CommitTransaction(tx *gorm2.DB) error {
	return tx.Commit().Error
}

func (repo *sqliteFlowRepositoryImpl) SaveProject(m *repository.ProjectDO, txs ...*gorm2.DB) (err error) {
	return repo.insertRecordWithCheck(repo.txIn(txs...), m)
}

func (repo *sqliteFlowRepositoryImpl) QueryProject(filter *repository.ProjectDO) (*repository.ProjectDO, error) {
	out := new(repository.ProjectDO)
	err := repo.db.
		Model(out).
		Order("created_at DESC").
		Where(filter).
		First(out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (repo *sqliteFlowRepositoryImpl) QueryProjects(filter *repository.ProjectDO) ([]*repository.ProjectDO, error) {
	out := make([]*repository.ProjectDO, 0, 10)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (repo *sqliteFlowRepositoryImpl) SaveMilestone(m *repository.MilestoneDO, txs ...*gorm2.DB) (err error) {
	return repo.insertRecordWithCheck(repo.txIn(txs...), m)
}

func (repo *sqliteFlowRepositoryImpl) QueryMilestone(filter *repository.MilestoneDO) (*repository.MilestoneDO, error) {
	out := new(repository.MilestoneDO)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		First(out).Error
	if err != nil {
		return nil, err
	}

	return out, err
}

func (repo *sqliteFlowRepositoryImpl) QueryMilestones(
	filter *repository.MilestoneDO) ([]*repository.MilestoneDO, error) {

	out := make([]*repository.MilestoneDO, 0, 10)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (repo *sqliteFlowRepositoryImpl) QueryMilestoneByBranchName(projectId int, branchName string,
) (*repository.MilestoneDO, error) {
	branch, err := repo.QueryBranch(&repository.BranchDO{
		ProjectID:  projectId,
		BranchName: branchName,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "could not locate branch:"+branchName)
	}

	milestone, err := repo.QueryMilestone(&repository.MilestoneDO{MilestoneID: branch.MilestoneID})
	return milestone, err
}

func (repo *sqliteFlowRepositoryImpl) SaveBranch(m *repository.BranchDO, txs ...*gorm2.DB) (err error) {
	return repo.insertRecordWithCheck(repo.txIn(txs...), m)
}

func (repo *sqliteFlowRepositoryImpl) BatchCreateBranch(records []*repository.BranchDO, txs ...*gorm2.DB) error {
	// uniq records with local database.
	uniq := make([]*repository.BranchDO, 0, len(records))
	for idx, v := range records {
		count := int64(0)
		err := repo.db.Model(v).Where(v).Count(&count).Error
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		uniq = append(uniq, records[idx])
	}

	// no need to create new records.
	if len(uniq) == 0 {
		return nil
	}

	return repo.batchCreate(uniq, len(uniq), txs...)
}

func (repo *sqliteFlowRepositoryImpl) QueryBranch(filter *repository.BranchDO) (*repository.BranchDO, error) {
	out := new(repository.BranchDO)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		First(out).Error

	return out, err
}

func (repo *sqliteFlowRepositoryImpl) QueryBranches(filter *repository.BranchDO) ([]*repository.BranchDO, error) {
	out := make([]*repository.BranchDO, 0, 10)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (repo *sqliteFlowRepositoryImpl) SaveIssue(m *repository.IssueDO, txs ...*gorm2.DB) (err error) {
	return repo.insertRecordWithCheck(repo.txIn(txs...), m)
}

func (repo *sqliteFlowRepositoryImpl) BatchCreateIssue(records []*repository.IssueDO, txs ...*gorm2.DB) error {
	// uniq records with local database.
	uniq := make([]*repository.IssueDO, 0, len(records))
	for idx, v := range records {
		count := int64(0)
		err := repo.db.Model(v).Where(v).Count(&count).Error
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		uniq = append(uniq, records[idx])
	}

	// no need to create new records.
	if len(uniq) == 0 {
		return nil
	}

	return repo.batchCreate(uniq, len(uniq), txs...)
}

func (repo *sqliteFlowRepositoryImpl) QueryIssue(filter *repository.IssueDO) (*repository.IssueDO, error) {
	out := new(repository.IssueDO)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		First(out).Error
	if err != nil {
		return nil, err
	}

	return out, err
}

func (repo *sqliteFlowRepositoryImpl) QueryIssues(filter *repository.IssueDO) ([]*repository.IssueDO, error) {
	out := make([]*repository.IssueDO, 0, 10)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	return out, err
}

func (repo *sqliteFlowRepositoryImpl) SaveMergeRequest(m *repository.MergeRequestDO, txs ...*gorm2.DB) error {
	return repo.insertRecordWithCheck(repo.txIn(txs...), m)
}

func (repo *sqliteFlowRepositoryImpl) BatchCreateMergeRequest(records []*repository.MergeRequestDO, txs ...*gorm2.DB) error {
	// uniq records with local database.
	uniq := make([]*repository.MergeRequestDO, 0, len(records))
	for idx, v := range records {
		count := int64(0)
		err := repo.db.Model(v).Where(v).Count(&count).Error
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		uniq = append(uniq, records[idx])
	}

	// no need to create new records.
	if len(uniq) == 0 {
		return nil
	}

	return repo.batchCreate(uniq, len(uniq), txs...)
}

func (repo *sqliteFlowRepositoryImpl) QueryMergeRequest(
	filter *repository.MergeRequestDO) (*repository.MergeRequestDO, error) {
	out := new(repository.MergeRequestDO)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		First(out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (repo *sqliteFlowRepositoryImpl) QueryMergeRequests(
	filter *repository.MergeRequestDO) ([]*repository.MergeRequestDO, error) {

	out := make([]*repository.MergeRequestDO, 0, 10)
	err := repo.db.
		Model(filter).
		Order("created_at DESC").
		Where(filter).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	return out, nil
}

// insertRecordWithCheck would insert data and checking data is exists or not.
// If data has been exists, function would return directly, otherwise function would
// insert into database. Externally, it would retry when insert got err `database is locked`.
// The way to retry is reopening after closing DB  with backoff algorithm.
// FIXED: resolve "Database is locked"
func (repo *sqliteFlowRepositoryImpl) insertRecordWithCheck(tx *gorm2.DB, m interface{}) (err error) {
	isTransaction := true
	if tx == nil {
		tx = repo.db
		isTransaction = false
	}

	insertFunc := func() (err error) {
		defer func() {
			if isDatabaseLocked(err) {
				repo.reinit()
			}
		}()

		count := int64(0)
		if err = tx.Model(m).Where(m).Count(&count).Error; err != nil {
			// database error
			log.
				WithFields(log.Fields{
					"filter": m, "count": count,
				}).
				Warnf("create recheck failed: %v", err)
			return err
		}

		if count > 0 {
			// has exists
			return nil
		}

		// do not exists, then create it.
		err = tx.Model(m).Create(m).Error
		if err != nil && isTransaction {
			_ = tx.Rollback()
		}
		if err != nil {
			log.
				WithFields(log.Fields{
					"err":           err,
					"data":          m,
					"isTransaction": isTransaction,
				}).
				Debugf("insertRecordWithCheck")
		}

		return err
	}

	// backoff retry
	backoffPolicy := backoff.NewExponentialBackOff()
	backoffPolicy.MaxElapsedTime = 45 * time.Second
	if err = backoff.Retry(insertFunc, backoffPolicy); err != nil {
		log.Errorf("insertRecordWithCheck FAILED finally, err=%v", err)
		if isDatabaseLocked(err) {
			log.Error("still database is locked")
		}

		_ = tx.Rollback()
		return err
	}

	return nil
}

func (repo *sqliteFlowRepositoryImpl) batchCreate(value interface{}, size int, txs ...*gorm2.DB) error {
	tx := repo.txIn(txs...)
	if tx == nil {
		tx = repo.db
	}

	if err := tx.CreateInBatches(value, size).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (repo *sqliteFlowRepositoryImpl) RemoveProjectAndRelatedData(projectId int) (err error) {
	if projectId <= 0 {
		return nil
	}

	tx := repo.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	proj := new(repository.ProjectDO)
	if err = tx.First(proj, &repository.ProjectDO{ProjectID: projectId}).Error; err != nil {
		if errors.Is(err, gorm2.ErrRecordNotFound) {
			return nil
		}

		return errors.Wrap(err, "could not locate project")
	}

	// remove project
	delCondition := &repository.ProjectDO{ProjectID: projectId}
	if err = tx.Unscoped().Delete(&repository.ProjectDO{}, delCondition).Error; err != nil {
		return err
	}
	// remove milestones
	delCondition2 := &repository.MilestoneDO{ProjectID: projectId}
	if err = tx.Unscoped().Delete(&repository.MilestoneDO{}, delCondition2).Error; err != nil {
		return err
	}
	// remove branches
	delCondition3 := &repository.BranchDO{ProjectID: projectId}
	if err = tx.Unscoped().Delete(&repository.BranchDO{}, delCondition3).Error; err != nil {
		return err
	}
	// remove issues
	delCondition4 := &repository.IssueDO{ProjectID: projectId}
	if err = tx.Unscoped().Delete(&repository.IssueDO{}, delCondition4).Error; err != nil {
		return err
	}
	// remove merge requests
	delCondition5 := &repository.MergeRequestDO{ProjectID: projectId}
	if err = tx.Unscoped().Delete(&repository.MergeRequestDO{}, delCondition5).Error; err != nil {
		return err
	}

	return nil
}
