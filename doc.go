// Package safebooru is a small, dependency-free client for the Safebooru API
// (https://safebooru.org, the Gelbooru-style "index.php?page=dapi" API).
//
// It can search posts by tag, count the posts for a search, list tags, and
// build image and thumbnail URLs:
//
//	client := safebooru.NewClient(safebooru.WithUserAgent("my-app/1.0"))
//
//	posts, err := client.Posts(ctx, safebooru.PostsQuery{
//		Tags:  []string{"hatsune_miku", "smile"},
//		Limit: 20,
//	})
//	if err != nil {
//		// handle err
//	}
//	for _, p := range posts {
//		fmt.Println(client.ImageURL(p))
//	}
//
// Every method takes a context.Context, so requests can be cancelled or given
// a deadline. A Client is safe for concurrent use, which makes it easy to fetch
// many pages in parallel; please keep the request rate polite.
//
// Errors: a non-200 response is a *StatusError, and a "soft" failure
// (HTTP 200 with {"success": false}) is an *APIError. When the API says its
// search is down, errors.Is(err, ErrSearchDown) is true.
package safebooru
