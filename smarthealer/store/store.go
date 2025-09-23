package store

import (
	"context"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
)

type PageStore interface {
	Add(ctx context.Context, entry PageEntry) (int, error)

	CheckPage(ctx context.Context, projectId, locator, contextId string) (bool, error)

	GetPageSourceInfo(ctx context.Context, pageId int) (PageSrcInfo, error)
}

type LocatorStore interface {
	Add(ctx context.Context, entry LocatorEntry) (int, error)

	UpdateDescription(ctx context.Context, locatorId int, desc string) error

	GetPageLocators(ctx context.Context, pageId int) ([]string, error)
	GetLatestPageDescription(ctx context.Context, pageId int) (string, error)
	GetLocator(ctx context.Context, locatorId int) (string, error)
}

type PageEntry struct {
	PageSource string
	Locator    string
	B64Png     string
	ContextId  string
	ProjectId  string
	Platform   platform.Platform
	PageType   page.PageType
}

type LocatorEntry struct {
	PageId      int
	Locator     string
	Description string
}

type PageSrcInfo struct {
	PageSource string
	PageType   page.PageType
}
