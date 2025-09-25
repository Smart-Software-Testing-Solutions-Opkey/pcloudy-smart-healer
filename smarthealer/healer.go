package smarthealer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/retrieval"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
)

var ErrResolveFailed = errors.New("failed to resolve locator")

type LocatorInfo struct {
	ProjectId  string
	PageSource string
	B64Png     string
	XPath      string
	ContextId  string
	Platform   platform.Platform
	PageType   page.PageType
}

type ResolveOptions struct {
	ComparisionMode retrieval.ComparisionMode
}

type Healer struct {
	cfg           config.Config
	intelSys      intelligence.IntelligenceSystem
	pageRetriever retrieval.PageRetriever
	uowFactory    store.UnitOfWorkFactory
}

func NewHealer(cfg config.Config) *Healer {
	return &Healer{
		cfg: cfg,
	}
}

func (h *Healer) ResolveLocator(ctx context.Context, info LocatorInfo, opt ResolveOptions) (string, error) {
	conformLocatorInfo(&info)

	u, err := h.uowFactory.NewUnitOfWork(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		err = u.Rollback()
	}()

	candidate, err := h.getCandidateLocators(ctx, info, opt.ComparisionMode, u)
	if err != nil {
		return "", err
	}

	if len(candidate.Locators) == 0 {
		r, err := h.handleNewEntry(ctx, info, u)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
		}
		return r, err
	}

	r, err := h.handleExistingEntry(ctx, info, candidate, u)
	if err != nil {
		err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
	}

	if err := u.Commit(); err != nil {
		return "", err
	}

	return r, err
}

func (h *Healer) handleNewEntry(ctx context.Context, info LocatorInfo, u *store.UnitOfWork) (string, error) {
	page, err := page.NewPage(info.PageSource, info.PageType)
	if err != nil {
		return "", err
	}

	ok, err := page.IsValidXPath(info.XPath)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("provided xpath doesn't exist in source")
	}

	entry, err := h.registerPageAndLocator(ctx, info, u)
	if err != nil {
		return "", err
	}

	if err := h.generateLocatorDescription(ctx, entry.LocatorId, entry.PageId, u); err != nil {
		return "", err
	}

	return info.XPath, nil
}

func (h *Healer) handleExistingEntry(ctx context.Context, info LocatorInfo, candidate *Candidate, u *store.UnitOfWork) (string, error) {
	page, err := page.NewPage(info.PageSource, info.PageType)
	if err != nil {
		return "", err
	}

	// check if any of the locators we have
	// stored works
	for _, locator := range candidate.Locators {
		ok, err := page.IsValidXPath(locator)
		if err != nil {
			// ! this is critical error, we expect all the xpath to be valid xpaths
			// todo: add some sort of way to inform we ran into critical error
			continue
		}
		if ok {
			return locator, nil
		}
	}

	// non of the stored locators work

	// generate a new locator
	locator, err := h.generateLocator(ctx, info, candidate.PageId, page, u)
	if err != nil {
		return "", err
	}

	// add the locator to the database
	if err := h.registerLocator(ctx, locator, candidate.PageId, u); err != nil {
		return "", err
	}

	return locator, nil
}

type Candidate struct {
	PageId   int
	Locators []string
}

func (h *Healer) getCandidateLocators(ctx context.Context, info LocatorInfo, mode retrieval.ComparisionMode, u *store.UnitOfWork) (*Candidate, error) {
	candidatePage, err := h.pageRetriever.RetrieveCandidatePages(ctx, retrieval.RetereivalOptions{
		ContextId: info.ContextId,
		B64Png:    info.B64Png,
		ProjectId: info.ProjectId,
		Locator:   info.XPath,
		Platform:  info.Platform,
	}, mode)
	if err != nil {
		if errors.Is(err, retrieval.ErrNoSimilarPage) {
			return &Candidate{
				PageId:   -1,
				Locators: []string{},
			}, nil
		}

		return nil, err
	}

	locators, err := u.Locators.GetPageLocators(ctx, candidatePage)
	if err != nil {
		return nil, err
	}

	return &Candidate{
		PageId:   candidatePage,
		Locators: locators,
	}, nil
}

func (h *Healer) registerLocator(ctx context.Context, locator string, pageId int, u *store.UnitOfWork) error {
	locatorId, err := u.Locators.Add(ctx,
		store.LocatorEntry{
			PageId:      pageId,
			Locator:     locator,
			Description: "",
		})
	if err != nil {
		return err
	}

	if err := h.generateLocatorDescription(ctx, locatorId, pageId, u); err != nil {
		return err
	}

	return nil
}

func (h *Healer) generateLocator(ctx context.Context, info LocatorInfo, pageId int, page page.Page, u *store.UnitOfWork) (string, error) {
	desc, err := u.Locators.GetLatestPageDescription(ctx, pageId)
	if err != nil {
		return "", err
	}

	return h.intelSys.GenerateLocator(ctx, desc, page, info.Platform)
}

type EntryId struct {
	PageId    int
	LocatorId int
}

func (h *Healer) registerPageAndLocator(ctx context.Context, info LocatorInfo, u *store.UnitOfWork) (*EntryId, error) {
	pageId, err := u.Pages.Add(ctx,
		store.PageEntry{
			PageSource: info.PageSource,
			Locator:    info.XPath,
			B64Png:     info.B64Png,
			ContextId:  info.ContextId,
			ProjectId:  info.ProjectId,
			Platform:   info.Platform,
			PageType:   info.PageType,
		})
	if err != nil {
		return nil, err
	}

	locatorId, err := u.Locators.Add(ctx,
		store.LocatorEntry{
			PageId:      pageId,
			Locator:     info.XPath,
			Description: "",
		})
	if err != nil {
		return nil, err
	}

	return &EntryId{
		PageId:    pageId,
		LocatorId: locatorId,
	}, nil
}

func (h *Healer) generateLocatorDescription(ctx context.Context, locatorId, pageId int, u *store.UnitOfWork) error {
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

	desc, err := h.intelSys.GenerateElementDescription(ctx, pageSrcInfo.PageSource, elemSrc)
	if err != nil {
		return err
	}

	return u.Locators.UpdateDescription(ctx, locatorId, desc)
}

func conformLocatorInfo(info *LocatorInfo) {
	const defaultContextId = "DEFAULT_CONTEXT_ID"

	const pngPrefix = "data:image/png;base64,"

	if strings.TrimSpace(info.ContextId) == "" {
		info.ContextId = defaultContextId
	}

	if !strings.HasPrefix(info.B64Png, pngPrefix) {
		info.B64Png = fmt.Sprintf("%s%s", pngPrefix, info.B64Png)
	}
}
