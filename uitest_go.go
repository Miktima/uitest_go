package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	url_chromedriver := ""
	// create context

	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), url_chromedriver)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Create context with a 30-second timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// run task list
	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://pkg.go.dev/time`),
		chromedp.Text(`.Documentation-overview`, &res, chromedp.NodeVisible),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(strings.TrimSpace(res))
}
