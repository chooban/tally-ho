package main

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"log/slog"
	"os"
)

const (
	DbPath      = "DB_PATH"
	MediaDir    = "MEDIA_DIR"
	WebPath     = "WEB_PATH"
	Port        = "PORT"
	Socket      = "SOCKET"
	MeUrl       = "MY_URL"
	MeName      = "MY_NAME"
	Title       = "SITE_TITLE"
	Description = "SITE_DESCRIPTION"
	BaseUrl     = "BASE_URL"
	MediaUrl    = "MEDIA_URL"
)

func parseConfig(logger *slog.Logger) config {

	var (
		configPath = flag.String("config", "./config.toml", "")
		webPath    = flag.String("web", "web", "")
		dbPath     = flag.String("db", "file::memory:", "")
		mediaDir   = flag.String("media-dir", "", "")
		port       = flag.String("port", "8080", "")
		socket     = flag.String("socket", "", "")
	)
	flag.Usage = usage
	flag.Parse()

	// First, try to read a config file
	var conf config
	if _, err := os.Stat(*configPath); errors.Is(err, os.ErrNotExist) {
		logger.Warn("Config file does not exist")
	} else {
		if _, err := toml.DecodeFile(*configPath, &conf); err != nil {
			logger.Info("config could not be decoded", slog.Any("err", err))
		}
	}

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

	return conf
}
