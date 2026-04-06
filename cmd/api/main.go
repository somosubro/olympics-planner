package main

import (
    "log/slog"
    "net/http"
    "os"

    "olympics-planner/internal/config"
    apphttp "olympics-planner/internal/transport/http"
)

func main() {
    cfg := config.Load()

    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    logger.Info("starting olympics planner api", "port", cfg.Port)

    router := apphttp.NewRouter(logger)

    if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
        logger.Error("server stopped", "error", err)
        os.Exit(1)
    }
}
