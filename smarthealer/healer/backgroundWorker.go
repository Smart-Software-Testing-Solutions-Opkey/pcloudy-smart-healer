package healer

import (
	"context"
	"encoding/json"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/filelog"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type BackgroundWorker struct {
	intelSystem intelligence.IntelligenceSystem
	uowFactory  *store.UnitOfWorkFactory

	desSem  *semaphore.Weighted
	healSem *semaphore.Weighted
}

func NewBGWorker(
	intel intelligence.IntelligenceSystem,
	uowF *store.UnitOfWorkFactory,
) (*BackgroundWorker, error) {
	ctx := context.Background()
	u, err := uowF.NewUnitOfWork(ctx)
	if err != nil {
		return nil, err
	}

	dLen, err := u.DescriptionQueue.Length(ctx)
	if err != nil {
		return nil, err
	}
	hlen, err := u.HealingQueue.Length(ctx)
	if err != nil {
		return nil, err
	}
	u.Rollback()

	filelog.Info("BackgroundWorker initializing: description_queue=%d, healing_queue=%d", dLen, hlen)

	// Semaphores need a max capacity, but start at 0 available tokens
	// We'll use a large max capacity and acquire all tokens initially
	const maxCapacity = 100000
	desSem := semaphore.NewWeighted(maxCapacity)
	healSem := semaphore.NewWeighted(maxCapacity)

	// Acquire all tokens initially (start with 0 available)
	desSem.Acquire(context.Background(), maxCapacity)
	healSem.Acquire(context.Background(), maxCapacity)

	// Release tokens for existing queue entries
	if dLen > 0 {
		desSem.Release(dLen)
		filelog.Info("Released %d tokens for description queue", dLen)
	}
	if hlen > 0 {
		healSem.Release(hlen)
		filelog.Info("Released %d tokens for healing queue", hlen)
	}

	return &BackgroundWorker{
		intelSystem: intel,
		uowFactory:  uowF,
		desSem:      desSem,
		healSem:     healSem,
	}, nil
}

const (
	workCount  = 1
	rateBurst  = 1 // Burst must be at least 1 to allow any requests
)

func (b *BackgroundWorker) NotifyDescriptionPosted() {
	b.desSem.Release(workCount)
}

func (b *BackgroundWorker) ProcessDescriptionsBG(ctx context.Context, limit rate.Limit) {
	limiter := rate.NewLimiter(limit, rateBurst)

	filelog.Info("Description worker loop starting, waiting for work...")

	for {
		filelog.Info("Description worker attempting to acquire semaphore...")
		if err := b.desSem.Acquire(ctx, workCount); err != nil {
			// ctx done was trigered
			// exit out of loop
			filelog.Info("Description worker semaphore acquire failed: %v", err)
			return
		}

		filelog.Info("Description worker acquired work token, processing...")
		b.processWork(ctx, limiter, func() error {
			return b.descriptionWork(ctx)
		})
	}
}

func (b *BackgroundWorker) NotifyHealingPosted() {
	b.healSem.Release(workCount)
}

func (b *BackgroundWorker) ProcessHealingBG(ctx context.Context, limit rate.Limit, healWork func(context.Context) error) {
	limiter := rate.NewLimiter(limit, rateBurst)

	for {
		if err := b.healSem.Acquire(ctx, workCount); err != nil {
			// ctx done was trigered
			// exit out of loop
			return
		}

		b.processWork(ctx, limiter, func() error {
			return healWork(ctx)
		})
	}
}

