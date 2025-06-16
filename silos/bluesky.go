package silos

import (
	"context"
	"errors"
	"log/slog"

	gobot "github.com/danrusei/gobot-bsky"
	"hawx.me/code/tally-ho/internal/mfutil"
)

const BlueskyUID = "https://bsky.app"

type BlueskyOptions struct {
	Handle string
	AppKey string
}

func Bluesky(options BlueskyOptions) (*blueskyClient, error) {
	ctx := context.Background()
	agent := gobot.NewAgent(ctx, "https://bsky.social", options.Handle, options.AppKey)
	//client, err := botsky.NewClient(ctx, options.Handle, options.AppKey)
	//if err != nil {
	//	slog.Error(err.Error())
	//	return nil, err
	//}
	agent.Connect(ctx)

	slog.Info("Returning bluesky client")
	return &blueskyClient{client: &agent, handle: options.Handle}, nil
}

type blueskyClient struct {
	client *gobot.BskyAgent
	handle string
}

func (c *blueskyClient) Name() string {
	return "@" + c.handle
}

func (c *blueskyClient) UID() string {
	return BlueskyUID
}

func conv[T any](x any) T {
	v, _ := x.(T)
	return v
}

func (c *blueskyClient) Create(data map[string][]interface{}) (location string, err error) {
	switch data["hx-kind"][0].(string) {
	case "note":
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
