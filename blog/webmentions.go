package blog

import (
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"hawx.me/code/tally-ho/internal/mfutil"
	"hawx.me/code/tally-ho/webmention"
)

func (b *Blog) sendWebmentions(location string, data map[string][]interface{}) {
	// ensure that the entry exists
	time.Sleep(time.Second)

	links := findMentionedLinks(data)
	slog.Info("sending webmentions", slog.Any("links", links))

	if !b.local {
		for _, link := range links {
			if err := webmention.Send(location, link); err != nil {
				slog.Error("send webmention", slog.String("source", location), slog.String("target", link), slog.Any("err", err))
			}
		}
	}
}

func (b *Blog) sendUpdateWebmentions(location string, oldData, newData map[string][]interface{}) {
	links := findMentionedLinks(newData)

	for _, oldLink := range findMentionedLinks(oldData) {
		if !slices.Contains(links, oldLink) {
			links = append(links, oldLink)
		}
	}

	slog.Info("sending webmentions", slog.Any("links", links))

	if !b.local {
		for _, link := range links {
			if err := webmention.Send(location, link); err != nil {
				slog.Error("send webmention", slog.String("source", location), slog.String("target", link), slog.Any("err", err))
			}
		}
	}
}

func findMentionedLinks(data map[string][]interface{}) []string {
	linkSet := map[string]struct{}{}

	for _, link := range findAs(data) {
		linkSet[link] = struct{}{}
	}

	for key, value := range data {
		if strings.HasPrefix(key, "hx-") ||
			strings.HasPrefix(key, "mp-") ||
			key == "url" ||
			len(value) == 0 {
			continue
		}

		if v, ok := mfutil.Get(data, key+".properties.url", key).(string); ok {
			if u, err := url.Parse(v); err == nil && u.IsAbs() {
				linkSet[v] = struct{}{}
			}
		}
	}

	var links []string
	for link := range linkSet {
		links = append(links, link)
	}

	return links
}

func findAs(data map[string][]interface{}) []string {
	content, ok := mfutil.SafeGet(data, "content.html")
	if !ok {
		return []string{}
	}

	htmlContent, ok := content.(string)
	if !ok {
		return []string{}
	}

	root, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		slog.Error("find-as", slog.Any("err", err))
		return []string{}
	}

	as := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode &&
			node.DataAtom == atom.A &&
			htmlutil.Has(node, "href")
	})

	var links []string
	for _, a := range as {
		if val := htmlutil.Attr(a, "href"); val != "" {
			links = append(links, val)
		}
	}

	return links
}
