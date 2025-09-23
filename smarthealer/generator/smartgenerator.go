package generator

import "context"

type SmartGenerator interface {
	GenerateElementDescription(ctx context.Context, root, elem string) (string, error)
	GenerateLocator(ctx context.Context, desc, root string) (string, error)
}
