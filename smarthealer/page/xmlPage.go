package page

import (
	"errors"
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
		return false, errors.Join(ErrInvalidXPath, err)
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
	return "", errors.ErrUnsupported
}
