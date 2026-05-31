package parser

import (
	"bytes"
	"sort"
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
	// Sắp xếp terms theo độ dài giảm dần để khớp các cụm từ dài trước (tránh Google bị khớp bởi Go)
	sortedTerms := make([]*domain.GlossaryTerm, len(terms))
	copy(sortedTerms, terms)
	sort.Slice(sortedTerms, func(i, j int) bool {
		return len(sortedTerms[i].Term) > len(sortedTerms[j].Term)
	})

	var dict []string
	for _, t := range sortedTerms {
		dict = append(dict, t.Term)
	}

	var matcher *ahocorasick.Matcher
	if len(dict) > 0 {
		matcher = ahocorasick.NewStringMatcher(dict)
	}

	return &GlossaryAnnotator{
		matcher: matcher,
		terms:   sortedTerms,
	}
}

// annotateTextNode đệ quy phân tách nút văn bản thành tập hợp các nút văn bản thường và nút span chứa tooltip
func annotateTextNode(n *html.Node, terms []*domain.GlossaryTerm) []*html.Node {
	text := n.Data
	if text == "" {
		return []*html.Node{n}
	}

	var bestMatchStart = -1
	var bestMatchLen = 0
	var bestTerm *domain.GlossaryTerm

	// Tìm vị trí xuất hiện đầu tiên của bất kỳ term nào trong văn bản
	for i := 0; i < len(text); i++ {
		for _, t := range terms {
			if i+len(t.Term) <= len(text) && text[i:i+len(t.Term)] == t.Term {
				bestMatchStart = i
				bestMatchLen = len(t.Term)
				bestTerm = t
				break
			}
		}
		if bestMatchStart != -1 {
			break
		}
	}

	if bestMatchStart == -1 {
		return []*html.Node{n}
	}

	beforeText := text[:bestMatchStart]
	matchText := text[bestMatchStart : bestMatchStart+bestMatchLen]
	afterText := text[bestMatchStart+bestMatchLen:]

	var result []*html.Node
	if beforeText != "" {
		result = append(result, &html.Node{
			Type: html.TextNode,
			Data: beforeText,
		})
	}

	// Tạo thẻ span cho glossary term
	span := &html.Node{
		Type: html.ElementNode,
		Data: "span",
		Attr: []html.Attribute{
			{Key: "class", Val: "glossary-term relative border-b border-dotted border-blue-500 cursor-help group"},
			{Key: "hx-get", Val: "/ui/glossary/tooltip/" + bestTerm.ID},
			{Key: "hx-trigger", Val: "mouseenter once"},
			{Key: "hx-target", Val: "find .tooltip-content"},
		},
	}
	span.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: matchText,
	})
	span.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "tooltip-content absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block z-50 min-w-max"},
		},
	})

	result = append(result, span)

	if afterText != "" {
		remainingNode := &html.Node{
			Type: html.TextNode,
			Data: afterText,
		}
		result = append(result, annotateTextNode(remainingNode, terms)...)
	}

	return result
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
			// Kiểm tra nhanh sự trùng khớp bằng StringMatcher trước khi chạy bộ lọc đệ quy
			matches := a.matcher.Match([]byte(n.Data))
			if len(matches) > 0 {
				newNodes := annotateTextNode(n, a.terms)
				if len(newNodes) > 1 || (len(newNodes) == 1 && newNodes[0] != n) {
					parent := n.Parent
					nextSibling := n.NextSibling
					parent.RemoveChild(n)
					for _, child := range newNodes {
						parent.InsertBefore(child, nextSibling)
					}
				}
				return
			}
		}

		// Duyệt đệ quy qua các node con
		var children []*html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}
		for _, c := range children {
			f(c)
		}
	}

	// Định vị thẻ body của cây DOM được sinh ra bởi html.Parse
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

