package parser

import (
	"bytes"
	"fmt"
	"strings"

	"axia-wiki/internal/domain"
	"github.com/cloudflare/ahocorasick"
	"golang.org/x/net/html"
)

type GlossaryAnnotator struct {
	matcher *ahocorasick.Matcher
	terms   []*domain.GlossaryTerm
}

func NewGlossaryAnnotator(terms []*domain.GlossaryTerm) *GlossaryAnnotator {
	var dict []string
	for _, t := range terms {
		dict = append(dict, t.Term)
	}
	matcher := ahocorasick.NewStringMatcher(dict)
	return &GlossaryAnnotator{
		matcher: matcher,
		terms:   terms,
	}
}

func (a *GlossaryAnnotator) AnnotateHTML(inputHTML string) string {
	if len(a.terms) == 0 {
		return inputHTML
	}

	doc, err := html.Parse(strings.NewReader(inputHTML))
	if err != nil {
		return inputHTML
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		// Bỏ qua các thẻ pre, code, a (để không gắn tooltip lồng nhau hoặc trong code)
		if n.Type == html.ElementNode && (n.Data == "pre" || n.Data == "code" || n.Data == "a") {
			return
		}

		if n.Type == html.TextNode {
			// Find matches in the text
			matches := a.matcher.Match([]byte(n.Data))
			if len(matches) > 0 {
				// We need to split the text node into multiple nodes (text -> span -> text)
				// For MVP simplicity and because altering AST in x/net/html is complex,
				// we just do a string replace if it matches, and hope it's clean since we only touch TextNodes.
				
				newData := n.Data
				// Iterate backwards to avoid index shifting, though ahocorasick returns indices.
				// Simpler approach for text nodes: string replacement of the dictionary words.
				// Since it's a TextNode, it's 100% safe to replace raw string.
				for _, t := range a.terms {
					if strings.Contains(newData, t.Term) {
						replacement := fmt.Sprintf(`<span class="glossary-term relative border-b border-dotted border-blue-500 cursor-help group" hx-get="/ui/glossary/tooltip/%s" hx-trigger="mouseenter once" hx-target="find .tooltip-content">%s<div class="tooltip-content absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block z-50 min-w-max"></div></span>`, t.ID, t.Term)
						newData = strings.ReplaceAll(newData, t.Term, replacement)
					}
				}
				
				if newData != n.Data {
					// Hack: Because we injected HTML string into a TextNode, x/net/html will escape it!
					// We must parse the newData into nodes and replace n.
					parsedFrag, _ := html.ParseFragment(strings.NewReader(newData), n.Parent)
					parent := n.Parent
					nextSibling := n.NextSibling
					parent.RemoveChild(n)
					for _, child := range parsedFrag {
						parent.InsertBefore(child, nextSibling)
					}
				}
			}
		}

		// Traverse children safely (copy slice since we might modify DOM)
		var children []*html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}
		for _, c := range children {
			f(c)
		}
	}

	// Bỏ qua thẻ html, head, body sinh ra bởi Parse
	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)

	if body != nil {
		for c := body.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
		
		var buf bytes.Buffer
		for c := body.FirstChild; c != nil; c = c.NextSibling {
			html.Render(&buf, c)
		}
		return buf.String()
	}

	return inputHTML
}
