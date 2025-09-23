package smarthealer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/generator"
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
	smartGen      generator.SmartGenerator
	pageStore     store.PageStore
	locatorStore  store.LocatorStore
	pageRetriever retrieval.PageRetriever
}

func NewHealer(cfg config.Config) *Healer {
	return &Healer{
		cfg: cfg,
	}
}

func (h *Healer) ResolveLocator(ctx context.Context, info LocatorInfo, opt ResolveOptions) (string, error) {
	conformLocatorInfo(&info)

	// is it a new entry
	ok, err := h.isNewEntry(ctx, info)
	if err != nil {
		return "", errors.Join(ErrResolveFailed, err)
	}

	if ok {
		r, err := h.handleNewEntry(ctx, info)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
		}
		return r, err
	}

	r, err := h.handleExistingEntry(ctx, info, opt)
	if err != nil {
		err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
	}
	return r, err
}

func (h *Healer) handleNewEntry(ctx context.Context, info LocatorInfo) (string, error) {
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

	entry, err := h.registerPageAndLocator(ctx, info)
	if err != nil {
		return "", err
	}

	if err := h.generateLocatorDescription(ctx, entry.LocatorId, entry.PageId); err != nil {
		return "", err
	}

	return info.XPath, nil
}

func (h *Healer) handleExistingEntry(ctx context.Context, info LocatorInfo, opt ResolveOptions) (string, error) {
	page, err := page.NewPage(info.PageSource, info.PageType)
	if err != nil {
		return "", err
	}

	// candidate locators are from db
	locators, err := h.getCandidateLocators(ctx, info, opt.ComparisionMode)
	if err != nil {
		return "", err
	}

	// check if any of the locators we have
	// stored works
	for _, locator := range locators.Locators {
		ok, err := page.IsValidXPath(locator)
		if err != nil {
			// this is critical error, we expect all the xpath to be valid xpaths
			// todo: add some sort of way to inform we ran into critical error
			continue
		}
		if ok {
			return locator, nil
		}
	}

	// non of the stored locators work

	// generate a new locator
	locator, err := h.generateLocator(ctx, info, locators.PageId)
	if err != nil {
		return "", err
	}

	// add the locator to the database
	if err := h.registerLocator(ctx, locator, locators.PageId); err != nil {
		return "", err
	}

	return locator, nil
}

type Candidate struct {
	PageId   int
	Locators []string
}

func (h *Healer) getCandidateLocators(ctx context.Context, info LocatorInfo, mode retrieval.ComparisionMode) (*Candidate, error) {
	//? should it be candidate page or candidate pages
	//? most commonly it should only be candidate page and not pages
	//? needs further verification
	candidatePage, err := h.pageRetriever.RetrieveCandidatePages(ctx, retrieval.RetereivalOptions{
		ContextId: info.ContextId,
		B64Png:    info.B64Png,
		ProjectId: info.ProjectId,
		Locator:   info.XPath,
		Platform:  info.Platform,
	}, mode)
	if err != nil {
		return nil, err
	}

	locators, err := h.locatorStore.GetPageLocators(ctx, candidatePage)
	if err != nil {
		return nil, err
	}

	return &Candidate{
		PageId:   candidatePage,
		Locators: locators,
	}, nil
}

func (h *Healer) registerLocator(ctx context.Context, locator string, pageId int) error {
	locatorId, err := h.locatorStore.Add(ctx,
		store.LocatorEntry{
			PageId:      pageId,
			Locator:     locator,
			Description: "",
		})
	if err != nil {
		return err
	}

	if err := h.generateLocatorDescription(ctx, locatorId, pageId); err != nil {
		return err
	}

	return nil
}

func (h *Healer) generateLocator(ctx context.Context, info LocatorInfo, pageId int) (string, error) {
	desc, err := h.locatorStore.GetLatestPageDescription(ctx, pageId)
	if err != nil {
		return "", err
	}

	return h.smartGen.GenerateLocator(ctx, desc, info.PageSource)
}

func (h *Healer) isNewEntry(ctx context.Context, info LocatorInfo) (bool, error) {
	return h.pageStore.CheckPage(ctx, info.ProjectId, info.XPath, info.ContextId)
}

type EntryId struct {
	PageId    int
	LocatorId int
}

func (h *Healer) registerPageAndLocator(ctx context.Context, info LocatorInfo) (*EntryId, error) {
	pageId, err := h.pageStore.Add(ctx,
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

	locatorId, err := h.locatorStore.Add(ctx,
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

func (h *Healer) generateLocatorDescription(ctx context.Context, locatorId, pageId int) error {
	newLocator, err := h.locatorStore.GetLocator(ctx, locatorId)
	if err != nil {
		return err
	}

	pageSrcInfo, err := h.pageStore.GetPageSourceInfo(ctx, pageId)
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

	desc, err := h.smartGen.GenerateElementDescription(ctx, pageSrcInfo.PageSource, elemSrc)
	if err != nil {
		return err
	}

	return h.locatorStore.UpdateDescription(ctx, locatorId, desc)
}

func conformLocatorInfo(info *LocatorInfo) {
	const defaultContextId = "DEFAULT_CONTEXT_ID"

	if strings.TrimSpace(info.ContextId) == "" {
		info.ContextId = defaultContextId
	}
}
