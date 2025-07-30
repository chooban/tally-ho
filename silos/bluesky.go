package silos

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	gobot "github.com/danrusei/gobot-bsky"
	"hawx.me/code/tally-ho/internal/mfutil"
)

const BlueskyUID = "https://bsky.app"

type BlueskyOptions struct {
	Handle string
	AppKey string
	Pds    *url.URL
}

func Bluesky(options BlueskyOptions) *BlueskyClient {
	ctx := context.Background()
	agent := gobot.NewAgent(ctx, options.Pds.String(), options.Handle, options.AppKey)

	return &BlueskyClient{client: &agent, handle: options.Handle}
}

type BlueskyClient struct {
	client *gobot.BskyAgent
	handle string
}

func (c *BlueskyClient) Name() string {
	return "@" + c.handle
}

func (c *BlueskyClient) UID() string {
	return BlueskyUID
}

func conv[T any](x any) T {
	v, _ := x.(T)
	return v
}

func (c *BlueskyClient) Create(data map[string][]interface{}) (location string, err error) {
	err = c.client.Connect(context.TODO())
	if err != nil {
		return "", err
	}
	switch data["hx-kind"][0].(string) {
	case "note":
		slog.Info("Posting note to bluesky")
		noteContent, ok := mfutil.Get(data, "content.text", "content").(string)
		if !ok {
			return "", errors.New("invalid note content")
		}
		//hashTags := conv[[]string](mfutil.GetAll(data, "category"))
		postText, err := gobot.NewPostBuilder(noteContent).Build()
		if err != nil {
			return "", err
		}

		_, uri, err := c.client.PostToFeed(context.Background(), postText)

		if err != nil {
			return "", err
		}

		return uri, nil
	}
	return "", ErrUnsure{data}
}
