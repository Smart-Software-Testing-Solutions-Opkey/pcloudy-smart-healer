package page

import (
	"errors"
)

type PageType int

const (
	XMLPageType PageType = iota
	HTMLPageType
)

const (
	xmlString     = "XMLPageType"
	htmlString    = "HTMLPageType"
	invalidString = "Invalid-PageType"
)

func (p PageType) String() string {
	switch p {
	case XMLPageType:
		return xmlString
	case HTMLPageType:
		return htmlString
	default:
		return invalidString
	}
}

func NewPageTypeFromString(s string) PageType {
	switch s {
	case xmlString:
		return XMLPageType
	case htmlString:
		return HTMLPageType
	default:
		return PageType(-1)
	}
}

var ErrInvalidPageType = errors.New("invalid page type provided")

func (p PageType) MarshalJSON() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *PageType) UnmarshalJSON(b []byte) error {
	*p = NewPageTypeFromString(string(b))

	return nil
}
