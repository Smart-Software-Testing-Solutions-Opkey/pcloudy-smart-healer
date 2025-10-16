package retrieval

import (
	"context"
	"errors"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/filelog"
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

const (
	automaticCmpStr  = "Automatic"
	manualCmpStr     = "Manual"
	screenshotCmpStr = "Screenshot"
	invalidCmpStr    = "invalid-comparision"
)

func (c ComparisionMode) String() string {
	switch c {
	case AutomaticComparisionMode:
		return automaticCmpStr
	case ManualComparisionMode:
		return manualCmpStr
	case ScreenshotComparisionMode:
		return screenshotCmpStr
	default:
		return invalidCmpStr
	}
}

func NewComparisionModeFromString(s string) ComparisionMode {
	switch s {
	case automaticCmpStr:
		return AutomaticComparisionMode
	case manualCmpStr:
		return ManualComparisionMode
	case screenshotCmpStr:
		return ScreenshotComparisionMode
	default:
		return ComparisionMode(-1)
	}
}

func (c ComparisionMode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.String() + `"`), nil
}

func (c *ComparisionMode) UnmarshalJSON(b []byte) error {
	// Remove quotes if present
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	*c = NewComparisionModeFromString(s)
	return nil
}

var (
	ErrRetrievalFailed = errors.New("failed to retrieve candidate pages")
	ErrNoSimilarPage   = errors.New("failed to find any similar page")
)

type PageRetriever struct {
	uowF  *store.UnitOfWorkFactory
	intel intelligence.IntelligenceSystem
}

func NewPageRetriever(uowF *store.UnitOfWorkFactory, intel intelligence.IntelligenceSystem) *PageRetriever {
	return &PageRetriever{
		uowF:  uowF,
		intel: intel,
	}
}

type RetereivalOptions struct {
	ContextId string
	B64Png    string
	ProjectId string
	Locator   string
	Platform  platform.Platform
}

func (p *PageRetriever) RetrieveCandidatePages(ctx context.Context, opt RetereivalOptions, mode ComparisionMode) ([]int, error) {
	if mode == AutomaticComparisionMode {
		switch opt.Platform {
		case platform.IosPlatform:
			mode = ScreenshotComparisionMode
		case platform.AndroidPlatform, platform.WebPlatform:
			mode = ManualComparisionMode
		default:
			return nil, fmt.Errorf("invalid platform specified: %w", ErrRetrievalFailed)
		}
	}

	u, err := p.uowF.NewUnitOfWork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit of work: %w", ErrRetrievalFailed)
	}
	defer u.Rollback() // this is readonly transaction but needs cleanup

	switch mode {
	case AutomaticComparisionMode:
		panic("RetrieveCandidatePages invalid state")
	case ManualComparisionMode:
		r, err := p.usingContextID(ctx, opt, u)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrRetrievalFailed)
		}
		return r, err
	case ScreenshotComparisionMode:
		r, err := p.usingScreenShot(ctx, opt, u)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrRetrievalFailed)
		}
		return r, err
	default:
		return nil, fmt.Errorf("invalid comparision mode provided: %w", ErrRetrievalFailed)
	}
}

func (p *PageRetriever) usingContextID(ctx context.Context, opt RetereivalOptions, u *store.UnitOfWork) ([]int, error) {
	r, err := u.Pages.GetAllPagesWithContext(ctx, opt.ProjectId, opt.Locator, opt.ContextId)
	if err != nil {
		if errors.Is(err, store.ErrEmptyData) {
			return nil, ErrNoSimilarPage
		}
		return nil, err
	}

	return r, nil
}

func (p *PageRetriever) usingScreenShot(ctx context.Context, opt RetereivalOptions, u *store.UnitOfWork) ([]int, error) {
	potentialPages, err := u.Pages.GetPages(ctx, opt.ProjectId, opt.Locator)
	if err != nil {
		return nil, err
	}

	// early return if we have only one page to bypass expensive screenshot comparison
	if len(potentialPages) == 1 {
		return potentialPages, nil
	}

	var matchingPages []int
	for _, pp := range potentialPages {
		storedPng, err := u.Pages.GetPagePNG(ctx, pp)
		if err != nil {
			filelog.Error("Failed to get page PNG for page_id %d during screenshot comparison: %v", pp, err)
			continue
		}

		ok, err := p.compareSS(ctx, storedPng, opt.B64Png)
		if err != nil {
			filelog.Error("Failed to compare screenshots for page_id %d: %v", pp, err)
			continue
		}

		if ok {
			matchingPages = append(matchingPages, pp)
		}
	}

	if len(matchingPages) == 0 {
		return nil, ErrNoSimilarPage
	}

	return matchingPages, nil
}

func (p *PageRetriever) compareSS(ctx context.Context, img1, img2 string) (bool, error) {
	return p.intel.CompareScreenShot(ctx, img1, img2)
}
