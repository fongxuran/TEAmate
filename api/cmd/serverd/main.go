package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	notionconnector "teammate/internal/connector/notion"
	notionhandler "teammate/internal/handler/rest/integrations/notion"
	messageshandler "teammate/internal/handler/rest/messages"
	mvphandler "teammate/internal/handler/rest/mvp"
	"teammate/internal/realtime"
	messagesrepo "teammate/internal/repository/messages"
)

func main() {
	port := getEnv("PORT", "8080")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close database: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	repo := messagesrepo.NewRepository(db)
	msgHandler := messageshandler.New(repo)
	hub := realtime.NewHub()
	mvp := mvphandler.New(hub)

	// Optional Notion integration (T-010): defaults to dry-run unless NOTION_DRY_RUN=false.
	notionClient := notionconnector.NewClient(notionconnector.LoadConfigFromEnv())
	notionREST := notionhandler.New(notionClient)

	authUsers := loadAuthUsers()
	corsAllowedOrigins := getEnvCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"})

	router := newRouter(authUsers, corsAllowedOrigins, hub, msgHandler, mvp, notionREST)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func loadAuthUsers() map[string]string {
	if getEnvBool("API_AUTH_DISABLED", false) {
		return nil
	}

	user := getEnv("API_BASIC_AUTH_USER", "admin")
	pass := getEnv("API_BASIC_AUTH_PASS", "password")
	if strings.TrimSpace(user) == "" {
		return nil
	}
	return map[string]string{user: pass}
}

func getEnvCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		res = append(res, p)
	}
	if len(res) == 0 {
		return fallback
	}
	return res
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
