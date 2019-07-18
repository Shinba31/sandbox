package rank

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/gocolly/colly"
)

type Fetcher interface {
	Fetch(ctx context.Context, searchWord, store string) (int, error)
}

type fetcher struct {
	collector *colly.Collector
	maxPage   int
}

func NewFetcher(collector *colly.Collector, maxPage int) Fetcher {
	return &fetcher{
		collector: collector,
		maxPage:   maxPage,
	}
}

func (f *fetcher) Fetch(ctx context.Context, searchWord, store string) (int, error) {
	const google = "https://www.google.com/search"

	p := &parser{
		maxPage:    f.maxPage,
		searchWord: searchWord,
		store:      store,
	}

	f.collector.OnHTML("a[href]", p.onHTMLAnker)
	f.collector.OnHTML("a#pnnext.pn", p.onHTMLAnkerPnnext)
	f.collector.OnHTML("div#search", p.onHTMLDivSearch)

	f.collector.OnResponse(func(r *colly.Response) {
		fmt.Println(r.StatusCode)
	})

	err := f.collector.Visit(google + buildQuery(searchWord))
	return p.rank, err
}

type parser struct {
	maxPage     int
	pageCounter int
	searchWord  string
	store       string
	rank        int
}

func (p *parser) onHTMLAnker(e *colly.HTMLElement) {
	link := e.Attr("href")
	if e.ChildText("span") == "さらに表示" {
		log.Println(url.QueryUnescape(link))
		e.Request.Visit(link)
		return
	}
}

func (p *parser) onHTMLAnkerPnnext(e *colly.HTMLElement) {
	if p.pageCounter >= p.maxPage {
		return
	}
	p.pageCounter++

	query := e.Response.Request.URL.Query()
	if !isLocalSearch(query) {
		return
	}
	link := e.Attr("href")
	if e.ChildText("span") == "次へ" {
		log.Println(url.QueryUnescape(link))
		e.Request.Visit(link)
		return
	}
}

func (p *parser) onHTMLDivSearch(e *colly.HTMLElement) {
	query := e.Response.Request.URL.Query()
	if !isLocalSearch(query) {
		return
	}
	start := defaultQueryInt32(e.Response.Request, "start", 0)
	counter := int(start + 1)
	e.ForEach("div[role]", func(_ int, ce *colly.HTMLElement) {
		if ce.Attr("role") != "heading" {
			return
		}
		ce.ForEach("div", func(_ int, cce *colly.HTMLElement) {
			//fmt.Printf("%s,%d\n", cce.Text, contentCounter)
			if cce.Text == p.store {
				// got it
				p.rank = counter
			}
			counter++
		})
	})
}

func buildQuery(q string) string {
	v := url.Values{}
	v.Set("q", q)
	v.Set("oq", q)
	v.Set("hl", "ja")
	v.Set("lr", "lang_ja")
	v.Set("ie", "UTF-8")

	return "?" + v.Encode()
}

func isLocalSearch(q url.Values) bool {
	return q.Get("tmb") == "lcl"
}

func defaultQueryInt32(r *colly.Request, key string, def int32) int32 {
	str := r.URL.Query().Get(key)
	res, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return def
	}
	return int32(res)
}
