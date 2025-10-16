package healer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/filelog"
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
	intelSys      intelligence.IntelligenceSystem
	pageRetriever *retrieval.PageRetriever
	uowFactory    *store.UnitOfWorkFactory
	bg            *BackgroundWorker
}

func NewHealer(
	intel intelligence.IntelligenceSystem,
	pageR *retrieval.PageRetriever,
	uowF *store.UnitOfWorkFactory,
	bg *BackgroundWorker,
) *Healer {
	return &Healer{
		intelSys:      intel,
		pageRetriever: pageR,
		uowFactory:    uowF,
		bg:            bg,
	}
}

func (h *Healer) ResolveLocator(ctx context.Context, info LocatorInfo, opt ResolveOptions, u *store.UnitOfWork) (string, error) {
	conformLocatorInfo(&info)

	var err error
	var ownsTransaction bool
	if u == nil {
		u, err = h.uowFactory.NewUnitOfWork(ctx)
		if err != nil {
			return "", err
		}
		ownsTransaction = true
		defer u.Rollback()
	}

	candidate, err := h.getCandidateLocators(ctx, info, opt.ComparisionMode, u)
	if err != nil {
		return "", fmt.Errorf("ResolveLocator: %w", err)
	}

	var r string
	if len(candidate.Pages) == 0 {
		// No similar pages found - this is a NEW element/page
		r, err = h.handleNewEntry(ctx, info, u)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
			return r, err
		}
	} else {
		// Found similar pages with candidate locators - trying to HEAL
		r, err = h.handleExistingEntry(ctx, info, candidate, u)
		if err != nil {
			err = fmt.Errorf("%w: %w", err, ErrResolveFailed)
			return r, err
		}
	}

	// Commit and notify only if we own the transaction
	// If transaction is passed in, caller is responsible for commit
	if ownsTransaction {
		if err := u.Commit(); err != nil {
			return "", err
		}
		h.bg.NotifyDescriptionPosted()
	}

	return r, nil
}