func (b *BackgroundWorker) HealWorkerFunc(
	resolver func(context.Context, LocatorInfo, ResolveOptions, *store.UnitOfWork) (string, error),
) func(context.Context) error {
	return func(ctx context.Context) error {
		// a work was acquired
		u, err := b.uowFactory.NewUnitOfWork(ctx)
		if err != nil {
			filelog.Error("Healing worker failed to create unit of work: %v", err)
			return err
		}

		committed := false
		defer func() {
			if !committed {
				u.Rollback()
			}
		}()

		e, err := u.HealingQueue.GetOldestEntry(ctx)
		if err != nil {
			filelog.Error("Healing worker failed to get oldest entry: %v", err)
			return err
		}

		var info LocatorInfo
		if err := json.Unmarshal([]byte(e.InfoJson), &info); err != nil {
			filelog.Error("Healing worker failed to unmarshal info JSON: queue_id=%d, error=%v", e.Id, err)
			return err
		}

		var opts ResolveOptions
		if err := json.Unmarshal([]byte(e.OptJson), &opts); err != nil {
			filelog.Error("Healing worker failed to unmarshal options JSON: queue_id=%d, error=%v", e.Id, err)
			return err
		}

		filelog.Info("Healing worker started: queue_id=%d, project_id=%s, xpath=%s", e.Id, info.ProjectId, info.XPath)

		healedLocator, err := resolver(ctx, info, opts, u)
		if err != nil {
			filelog.Error("Healing worker failed to resolve locator: queue_id=%d, xpath=%s, error=%v", e.Id, info.XPath, err)
			return err
		}

		filelog.Info("Healing worker generated new locator: queue_id=%d, original_xpath=%s, healed_locator=%s", e.Id, info.XPath, healedLocator)

		if err := u.HealingQueue.Remove(ctx, e.Id); err != nil {
			filelog.Error("Healing worker failed to remove from queue: queue_id=%d, error=%v", e.Id, err)
			return err
		}

		if err := u.Commit(); err != nil {
			filelog.Error("Healing worker failed to commit: queue_id=%d, error=%v", e.Id, err)
			return err
		}
		committed = true

		filelog.Info("Healing worker completed: queue_id=%d", e.Id)

		return nil
	}
}

func (b *BackgroundWorker) descriptionWork(ctx context.Context) error {
	// a work was acquired
	u, err := b.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		filelog.Error("Description worker failed to create unit of work: %v", err)
		return err
	}

	committed := false
	defer func() {
		if !committed {
			u.Rollback()
		}
	}()

	e, err := u.DescriptionQueue.GetOldestEntry(ctx)
	if err != nil {
		filelog.Error("Description worker failed to get oldest entry: %v", err)
		return err
	}

	filelog.Info("Description worker started: locator_id=%d, page_id=%d", e.LocatorId, e.PageId)

	if err := b.generateLocatorDescription(ctx, e.LocatorId, e.PageId, u); err != nil {
		filelog.Error("Description worker failed to generate description: locator_id=%d, page_id=%d, error=%v", e.LocatorId, e.PageId, err)
		return err
	}

	if err := u.DescriptionQueue.Remove(ctx, e.LocatorId, e.PageId); err != nil {
		filelog.Error("Description worker failed to remove from queue: locator_id=%d, page_id=%d, error=%v", e.LocatorId, e.PageId, err)
		return err
	}

	if err := u.Commit(); err != nil {
		filelog.Error("Description worker failed to commit: locator_id=%d, page_id=%d, error=%v", e.LocatorId, e.PageId, err)
		return err
	}
	committed = true

	filelog.Info("Description worker completed: locator_id=%d, page_id=%d", e.LocatorId, e.PageId)

	return nil
}

func (b *BackgroundWorker) processWork(ctx context.Context, limiter *rate.Limiter, work func() error) {
	completed := true
	defer func() {
		if !completed {
			filelog.Warn("Work not completed, releasing semaphore token back")
			b.desSem.Release(workCount)
		}
	}()

	filelog.Info("Waiting for rate limiter...")
	if err := limiter.Wait(ctx); err != nil {
		filelog.Error("Rate limiter wait failed: %v", err)
		completed = false
		return
	}

	filelog.Info("Rate limiter passed, executing work...")
	if err := work(); err != nil {
		filelog.Error("Work function failed: %v", err)
		completed = false
		return
	}

	filelog.Info("Work completed successfully")
}

func (b *BackgroundWorker) generateLocatorDescription(ctx context.Context, locatorId, pageId int, u *store.UnitOfWork) error {
	newLocator, err := u.Locators.GetLocator(ctx, locatorId)
	if err != nil {
		return err
	}

	pageSrcInfo, err := u.Pages.GetPageSourceInfo(ctx, pageId)
	if err != nil {
		return err
	}

	page, err := page.NewPage(pageSrcInfo.PageSource, pageSrcInfo.PageType)
	if err != nil {
		return err
	}

	elemSrc, err := page.GetElementSrc(newLocator)
	if err != nil {
		return err
	}

	desc, err := b.intelSystem.GenerateElementDescription(ctx, pageSrcInfo.PageSource, elemSrc)
	if err != nil {
		return err
	}

	filelog.Info("Generated description for locator_id=%d: xpath=%s, description=%s", locatorId, newLocator, desc)

	return u.Locators.UpdateDescription(ctx, locatorId, desc)
}
