package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AnnaTarantina/CommentServise/filter"
	"github.com/AnnaTarantina/CommentServise/models"
	"github.com/AnnaTarantina/CommentServise/storage"
	"github.com/google/uuid"

	_ "github.com/lib/pq" // драйвер PostgreSQL
)

var commentStorage *storage.DatabaseStorage

// addCommentHandler обрабатывает POST-запросы для добавления комментария
func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if comment.Text == "" || comment.NewsID == "" || comment.Author == "" {
		http.Error(w, "Text, NewsID and Author are required", http.StatusBadRequest)
		return
	}

	comment.ID = generateID()
	comment.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	filterResult := filter.CheckComment(comment.Text)
	comment.IsApproved = filterResult.IsApproved

	if err := commentStorage.SaveComment(&comment); err != nil {
		log.Printf("Error saving comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&comment); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// getCommentsHandler обрабатывает GET-запросы для получения комментариев по ID новости
func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	newsID := r.URL.Query().Get("news_id")
	if newsID == "" {
		http.Error(w, "news_id is required", http.StatusBadRequest)
		return
	}

	comments, err := commentStorage.GetCommentsByNewsID(newsID)
	if err != nil {
		log.Printf("Error getting comments: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// startFilterWorker запускает асинхронный процесс проверки комментариев
func startFilterWorker() {
	go func() {
		log.Println("Filter worker started")
		for {
			time.Sleep(5 * time.Second)
		}
	}()
}

// generateID генерирует уникальный ID для комментария
func generateID() string {
	return "comment_" + uuid.New().String()
}

// initLogging настраивает базовое логирование
func initLogging() {
	log.SetPrefix("[CommentService] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	initLogging()

	connectionString := "host=localhost port=5432 user=postgres password=password dbname=comment_db sslmode=disable"

	var err error
	commentStorage, err = storage.NewDatabaseStorage(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established")

	if err = commentStorage.InitializeSchema(); err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}

	if err = commentStorage.CheckConnection(); err != nil {
		log.Fatalf("DB connection check failed: %v", err)
	}

	startFilterWorker()

	http.HandleFunc("/comment", addCommentHandler)
	http.HandleFunc("/comments", getCommentsHandler)

	log.Println("Comment Service is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
