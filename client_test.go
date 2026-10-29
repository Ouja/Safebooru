package safebooru

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
