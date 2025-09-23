package page

import (
	"errors"
	"fmt"
	"io"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type htmlPage struct {
	pageSrc string
	root    *html.Node
}

func NewHTMLPage(r io.Reader) (*htmlPage, error) {
	html, err := htmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML page: %w", err)
	}

	return &htmlPage{
		pageSrc: htmlquery.OutputHTML(html, true),
		root:    html,
	}, nil
}

func (p *htmlPage) IsValidXPath(xpath string) (bool, error) {
	elem, err := htmlquery.Query(p.root, xpath)
	if err != nil {
		return false, errors.Join(ErrInvalidXPath, err)
	}
	return elem != nil, nil
}

func (p *htmlPage) String() string {
	return p.pageSrc
}

func (p *htmlPage) PageType() PageType {
	return HTMLPageType
}

func (p *htmlPage) GetElementSrc(xpath string) (string, error) {
	return "", errors.ErrUnsupported
}
