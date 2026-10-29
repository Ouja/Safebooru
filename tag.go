package safebooru

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
)

// Tag is one tag. Count is the number of posts using it. Type is the
// booru tag category (usually 0 general, 1 artist, 3 copyright,
// 4 character, 5 meta).
type Tag struct {
	ID    int64  `xml:"id,attr" json:"id"`
	Name  string `xml:"name,attr" json:"name"`
	Count int    `xml:"count,attr" json:"count"`
	Type  int    `xml:"type,attr" json:"type"`
}

// tagsXML is the XML wrapper around the list of tags.
type tagsXML struct {
	XMLName xml.Name `xml:"tags"`
	Tags    []Tag    `xml:"tag"`
}

// TagsQuery holds the parameters for Tags.
type TagsQuery struct {
	ID    int64 // fetch one specific tag (0 = not set)
	Page  int   // the API's pid, starts at 0
	Limit int   // 0 = API default of 100
}

// Tags returns one page of tags, ordered by ID. The tag API does not
// report a total; keep asking for the next page until it returns fewer
// than q.Limit tags.
func (c *Client) Tags(ctx context.Context, q TagsQuery) ([]Tag, error) {
	params := url.Values{}
	params.Set("s", "tag")
	params.Set("q", "index")
	params.Set("json", "1") // ignored if the endpoint only speaks XML
	if q.Limit > 0 {
		params.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Page > 0 {
		params.Set("pid", strconv.Itoa(q.Page))
	}
	if q.ID > 0 {
		params.Set("id", strconv.FormatInt(q.ID, 10))
	}

	body, err := c.get(ctx, params)
	if err != nil {
		return nil, err
	}

	switch {
	case len(body) == 0:
		return []Tag{}, nil
	case body[0] == '[': // JSON
		var tags []Tag
		if err := json.Unmarshal(body, &tags); err != nil {
			return nil, fmt.Errorf("safebooru: decode tags json: %w", err)
		}
		return tags, nil
	case body[0] == '<': // XML
		var res tagsXML
		if err := xml.Unmarshal(body, &res); err != nil {
			return nil, fmt.Errorf("safebooru: decode tags xml: %w", err)
		}
		return res.Tags, nil
	default:
		return nil, fmt.Errorf("safebooru: unexpected tags response format")
	}
}
