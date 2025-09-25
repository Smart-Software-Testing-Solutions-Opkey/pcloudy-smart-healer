package page

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

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
	n, err := htmlquery.Query(p.root, xpath)
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrInvalidXPath)
	}

	return htmlquery.OutputHTML(n, true), nil
}

func ConvertCssSelectorToXpath(selector string) string {
	xpath := selector

	// --- Basic selectors ---
	// Universal selector: * → *
	xpath = strings.ReplaceAll(xpath, "*", "*")

	// IDs: #id → [@id='id']
	reID := regexp.MustCompile(`#([a-zA-Z0-9\-_]+)`)
	xpath = reID.ReplaceAllString(xpath, "[@id='$1']")

	// Classes: .class → contains(@class,'class')
	reClass := regexp.MustCompile(`\.([a-zA-Z0-9\-_]+)`)
	xpath = reClass.ReplaceAllString(xpath,
		"[contains(concat(' ', normalize-space(@class), ' '), ' $1 ')]")

	// Attributes
	reAttrEq := regexp.MustCompile(`\[([a-zA-Z0-9\-_]+)=['"]?([^\]'"]+)['"]?\]`)
	xpath = reAttrEq.ReplaceAllString(xpath, "[@$1='$2']")

	reAttrStarts := regexp.MustCompile(`\[([a-zA-Z0-9\-_]+)\^=['"]?([^\]'"]+)['"]?\]`)
	xpath = reAttrStarts.ReplaceAllString(xpath, "[starts-with(@$1,'$2')]")

	reAttrEnds := regexp.MustCompile(`\[([a-zA-Z0-9\-_]+)\$=['"]?([^\]'"]+)['"]?\]`)
	xpath = reAttrEnds.ReplaceAllString(xpath, "[substring(@$1,string-length(@$1)-string-length('$2')+1)='$2']")

	reAttrContains := regexp.MustCompile(`\[([a-zA-Z0-9\-_]+)\*=['"]?([^\]'"]+)['"]?\]`)
	xpath = reAttrContains.ReplaceAllString(xpath, "[contains(@$1,'$2')]")

	// --- Combinators ---
	// Child combinator (>) → /
	reChild := regexp.MustCompile(`\s*>\s*`)
	xpath = reChild.ReplaceAllString(xpath, "/")

	// Adjacent sibling (+) → /following-sibling::*[1]/
	reAdj := regexp.MustCompile(`\s*\+\s*`)
	xpath = reAdj.ReplaceAllString(xpath, "/following-sibling::*[1]/")

	// General sibling (~) → /following-sibling::/
	reGen := regexp.MustCompile(`\s*~\s*`)
	xpath = reGen.ReplaceAllString(xpath, "/following-sibling::")

	// Descendant (space) → //
	reDesc := regexp.MustCompile(`\s+`)
	xpath = reDesc.ReplaceAllString(xpath, "//")

	// --- Pseudo-classes ---
	// :nth-child(n) → [position()=n]
	reNth := regexp.MustCompile(`:nth-child\((\d+)\)`)
	xpath = reNth.ReplaceAllString(xpath, "[position()=$1]")

	// :first-child → [position()=1]
	xpath = strings.ReplaceAll(xpath, ":first-child", "[position()=1]")

	// :last-child → [position()=last()]
	xpath = strings.ReplaceAll(xpath, ":last-child", "[position()=last()]")

	// :not(selector) → [not(...)] – naive handling, single simple selector only
	reNot := regexp.MustCompile(`:not\(([^)]+)\)`)
	xpath = reNot.ReplaceAllString(xpath, "[not($1)]")

	// Prepend // if not already
	if !strings.HasPrefix(xpath, "//") {
		xpath = "//" + xpath
	}

	return xpath
}
