package safebooru_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Ouja/safebooru"
)

func Example() {
	client := safebooru.NewClient(safebooru.WithUserAgent("my-app/1.0"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	posts, err := client.Posts(ctx, safebooru.PostsQuery{
		Tags:  []string{"hatsune_miku"},
		Limit: 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range posts {
		fmt.Println(p.ID, client.ThumbnailURL(p))
	}
}

func ExampleClient_PostCount() {
	client := safebooru.NewClient()

	count, err := client.PostCount(context.Background(), []string{"mika_melatika"})
	if err != nil {
		log.Fatal(err)
	}

	const perPage = 100
	pages := (count + perPage - 1) / perPage
	fmt.Printf("%d posts in %d pages\n", count, pages)
}

func ExampleStatusError() {
	client := safebooru.NewClient()

	_, err := client.Posts(context.Background(), safebooru.PostsQuery{Tags: []string{"smile"}})

	var statusErr *safebooru.StatusError
	switch {
	case errors.As(err, &statusErr) && statusErr.StatusCode == 429:
		fmt.Println("rate limited: slow down")
	case errors.Is(err, safebooru.ErrSearchDown):
		fmt.Println("search is down, try later")
	case err != nil:
		fmt.Println("request failed:", err)
	}
}