func (h *Healer) handleNewEntry(ctx context.Context, info LocatorInfo, u *store.UnitOfWork) (string, error) {
	page, err := page.NewPage(info.PageSource, info.PageType)
	if err != nil {
		return "", fmt.Errorf("handleNewEntry: failed to create page from source: %w", err)
	}

	ok, err := page.IsValidXPath(info.XPath)
	if err != nil {
		return "", fmt.Errorf("handleNewEntry: failed to validate xpath: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("handleNewEntry: provided xpath doesn't exist in source (xpath=%s)", info.XPath)
	}

	entry, err := h.registerPageAndLocator(ctx, info, u)
	if err != nil {
		return "", fmt.Errorf("handleNewEntry: failed to register page and locator: %w", err)
	}

	if err := h.generateLocatorDescription(ctx, entry.LocatorId, entry.PageId, u); err != nil {
		return "", fmt.Errorf("handleNewEntry: failed to queue description generation: %w", err)
	}

	return info.XPath, nil
}

func (h *Healer) handleExistingEntry(ctx context.Context, info LocatorInfo, candidate *Candidate, u *store.UnitOfWork) (string, error) {
	currentPage, err := page.NewPage(info.PageSource, info.PageType)
	if err != nil {
		return "", fmt.Errorf("handleExistingEntry: failed to create page from source: %w", err)
	}

	// Try locators from all candidate pages
	for _, candidatePage := range candidate.Pages {
		for _, locator := range candidatePage.Locators {
			ok, err := currentPage.IsValidXPath(locator)
			if err != nil {
				// ! this is critical error, we expect all the xpath to be valid xpaths
				filelog.Error("CRITICAL: Invalid XPath found in database - locator: %s, page_id: %d, project_id: %s, error: %v",
					locator, candidatePage.PageId, info.ProjectId, err)
				continue
			}
			if ok {
				filelog.Info("Found matching locator from page_id=%d: %s", candidatePage.PageId, locator)
				return locator, nil
			}
		}
	}

	// None of the stored locators work - need to generate a new one
	filelog.Info("No existing locators work, generating new healed locator...")

	// Use the first candidate page for AI generation context
	firstPageId := candidate.Pages[0].PageId
	locator, err := h.generateLocator(ctx, info, firstPageId, currentPage, u)
	if err != nil {
		return "", fmt.Errorf("handleExistingEntry: failed to generate new locator using AI: %w", err)
	}

	filelog.Info("Generated healed locator: %s, saving as new page variant", locator)

	// Register as a NEW page variant
	// IMPORTANT: Store the ORIGINAL locator (info.XPath) in the page for future retrieval
	// Store the HEALED locator separately so it can be returned when the original is sent again
	entry, err := h.registerPageAndLocatorWithHealed(ctx, info, locator, u)
	if err != nil {
		return "", fmt.Errorf("handleExistingEntry: failed to register healed page and locator: %w", err)
	}

	// Queue description generation for the new healed locator
	if err := h.generateLocatorDescription(ctx, entry.LocatorId, entry.PageId, u); err != nil {
		filelog.Error("Failed to queue description for healed locator: %v", err)
		// Don't fail the whole operation if description queueing fails
	}

	return locator, nil
}

type Candidate struct {
	Pages []CandidatePage
}

type CandidatePage struct {
	PageId   int
	Locators []string
}

func (h *Healer) getCandidateLocators(ctx context.Context, info LocatorInfo, mode retrieval.ComparisionMode, u *store.UnitOfWork) (*Candidate, error) {
	candidatePageIds, err := h.pageRetriever.RetrieveCandidatePages(ctx, retrieval.RetereivalOptions{
		ContextId: info.ContextId,
		B64Png:    info.B64Png,
		ProjectId: info.ProjectId,
		Locator:   info.XPath,
		Platform:  info.Platform,
	}, mode)
	if err != nil {
		if errors.Is(err, retrieval.ErrNoSimilarPage) {
			// No similar pages found in database - this is a new element/page
			return &Candidate{
				Pages: []CandidatePage{},
			}, nil
		}

		return nil, fmt.Errorf("getCandidateLocators: failed to retrieve candidate pages: %w", err)
	}

	// Get locators for each candidate page
	var pages []CandidatePage
	for _, pageId := range candidatePageIds {
		locators, err := u.Locators.GetPageLocators(ctx, pageId)
		if err != nil {
			filelog.Error("Failed to get locators for page_id %d: %v", pageId, err)
			continue
		}

		pages = append(pages, CandidatePage{
			PageId:   pageId,
			Locators: locators,
		})
	}

	return &Candidate{
		Pages: pages,
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
	return h.registerPageAndLocatorWithHealed(ctx, info, info.XPath, u)
}

func (h *Healer) registerPageAndLocatorWithHealed(ctx context.Context, info LocatorInfo, healedLocator string, u *store.UnitOfWork) (*EntryId, error) {
	pageId, err := u.Pages.Add(ctx,
		store.PageEntry{
			PageSource: info.PageSource,
			Locator:    info.XPath, // Store ORIGINAL locator for retrieval
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
			Locator:     healedLocator, // Store HEALED locator as the working solution
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
	return u.DescriptionQueue.Add(ctx, locatorId, pageId)
}

func conformLocatorInfo(info *LocatorInfo) {
	const defaultContextId = "DEFAULT_CONTEXT_ID"

	const pngPrefix = "data:image/png;base64,"

	if strings.TrimSpace(info.ContextId) == "" {
		info.ContextId = defaultContextId
	}

	// Only add prefix if B64Png is not empty and doesn't already have the prefix
	if info.B64Png != "" && !strings.HasPrefix(info.B64Png, pngPrefix) {
		info.B64Png = fmt.Sprintf("%s%s", pngPrefix, info.B64Png)
	}
}
