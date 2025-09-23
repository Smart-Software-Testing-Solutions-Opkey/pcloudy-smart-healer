package retrieval

import (
	"context"
	"errors"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
)

type ComparisionMode int

const (
	AutomaticComparisionMode ComparisionMode = iota
	ManualComparisionMode
	ScreenshotComparisionMode
)

var (
	ErrRetrievalFailed = errors.New("failed to retrieve candidate pages")
	ErrNoSimilarPage   = errors.New("failed to find any similar page")
)

type PageRetriever struct {
	pageStore store.PageStore
	intel     intelligence.IntelligenceSystem
}

func NewPageRetriever(pageStore store.PageStore) *PageRetriever {
	return &PageRetriever{
		pageStore: pageStore,
	}
}

type RetereivalOptions struct {
	ContextId string
	B64Png    string
	ProjectId string
	Locator   string
	Platform  platform.Platform
}

func (p *PageRetriever) RetrieveCandidatePages(ctx context.Context, opt RetereivalOptions, mode ComparisionMode) (int, error) {
	if mode == AutomaticComparisionMode {
		switch opt.Platform {
		case platform.IosPlatform:
			mode = ScreenshotComparisionMode
		case platform.AndroidPlatform, platform.WebPlatform:
			mode = ManualComparisionMode
		default:
			return -1, fmt.Errorf("invalid platform specified: %w", ErrRetrievalFailed)
		}
	}

	switch mode {
	case AutomaticComparisionMode:
		panic("RetrieveCandidatePages invalid state")
	case ManualComparisionMode:
		r, err := p.usingContextID(ctx, opt)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrRetrievalFailed)
		}
		return r, err
	case ScreenshotComparisionMode:
		r, err := p.usingScreenShot(ctx, opt)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrRetrievalFailed)
		}
		return r, err
	default:
		return -1, fmt.Errorf("invalid comparision mode provided: %w", ErrRetrievalFailed)
	}
}

func (p *PageRetriever) usingContextID(ctx context.Context, opt RetereivalOptions) (int, error) {
	r, err := p.pageStore.GetFirstPageWithContext(ctx, opt.ProjectId, opt.Locator, opt.ContextId)
	if err != nil {
		if errors.Is(err, store.ErrEmptyData) {
			return -1, ErrNoSimilarPage
		}
		return -1, err
	}

	return r, nil
}

func (p *PageRetriever) usingScreenShot(ctx context.Context, opt RetereivalOptions) (int, error) {
	potentialPages, err := p.pageStore.GetPages(ctx, opt.ProjectId, opt.Locator)
	if err != nil {
		return -1, err
	}

	// early return if we have only one page to bypass expensive screenshot comparision
	// todo: is this valid ?
	if len(potentialPages) == 1 {
		return potentialPages[0], nil
	}

	for _, pp := range potentialPages {
		storedPng, err := p.pageStore.GetPagePNG(ctx, pp)
		if err != nil {
			// todo: report to user that there was an error here somehow
			continue
		}

		ok, err := p.compareSS(ctx, storedPng, opt.B64Png)
		if err != nil {
			// todo: report to user that there was an error here somehow
			continue
		}

		if ok {
			return pp, nil
		}
	}

	return -1, ErrNoSimilarPage
}

func (p *PageRetriever) compareSS(ctx context.Context, img1, img2 string) (bool, error) {
	return p.intel.CompareScreenShot(ctx, img1, img2)
}
