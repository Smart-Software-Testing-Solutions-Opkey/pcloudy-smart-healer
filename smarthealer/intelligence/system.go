package intelligence

import (
	"context"
	"errors"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
)

var (
	ErrDescriptionGenerationFailed = errors.New("failed to generate element description")
	ErrSSComparisionFailed         = errors.New("failed to compare screenshots")
	ErrNewLocatorGenerationFailed  = errors.New("failed to generate a new locator")
)

type IntelligenceSystem interface {
	GenerateElementDescription(ctx context.Context, root, elem string) (string, error)
	GenerateLocator(ctx context.Context, desc string, root page.Page, platform platform.Platform) (string, error)

	CompareScreenShot(ctx context.Context, img1, img2 string) (bool, error)
}
