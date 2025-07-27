package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/media"
	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/silos"
	"hawx.me/code/tally-ho/webmention"
	"hawx.me/code/tally-ho/websub"
)

func usage() {
	fmt.Println(`Usage: tally-ho [options]

	--web DIR=web
	--db PATH=file::memory
	--media-dir DIR
	--port PORT=8080
	--socket PATH`)
}

type config struct {
	Me               string
	Name             string
	Title            string
	Description      string
	BaseURL          string
	MediaURL         string
	MediaDir         string
	DbPath           string
	WebPath          string
	AuthEndpoint     string
	TokenEndpoint    string
	BypassValidation bool

	Flickr, Twitter struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}

	Github struct {
		AccessToken string
	}
	Bluesky struct {
		Handle string
		AppKey string
		Pds    string
	}
	Port   string
	Socket string
}

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelInfo)

	var conf = parseConfig()

	baseURL, err := url.Parse(conf.BaseURL)
	if err != nil {
		logger.Error("base url invalid", slog.Any("err", err))
		return
	}

	mediaURL, err := url.Parse(conf.MediaURL)
	if err != nil {
		logger.Error("media url invalid", slog.Any("err", err))
		return
	}

	db, err := sql.Open("sqlite3", conf.DbPath)
	if err != nil {
		logger.Error("error opening sqlite file", slog.String("path", conf.DbPath), slog.Any("err", err))
		return
	}

	fw := &blog.FileWriter{
		MediaDir: conf.MediaDir,
		MediaURL: mediaURL,
	}

	var blogSilos []any
	var micropubSyndicateTo []micropub.SyndicateTo

	if conf.Twitter.ConsumerKey != "" {
		twitter, err := silos.Twitter(silos.TwitterOptions{
			ConsumerKey:       conf.Twitter.ConsumerKey,
			ConsumerSecret:    conf.Twitter.ConsumerSecret,
			AccessToken:       conf.Twitter.AccessToken,
			AccessTokenSecret: conf.Twitter.AccessTokenSecret,
		}, fw)
		if err != nil {
			logger.Warn("twitter", slog.Any("err", err))
		} else {
			blogSilos = append(blogSilos, twitter)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  twitter.UID(),
				Name: twitter.Name(),
			})
		}
	}

	if conf.Flickr.ConsumerKey != "" {
		flickr, err := silos.Flickr(silos.FlickrOptions{
			ConsumerKey:       conf.Flickr.ConsumerKey,
			ConsumerSecret:    conf.Flickr.ConsumerSecret,
			AccessToken:       conf.Flickr.AccessToken,
			AccessTokenSecret: conf.Flickr.AccessTokenSecret,
		})
		if err != nil {
			logger.Warn("flickr", slog.Any("err", err))
		} else {
			blogSilos = append(blogSilos, flickr)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  flickr.UID(),
				Name: flickr.Name(),
			})
		}
	}

	if conf.Github.AccessToken != "" {
		github, err := silos.Github(silos.GithubOptions{
			AccessToken: conf.Github.AccessToken,
		})
		if err != nil {
			logger.Warn("github", slog.Any("err", err))
		} else {
			blogSilos = append(blogSilos, github)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  github.UID(),
				Name: github.Name(),
			})
		}
	}

	if conf.Bluesky.AppKey != "" {
		pdsUrl, _ := url.Parse(conf.Bluesky.Pds)
		bluesky := silos.Bluesky(silos.BlueskyOptions{
			Handle: conf.Bluesky.Handle,
			AppKey: conf.Bluesky.AppKey,
			Pds:    pdsUrl,
		})
		logger.Info("Adding bluesky syndication")
		blogSilos = append(blogSilos, bluesky)
		micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
			UID:  bluesky.UID(),
			Name: bluesky.Name(),
		})
	} else {
		logger.Info("Not configuring Bluesky syndicator")
	}

	hubStore, err := blog.NewHubStore(db)
	if err != nil {
		logger.Error("problem initialising hub store", slog.Any("err", err))
		return
	}

	mediaEndpointURL, _ := url.Parse("/-/media")
	hubEndpointURL, _ := url.Parse("/-/hub")

	websubhub := websub.New(baseURL.ResolveReference(hubEndpointURL).String(), hubStore)

	authURL, _ := url.Parse(conf.AuthEndpoint)
	tokenURL, _ := url.Parse(conf.TokenEndpoint)
	myUrl, _ := url.Parse(conf.Me)

	b, err := blog.New(logger, blog.Config{
		Me:          myUrl,
		Name:        conf.Name,
		Title:       conf.Title,
		Description: conf.Description,
		BaseURL:     baseURL,
		MediaURL:    mediaURL,
		AuthURL:     authURL,
		TokenURL:    tokenURL,
		MediaDir:    conf.MediaDir,
		HubURL:      baseURL.ResolveReference(hubEndpointURL).String(),
	}, db, websubhub, blogSilos)
	if err != nil {
		logger.Error("problem initialising blog", slog.Any("err", err))
		return
	}
	defer b.Close()

	http.Handle("/", b.Handler())

	http.Handle("/-/media-file/",
		http.StripPrefix("/-/media-file/", http.FileServer(http.Dir(conf.MediaDir))),
	)
	http.Handle("/public/",
		http.StripPrefix("/public/",
			http.FileServer(
				http.Dir(filepath.Join(conf.WebPath, "static")))))

	http.Handle("/-/micropub", micropub.Endpoint(
		b,
		conf.Me,
		baseURL.ResolveReference(mediaEndpointURL).String(),
		micropubSyndicateTo,
		fw,
		conf.BypassValidation,
	))
	http.Handle("/-/webmention", webmention.Endpoint(b))
	http.Handle("/-/media", auth.Only(conf.Me, media.Endpoint(fw, auth.HasScope)))
	http.Handle("/-/hub", websubhub)

	serve.Serve(conf.Port, conf.Socket, http.DefaultServeMux)
}
