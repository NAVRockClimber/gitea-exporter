package main

import (
	"flag"
	"gitea-exporter/prom"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	configFile := flag.String("config", "config.yaml", "File path to the targets config")
	listenAddress := flag.String("server", ":9115", "Port the server is listening on")
	probePath := flag.String("probe", "/probe", "Path for the probe")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("Starte API Exporter", "address", *listenAddress, "path", *probePath, "config file", *configFile)

	handler := prom.NewHandler(*configFile, logger)
	// handler.TestDate(*configFile)

	http.HandleFunc(*probePath, handler.ProbeHandler)
	http.ListenAndServe(*listenAddress, nil)
}
