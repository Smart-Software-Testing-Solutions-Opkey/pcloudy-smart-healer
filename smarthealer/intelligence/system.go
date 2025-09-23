package intelligence

import "context"

type IntelligenceSystem interface {
	GenerateElementDescription(ctx context.Context, root, elem string) (string, error)
	GenerateLocator(ctx context.Context, desc, root string) (string, error)

	CompareScreenShot(ctx context.Context, img1, img2 string) (bool, error)
}
