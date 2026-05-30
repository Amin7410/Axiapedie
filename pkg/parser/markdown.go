package parser

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	// Regex matches [[Page Title]] or [[Page Title|Display Text]]
	wikiLinkRegex = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
)

// ParseToHTML converts Wiki-flavored Markdown to HTML.
func ParseToHTML(source string, existFn func(string) bool) (string, error) {
	// 1. Pre-process Wiki links: [[Title]] or [[Title|Label]] -> custom HTML links
	processedSource := wikiLinkRegex.ReplaceAllStringFunc(source, func(match string) string {
		content := match[2 : len(match)-2] // Strip [[ and ]]
		
		target := content
		label := content
		
		if idx := strings.Index(content, "|"); idx != -1 {
			target = strings.TrimSpace(content[:idx])
			label = strings.TrimSpace(content[idx+1:])
		} else {
			target = strings.TrimSpace(content)
			label = target
		}
		
		if target == "" {
			return match
		}

		if existFn(target) {
			// Blue link (Page exists)
			return fmt.Sprintf(`<a href="/wiki/%s" hx-get="/wiki/%s" hx-target="#wiki-content" hx-push-url="true" class="text-blue-600 hover:underline">%s</a>`, 
				url.PathEscape(target), url.PathEscape(target), label)
		} else {
			// Red link (Page does not exist)
			return fmt.Sprintf(`<a href="/editor/%s" hx-get="/editor/%s" hx-target="#wiki-content" hx-push-url="true" class="wiki-link-red" title="This page does not exist yet. Click to create.">%s</a>`, 
				url.PathEscape(target), url.PathEscape(target), label)
		}
	})

	// 2. Setup Goldmark with GitHub Flavored Markdown
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(), // Allow raw HTML inside markdown for flexibility
		),
	)

	// 3. Convert to HTML
	var buf bytes.Buffer
	if err := md.Convert([]byte(processedSource), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
