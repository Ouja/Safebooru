package safebooru

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// Post is one post, with the field names the Safebooru JSON API uses.
// Use Client.ImageURL and Client.ThumbnailURL to get its image URLs.
type Post struct {
	ID        int64  `json:"id"`
	Directory int64  `json:"directory"` // a number in the API, e.g. 4321
	Image     string `json:"image"`
	Hash      string `json:"hash"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Tags      string `json:"tags"` // space-separated
	Rating    string `json:"rating"`
	Score     int    `json:"score"` // null in the JSON becomes 0
	Owner     string `json:"owner"`
	Change    int64  `json:"change"` // unix time of the last change
}

// TagList splits the space-separated Tags string.
func (p Post) TagList() []string {
	return strings.Fields(p.Tags)
}

// PostsQuery holds the search parameters for Posts.
type PostsQuery struct {
	Tags  []string // joined with spaces; "-tag" excludes a tag
	Page  int      // the API's pid, starts at 0
	Limit int      // 1..MaxPostLimit (0 = API default of 100)
}

// Posts returns one page of posts matching q.Tags, newest first.
// When nothing matches, it returns an empty slice and no error.
// Use PostCount to find out how many pages there are.
func (c *Client) Posts(ctx context.Context, q PostsQuery) ([]Post, error) {
	params := url.Values{}
	params.Set("s", "post")
	params.Set("q", "index")
	params.Set("json", "1")
	params.Set("pid", strconv.Itoa(max(q.Page, 0)))
	if q.Limit > 0 {
		params.Set("limit", strconv.Itoa(min(q.Limit, MaxPostLimit)))
	}
	if len(q.Tags) > 0 {
		params.Set("tags", strings.Join(q.Tags, " "))
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return []Post{}, nil // the API sends an empty body when nothing matches
	}

	var posts []Post
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, fmt.Errorf("safebooru: decode posts: %w", err)
	}
	return posts, nil
}

// ImageURL builds the URL of the post's full-size image.
func (c *Client) ImageURL(p Post) string {
	return c.baseURL + "/images/" + strconv.FormatInt(p.Directory, 10) + "/" + p.Image
}

// ThumbnailURL builds the URL of the post's small preview image.
// Thumbnails are always .jpg, even when the original is .png or .gif.
func (c *Client) ThumbnailURL(p Post) string {
	name := strings.TrimSuffix(p.Image, path.Ext(p.Image))
	return c.baseURL + "/thumbnails/" + strconv.FormatInt(p.Directory, 10) + "/thumbnail_" + name + ".jpg"
}

// PostCount returns how many posts match the tags, which tells you how
// many pages Posts can return. (The JSON API has no total, so this makes
// one small XML request and reads only its <posts count="..."> attribute.)
func (c *Client) PostCount(ctx context.Context, tags []string) (int, error) {
	params := url.Values{}
	params.Set("s", "post")
	params.Set("q", "index")
	params.Set("limit", "1") // we only need the count, not the posts
	if len(tags) > 0 {
		params.Set("tags", strings.Join(tags, " "))
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return 0, err
	}
	if len(body) == 0 {
		return 0, nil
	}

	root, err := firstElement(xml.NewDecoder(bytes.NewReader(body)))
	if err != nil {
		return 0, fmt.Errorf("safebooru: decode post count: %w", err)
	}
	if root.Name.Local != "posts" {
		return 0, fmt.Errorf("safebooru: decode post count: unexpected root <%s>", root.Name.Local)
	}
	for _, attr := range root.Attr {
		if attr.Name.Local == "count" {
			count, err := strconv.Atoi(attr.Value)
			if err != nil {
				return 0, fmt.Errorf("safebooru: decode post count: %w", err)
			}
			return count, nil
		}
	}
	return 0, errors.New("safebooru: decode post count: no count attribute")
}
