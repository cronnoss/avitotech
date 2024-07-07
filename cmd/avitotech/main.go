package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cronnoss/avitotech/internal/app"
	"github.com/cronnoss/avitotech/internal/logger"
	internalhttp "github.com/cronnoss/avitotech/internal/server/http"
	"github.com/cronnoss/avitotech/internal/storage"
)

func main() {
	conf := NewConfig().AvitotechConf
	storage := storage.NewStorage(conf.Storage)
	logger := logger.NewLogger(conf.Logger.Level, os.Stdout)
	avitotech := app.NewAvitotech(logger, conf, storage)
	httpsrv := internalhttp.NewServer(logger, avitotech, conf.HTTP.Host, conf.HTTP.Port)

	avitotech.Run(httpsrv)

	filename := filepath.Base(os.Args[0])
	fmt.Printf("%s stopped\n", filename)
}
