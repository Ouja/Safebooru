# safebooru

[![Go Reference](https://pkg.go.dev/badge/github.com/Ouja/safebooru.svg)](https://pkg.go.dev/github.com/Ouja/safebooru)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**A tiny, zero-dependency Go client for the [Safebooru](https://safebooru.org) API.**
Search posts by tag, count results, list tags, and get image and thumbnail URLs.
Standard library only, `context` everywhere, and safe to share between goroutines.

> 🌱 **A side quest from a learning journey.**
> This package started inside [go-booru](https://github.com/Ouja/go-booru), a project I built
> to learn **Go concurrency** (worker pools, rate limits, retries, cancellation). The API client
> turned out clean enough to stand on its own, so here it is. It's small and still growing,
> and feedback, issues and suggestions are genuinely welcome.

## Install

```bash
go get github.com/Ouja/safebooru
```

Requires Go 1.22 or newer.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Ouja/safebooru"
)

func main() {
	client := safebooru.NewClient(safebooru.WithUserAgent("my-app/1.0"))

	posts, err := client.Posts(context.Background(), safebooru.PostsQuery{
		Tags:  []string{"mika_melatika", "smile"},
		Limit: 10,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range posts {
		fmt.Println(p.ID, client.ImageURL(p))
	}
}
```

## What it can do

| Method | What you get |
|---|---|
| `Posts(ctx, PostsQuery)` | One page of posts for a tag search (newest first) |
| `PostCount(ctx, tags)` | The total number of posts, so you know how many pages exist |
| `Tags(ctx, TagsQuery)` | One page of tags with their post counts and categories |
| `ImageURL(post)` | The full-size image URL |
| `ThumbnailURL(post)` | The thumbnail URL (always `.jpg`) |

### Paging through a whole tag

```go
count, err := client.PostCount(ctx, []string{"mika_melatika"})
if err != nil {
	log.Fatal(err)
}

const perPage = 100
pages := (count + perPage - 1) / perPage // round up

for page := 0; page < pages; page++ {
	posts, err := client.Posts(ctx, safebooru.PostsQuery{
		Tags:  []string{"mika_melatika"},
		Page:  page, // starts at 0
		Limit: perPage,
	})
	// ...
}
```

A `Client` is safe for concurrent use, so these pages can also be fetched in parallel by a
worker pool. [go-booru's crawler](https://github.com/Ouja/go-booru/blob/main/internal/crawler/crawler.go)
shows one way to do it with a shared rate limit and retries. **Please keep it gentle**
(see below).

## Errors

```go
posts, err := client.Posts(ctx, query)

var statusErr *safebooru.StatusError
var apiErr *safebooru.APIError
switch {
case errors.As(err, &statusErr) && statusErr.StatusCode == 429:
	// rate limited: slow down
case errors.Is(err, safebooru.ErrSearchDown):
	// the API says its search is down: try later
case errors.As(err, &apiErr):
	// HTTP 200 but {"success": false}; apiErr.Message says why
case err != nil:
	// network error, timeout, cancelled context, unexpected response...
}
```

A search with no results is **not** an error: you get an empty slice.

## Options

```go
client := safebooru.NewClient(
	safebooru.WithUserAgent("my-app/1.0 (+https://example.com)"), // please identify your app
	safebooru.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
	safebooru.WithBaseURL("https://safebooru.org"),
)
```

`WithBaseURL` can point the client at other sites that use the same Gelbooru-style `dapi`
API. That's untested territory, so your mileage may vary.

## Be a good API citizen

Safebooru is a community site that kindly offers a free API. Let's keep it that way:

- **Set a User-Agent** that identifies your app.
- **Don't hammer it.** Keep requests to a few per second, and back off on errors.
- **Respect the artists.** The images belong to their creators. Credit them, link to the
  original, and don't redistribute their work without permission.

## Not affiliated

This is an unofficial client. It is not affiliated with or endorsed by Safebooru.
API reference: [Safebooru API documentation](https://safebooru.org/index.php?page=help&topic=dapi).

## License

[MIT](LICENSE). This covers the code in this repository, not any content you fetch with it.
