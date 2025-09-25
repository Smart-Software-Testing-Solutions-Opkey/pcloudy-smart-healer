package smarthealer

import (
	"context"
	"time"

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

	desSym *semaphore.Weighted
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

	len, err := u.DescriptionQueue.Length(ctx)
	if err != nil {
		return nil, err
	}
	u.Rollback()

	return &BackgroundWorker{
		cfg:         cfg,
		intelSystem: intel,
		uowFactory:  uowF,
		desSym:      semaphore.NewWeighted(len),
	}, nil
}

func (b *BackgroundWorker) NotifyDescriptionPosted() {
	b.desSym.Release(1)
}

const workCount = 1

func (b *BackgroundWorker) ProcessDescriptionsBG(ctx context.Context) {

	limit := rate.Every(1 * time.Second)

	limiter := rate.NewLimiter(limit, 0)

	for {
		if err := b.desSym.Acquire(ctx, workCount); err != nil {
			// ctx done was trigered
			// exit out of loop
			return
		}

		b.processDescription(ctx, limiter)
	}
}

func (b *BackgroundWorker) processDescription(ctx context.Context, limiter *rate.Limiter) {
	completed := true
	defer func() {
		if !completed {
			b.desSym.Release(workCount)
		}
	}()

	if err := limiter.Wait(ctx); err != nil {
		completed = false
		return
	}

	// a work was acquired
	u, err := b.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		completed = false
		return
	}
	defer func() {
		if err := u.Rollback(); err != nil {
			completed = false
		}
	}()

	e, err := u.DescriptionQueue.GetOldestEntry(ctx)
	if err != nil {
		completed = false
		return
	}

	if err := b.generateLocatorDescription(ctx, e.LocatorId, e.PageId, u); err != nil {
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
