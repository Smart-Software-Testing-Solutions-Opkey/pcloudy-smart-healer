package page

import "errors"

type PageType int

const (
	XMLPageType PageType = iota
	HTMLPageType
)

func (p PageType) String() string {
	switch p {
	case XMLPageType:
		return "XMLPageType"
	case HTMLPageType:
		return "HTMLPageType"
	default:
		return "Invalid-PageType"
	}
}

var ErrInvalidPageType = errors.New("invalid page type provided")
