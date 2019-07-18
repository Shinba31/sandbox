package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Shinba31/sandbox/rank_fetcher/rank"
	"github.com/gocolly/colly"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var word, store string
	var maxPage int
	flag.StringVar(&word, "word", "渋谷+居酒屋", "search word")
	flag.StringVar(&store, "store", "隠れ野", "store name")
	flag.IntVar(&maxPage, "max", 2, "max page for search")
	flag.Parse()

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.90 Safari/537.36"),
	)
	fetcher := rank.NewFetcher(c, maxPage)

	rank, err := fetcher.Fetch(ctx, word, store)
	if err != nil {
		panic(err)
	}

	fmt.Printf("result: %s, %d\n", store, rank)
}
