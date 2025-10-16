package page

import (
	"errors"
	"strings"
)

type Page interface {
	IsValidXPath(xpath string) (bool, error)
	String() string
	PageType() PageType

	GetElementSrc(xpath string) (string, error)
}

var ErrInvalidXPath = errors.New("invalid xpath provided")

func NewPage(src string, pageType PageType) (Page, error) {
	srcReader := strings.NewReader(src)

	switch pageType {
	case XMLPageType:
		return NewXMLPage(srcReader)
	case HTMLPageType:
		return NewHTMLPage(srcReader)
	default:
		return nil, ErrInvalidPageType
	}
}
