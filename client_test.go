package safebooru

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient starts a fake Safebooru that always answers with body.
func newTestClient(t *testing.T, status int, body string, gotQuery *string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotQuery != nil {
			*gotQuery = r.URL.RawQuery
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewClient(WithBaseURL(srv.URL))
}

func TestPosts(t *testing.T) {
	var query string
	c := newTestClient(t, 200, `[{"id":7,"directory":12,"image":"a.jpg","tags":"miku smile","score":null}]`, &query)

	posts, err := c.Posts(context.Background(), PostsQuery{Tags: []string{"miku", "smile"}, Page: 2, Limit: 5000})
	if err != nil {
		t.Fatalf("Posts error: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != 7 || len(posts[0].TagList()) != 2 {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if want := "json=1&limit=1000&page=dapi&pid=2&q=index&s=post&tags=miku+smile"; query != want {
		t.Fatalf("query = %q, want %q", query, want)
	}
	if got := c.ImageURL(posts[0]); got != c.baseURL+"/images/12/a.jpg" {
		t.Fatalf("ImageURL = %q", got)
	}
	png := Post{Directory: 4438, Image: "7adb5fcd.png"}
	if got, want := c.ThumbnailURL(png), c.baseURL+"/thumbnails/4438/thumbnail_7adb5fcd.jpg"; got != want {
		t.Fatalf("ThumbnailURL = %q, want %q", got, want)
	}
}

func TestPostsEmptyBody(t *testing.T) {
	posts, err := newTestClient(t, 200, "", nil).Posts(context.Background(), PostsQuery{})
	if err != nil || len(posts) != 0 {
		t.Fatalf("posts = %v, err = %v; want empty, nil", posts, err)
	}
}

func TestSearchDown(t *testing.T) {
	_, err := newTestClient(t, 200, `{"success":false,"message":"search down"}`, nil).Posts(context.Background(), PostsQuery{})
	if !errors.Is(err, ErrSearchDown) {
		t.Fatalf("err = %v, want ErrSearchDown", err)
	}
}

func TestStatusError(t *testing.T) {
	_, err := newTestClient(t, 503, "", nil).Posts(context.Background(), PostsQuery{})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != 503 {
		t.Fatalf("err = %v, want StatusError 503", err)
	}
}

func TestTagsXMLAndJSON(t *testing.T) {
	xmlBody := `<?xml version="1.0" encoding="UTF-8"?><tags type="array"><tag type="4" count="10" name="hatsune_miku" ambiguous="false" id="1"/></tags>`
	tags, err := newTestClient(t, 200, xmlBody, nil).Tags(context.Background(), TagsQuery{})
	if err != nil || len(tags) != 1 || tags[0].Name != "hatsune_miku" || tags[0].Count != 10 {
		t.Fatalf("xml tags = %+v, err = %v", tags, err)
	}

	tags, err = newTestClient(t, 200, `[{"id":2,"name":"smile","count":3,"type":0}]`, nil).Tags(context.Background(), TagsQuery{})
	if err != nil || len(tags) != 1 || tags[0].Name != "smile" {
		t.Fatalf("json tags = %+v, err = %v", tags, err)
	}
}

func TestPostCount(t *testing.T) {
	var query string
	body := `<?xml version="1.0" encoding="UTF-8"?><posts count="91" offset="0"><post id="1" tags=" a b "/></posts>`
	count, err := newTestClient(t, 200, body, &query).PostCount(context.Background(), []string{"mika_melatika"})
	if err != nil || count != 91 {
		t.Fatalf("count = %d, err = %v; want 91, nil", count, err)
	}
	if want := "limit=1&page=dapi&q=index&s=post&tags=mika_melatika"; query != want {
		t.Fatalf("query = %q, want %q", query, want)
	}
}

func TestSearchDownXML(t *testing.T) {
	body := `<?xml version="1.0" encoding="UTF-8"?><response success="false" reason="search down"/>`
	_, err := newTestClient(t, 200, body, nil).PostCount(context.Background(), []string{"smile"})
	if !errors.Is(err, ErrSearchDown) {
		t.Fatalf("err = %v, want ErrSearchDown", err)
	}
}

func TestAPIErrorMessage(t *testing.T) {
	cases := map[string]string{
		"json":     `{"success":false,"message":"too deep"}`,
		"xml attr": `<?xml version="1.0"?><response success="false" reason="too deep"/>`,
		"xml text": `<?xml version="1.0"?><response success="false">too deep</response>`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := newTestClient(t, 200, body, nil).PostCount(context.Background(), nil)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Message != "too deep" {
				t.Fatalf("err = %v, want APIError \"too deep\"", err)
			}
			if errors.Is(err, ErrSearchDown) {
				t.Fatalf("a non-search-down APIError must not match ErrSearchDown")
			}
		})
	}
}

func TestResponseTooLarge(t *testing.T) {
	big := "[" + strings.Repeat(" ", maxBodySize) + "]"
	_, err := newTestClient(t, 200, big, nil).Posts(context.Background(), PostsQuery{})
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("err = %v, want ErrResponseTooLarge", err)
	}
}

func TestOptions(t *testing.T) {
	c := NewClient(WithBaseURL("https://example.com/"), WithHTTPClient(nil))
	if c.baseURL != "https://example.com" {
		t.Fatalf("baseURL = %q, trailing slash not trimmed", c.baseURL)
	}
	if c.httpClient == nil {
		t.Fatal("WithHTTPClient(nil) replaced the default client with nil")
	}
	if got := c.ImageURL(Post{Directory: 1, Image: "a.png"}); got != "https://example.com/images/1/a.png" {
		t.Fatalf("ImageURL = %q", got)
	}
}

func TestUserAgent(t *testing.T) {
	gotUA := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA <- r.UserAgent()
	}))
	t.Cleanup(srv.Close)

	c := NewClient(WithBaseURL(srv.URL), WithUserAgent("test-app/1.0"))
	if _, err := c.Posts(context.Background(), PostsQuery{}); err != nil {
		t.Fatal(err)
	}
	if ua := <-gotUA; ua != "test-app/1.0" {
		t.Fatalf("User-Agent = %q", ua)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := newTestClient(t, 200, "[]", nil).Posts(ctx, PostsQuery{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestPostCountEdgeCases(t *testing.T) {
	cases := map[string]struct {
		body    string
		want    int
		wantErr bool
	}{
		"zero":       {`<?xml version="1.0"?><posts count="0" offset="0"></posts>`, 0, false},
		"empty body": {``, 0, false},
		"no count":   {`<?xml version="1.0"?><posts offset="0"></posts>`, 0, true},
		"wrong root": {`<?xml version="1.0"?><tags></tags>`, 0, true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := newTestClient(t, 200, tc.body, nil).PostCount(context.Background(), nil)
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("got %d, err %v; want %d, wantErr %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
