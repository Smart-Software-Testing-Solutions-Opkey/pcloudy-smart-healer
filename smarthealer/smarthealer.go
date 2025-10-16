package smarthealer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/filelog"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/healer"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/llm"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/retrieval"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store/sqliteimpl"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/time/rate"
)

type SmartHealer struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	uofW   *store.UnitOfWorkFactory
	healer *healer.Healer
	bg     *healer.BackgroundWorker
}

func NewSmartHealer(cfg config.Config) (*SmartHealer, error) {
	dbPath, err := getDbPath(cfg.Db.Path)
	if err != nil {
		return nil, err
	}

	// Initialize file logging to the same directory as the database
	if err := filelog.Init(dbPath); err != nil {
		return nil, err
	}

	if err := store.EnsureMigrations(dbPath); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	uowF := store.NewUnitOfWorkFactory(db, store.FactoryParams{
		PageStoreFactory: func(tx *sqlx.Tx) store.PageStore {
			return sqliteimpl.NewSqlitePageStore(tx)
		},
		LocatorStoreFactory: func(tx *sqlx.Tx) store.LocatorStore {
			return sqliteimpl.NewSqliteLocatorStore(tx)
		},
		DescriptionQueueFactory: func(tx *sqlx.Tx) store.DescriptionQueue {
			return sqliteimpl.NewSqliteDescriptionQueueStore(tx)
		},
		HealingQueueFactory: func(tx *sqlx.Tx) store.HealingQueue {
			return sqliteimpl.NewSqliteHealingQueueStore(tx)
		},
	})

	l := llm.NewOpenAILLM(cfg.Ai.SecretKey)
	i := intelligence.NewLLmIntelSystem(l, cfg.Ai.SecretKey)

	p := retrieval.NewPageRetriever(uowF, i)

	bg, err := healer.NewBGWorker(i, uowF)
	if err != nil {
		return nil, err
	}

	h := healer.NewHealer(i, p, uowF, bg)

	ctx, cancel := context.WithCancel(context.Background())

	return &SmartHealer{
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},
		uofW:   uowF,
		healer: h,
		bg:     bg,
	}, nil
}

func (s *SmartHealer) Close() {
	s.cancel()

	s.wg.Wait()

	// Close file log
	filelog.Close()
}

func (s *SmartHealer) StartBackgroundWorkers() {
	limit := rate.Every(2 * time.Second)

	filelog.Info("Starting background workers with rate limit: 1 request per 2 seconds")

	s.wg.Go(func() {
		filelog.Info("Description background worker started")
		s.bg.ProcessDescriptionsBG(s.ctx, limit)
		filelog.Info("Description background worker stopped")
	})

	s.wg.Go(func() {
		filelog.Info("Healing background worker started")
		s.bg.ProcessHealingBG(
			s.ctx,
			limit,
			s.bg.HealWorkerFunc(s.healer.ResolveLocator),
		)
		filelog.Info("Healing background worker stopped")
	})
}

func (s *SmartHealer) ResolveLocator(
	info healer.LocatorInfo,
	opt healer.ResolveOptions,
) (string, error) {
	return s.healer.ResolveLocator(s.ctx, info, opt, nil)
}

func (s *SmartHealer) ResolveLocatorAsync(
	info healer.LocatorInfo,
	opt healer.ResolveOptions,
) error {
	u, err := s.uofW.NewUnitOfWork(s.ctx)
	if err != nil {
		return err
	}
	defer u.Rollback()

	infoJson, err := json.Marshal(info)
	if err != nil {
		return err
	}
	optJson, err := json.Marshal(opt)
	if err != nil {
		return err
	}

	if err := u.HealingQueue.Add(s.ctx, string(infoJson), string(optJson)); err != nil {
		return err
	}

	if err := u.Commit(); err != nil {
		return err
	}

	s.bg.NotifyHealingPosted()

	return nil
}

func getDbPath(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dirPath := filepath.Join(home, ".smarthealer")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dirPath, "smarthealer.db"), nil
}
