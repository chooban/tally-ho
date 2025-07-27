package main

import (
	"flag"
	"os"
)

const (
	DbPath           = "DB_PATH"
	MediaDir         = "MEDIA_DIR"
	WebPath          = "WEB_PATH"
	Port             = "PORT"
	Socket           = "SOCKET"
	MeUrl            = "MY_URL"
	MeName           = "MY_NAME"
	Title            = "SITE_TITLE"
	Description      = "SITE_DESCRIPTION"
	BaseUrl          = "BASE_URL"
	MediaUrl         = "MEDIA_URL"
	BlueskyHandle    = "BLUESKY_HANDLE"
	BlueskyAppKey    = "BLUESKY_APP_KEY"
	BlueskyPdsUrl    = "BLUESKY_PDS_URL"
	AuthUrl          = "AUTH_ENDPOINT"
	TokenUrl         = "TOKEN_ENDPOINT"
	BypassValidation = "BYPASS_VALIDATION"
)

func parseConfig() config {
	var (
		webPath  = flag.String("web", "web", "")
		dbPath   = flag.String("db", "file::memory:", "")
		mediaDir = flag.String("media-dir", "", "")
		port     = flag.String("port", "8080", "")
		socket   = flag.String("socket", "", "")
	)
	flag.Usage = usage
	flag.Parse()

	var conf config

	conf.DbPath = *dbPath
	if p := os.Getenv(DbPath); p != "" {
		conf.DbPath = p
	}

	conf.MediaDir = *mediaDir
	if p := os.Getenv(MediaDir); p != "" {
		conf.MediaDir = p
	}
	conf.WebPath = *webPath
	if p := os.Getenv(WebPath); p != "" {
		conf.WebPath = p
	}
	conf.Port = *port
	if p := os.Getenv(Port); p != "" {
		conf.Port = p
	}
	conf.Socket = *socket
	if p := os.Getenv(Socket); p != "" {
		conf.Socket = p
	}

	if p := os.Getenv(MeUrl); p != "" {
		conf.Me = p
	}
	if p := os.Getenv(MeName); p != "" {
		conf.Name = p
	}
	if p := os.Getenv(Title); p != "" {
		conf.Title = p
	}
	if p := os.Getenv(Description); p != "" {
		conf.Description = p
	}
	if p := os.Getenv(BaseUrl); p != "" {
		conf.BaseURL = p
	}
	if p := os.Getenv(MediaUrl); p != "" {
		conf.MediaURL = p
	}
	if p := os.Getenv(AuthUrl); p != "" {
		conf.AuthEndpoint = p
	}
	if p := os.Getenv(TokenUrl); p != "" {
		conf.TokenEndpoint = p
	}
	if p := os.Getenv(BlueskyAppKey); p != "" {
		conf.Bluesky.AppKey = p
	}
	if p := os.Getenv(BlueskyHandle); p != "" {
		conf.Bluesky.Handle = p
	}
	if p := os.Getenv(BlueskyPdsUrl); p != "" {
		conf.Bluesky.Pds = p
	} else {
		conf.Bluesky.Pds = "https://bsky.social"
	}
	if p := os.Getenv(BypassValidation); p != "" {
		if p == "true" {
			conf.BypassValidation = true
		}
	}

	return conf
}
