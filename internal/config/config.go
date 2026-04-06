package config

import "os"

type Config struct {
    Port            string
    SessionsFile    string
    PreferencesFile string
}

func Load() Config {
    return Config{
        Port:            getEnv("PORT", "8080"),
        SessionsFile:    getEnv("SESSIONS_FILE", "data/sessions.json"),
        PreferencesFile: getEnv("PREFERENCES_FILE", "data/preferences.json"),
    }
}

func getEnv(key, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }
    return value
}
