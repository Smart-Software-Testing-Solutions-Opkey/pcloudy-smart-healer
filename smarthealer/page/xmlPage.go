package page

import (
	"fmt"
	"io"

	"github.com/antchfx/xmlquery"
)

type xmlPage struct {
	pageSrc string
	root    *xmlquery.Node
}

func NewXMLPage(r io.Reader) (*xmlPage, error) {
	xml, err := xmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML page: %w", err)
	}

	return &xmlPage{
		pageSrc: xml.OutputXML(true),
		root:    xml,
	}, nil
}

func (p *xmlPage) IsValidXPath(xpath string) (bool, error) {
	elem, err := xmlquery.Query(p.root, xpath)
	if err != nil {
		return false, fmt.Errorf("%w: %w", err, ErrInvalidXPath)
	}
	return elem != nil, nil
}

func (p *xmlPage) String() string {
	return p.pageSrc
}

func (p *xmlPage) PageType() PageType {
	return XMLPageType
}

func (p *xmlPage) GetElementSrc(xpath string) (string, error) {
	n, err := xmlquery.Query(p.root, xpath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrInvalidXPath)
	}

	if n == nil {
		return "", fmt.Errorf("xpath does not match any element: %s: %w", xpath, ErrInvalidXPath)
	}

	return n.OutputXML(true), nil
}
