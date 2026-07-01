package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Comment struct {
	ID        string `json:"id"`
	NewsID    string `json:"news_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Text      string `json:"text"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
}

var db *sql.DB

func initDB() {
	connStr := "host=localhost port=5432 user=postgres password=password dbname=comment_db sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		news_id TEXT NOT NULL,
		parent_id TEXT,
		text TEXT NOT NULL,
		author TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Schema init failed: %v", err)
	}
	log.Println("[*] Database connected and schema initialized")
}

type contextKey string

const reqIDKey contextKey = "request_id"

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.URL.Query().Get("request_id")
		if reqID == "" {
			reqID = r.Header.Get("X-Request-ID")
		}
		if reqID == "" {
			reqID = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), reqIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		reqID, _ := r.Context().Value(reqIDKey).(string)
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = strings.Split(fwd, ",")[0]
		}
		log.Printf("[%s] %s | %s %s | %d | %v",
			reqID, ip, r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var c Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if c.Text == "" || c.NewsID == "" || c.Author == "" {
		http.Error(w, `{"error":"text, news_id and author are required"}`, http.StatusBadRequest)
		return
	}

	c.ID = "comment_" + uuid.New().String()
	c.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	query := `INSERT INTO comments (id, news_id, parent_id, text, author, created_at) VALUES ($1,$2,$3,$4,$5,$6)`
	if _, err := db.Exec(query, c.ID, c.NewsID, c.ParentID, c.Text, c.Author, c.CreatedAt); err != nil {
		log.Printf("Error saving comment: %v", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	newsID := r.URL.Query().Get("news_id")
	if newsID == "" {
		http.Error(w, `{"error":"news_id is required"}`, http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`SELECT id, news_id, parent_id, text, author, created_at FROM comments WHERE news_id=$1 ORDER BY created_at`, newsID)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := make([]Comment, 0)
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.NewsID, &c.ParentID, &c.Text, &c.Author, &c.CreatedAt)
		comments = append(comments, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/comment", addCommentHandler)
	mux.HandleFunc("/comments", getCommentsHandler)

	handler := RequestIDMiddleware(LoggingMiddleware(mux))

	server := &http.Server{Addr: ":8081", Handler: handler}

	go func() {
		log.Println("[*] Comments Service HTTP server is started on localhost:8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[*] Comments Service HTTP server has been stopped. Reason: got %s", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
