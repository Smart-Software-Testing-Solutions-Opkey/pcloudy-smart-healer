package smarthealer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type BackgroundWorker struct {
	cfg         config.Config
	intelSystem intelligence.IntelligenceSystem
	uowFactory  store.UnitOfWorkFactory

	desSem  *semaphore.Weighted
	healSem *semaphore.Weighted
}

func NewBGWorker(
	cfg config.Config,
	intel intelligence.IntelligenceSystem,
	uowF store.UnitOfWorkFactory,
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

	return &BackgroundWorker{
		cfg:         cfg,
		intelSystem: intel,
		uowFactory:  uowF,
		desSem:      semaphore.NewWeighted(dLen),
		healSem:     semaphore.NewWeighted(hlen),
	}, nil
}

const workCount = 1

func (b *BackgroundWorker) NotifyDescriptionPosted() {
	b.desSem.Release(workCount)
}

func (b *BackgroundWorker) ProcessDescriptionsBG(ctx context.Context, limit rate.Limit) {

	limiter := rate.NewLimiter(limit, 0)

	for {
		if err := b.desSem.Acquire(ctx, workCount); err != nil {
			// ctx done was trigered
			// exit out of loop
			return
		}

		b.processWork(ctx, limiter, func() error {
			return b.descriptionWork(ctx)
		})
	}
}

func (b *BackgroundWorker) NotifyHealingPosted() {
	b.healSem.Release(workCount)
}

func (b *BackgroundWorker) ProcessHealingBG(ctx context.Context, limit rate.Limit, healWork func(context.Context) error) {
	limiter := rate.NewLimiter(limit, 0)

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
			return err
		}
		defer func() {
			rollBackErr := u.Rollback()
			if err != nil {
				err = fmt.Errorf("%w: %w", err, rollBackErr)
			} else {
				err = rollBackErr
			}
		}()

		e, err := u.HealingQueue.GetOldestEntry(ctx)
		if err != nil {
			return err
		}

		var info LocatorInfo
		if err := json.Unmarshal([]byte(e.InfoJson), &info); err != nil {
			return err
		}

		var opts ResolveOptions
		if err := json.Unmarshal([]byte(e.OptJson), &opts); err != nil {
			return err
		}

		_, err = resolver(ctx, info, opts, u)
		if err != nil {
			return err
		}

		return err
	}
}

func (b *BackgroundWorker) descriptionWork(ctx context.Context) error {
	// a work was acquired
	u, err := b.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = u.Rollback()
	}()

	e, err := u.DescriptionQueue.GetOldestEntry(ctx)
	if err != nil {
		return err
	}

	if err := b.generateLocatorDescription(ctx, e.LocatorId, e.PageId, u); err != nil {
		return err
	}

	return err
}

func (b *BackgroundWorker) processWork(ctx context.Context, limiter *rate.Limiter, work func() error) {
	completed := true
	defer func() {
		if !completed {
			b.desSem.Release(workCount)
		}
	}()

	if err := limiter.Wait(ctx); err != nil {
		completed = false
		return
	}

	if err := work(); err != nil {
		completed = false
		return
	}
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

	return u.Locators.UpdateDescription(ctx, locatorId, desc)
}
