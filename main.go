package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"./filter"
	"./models"
	"./storage"
)

var commentStorage *storage.DatabaseStorage

// addCommentHandler обрабатывает POST‑запросы для добавления комментария
func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим JSON‑тело запроса
	var comment models.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if comment.Text == "" || comment.NewsID == "" || comment.Author == "" {
		http.Error(w, "Text, NewsID and Author are required", http.StatusBadRequest)
		return
	}

	// Генерируем ID и время создания
	comment.ID = generateID()
	comment.CreatedAt = time.Now().Format(time.RFC3339)

	// Проверяем комментарий на запрещённые слова
	filterResult := filter.CheckComment(comment.Text)
	comment.IsApproved = filterResult.IsApproved

	// Сохраняем комментарий в БД
	err = commentStorage.SaveComment(&comment)
	if err != nil {
		log.Printf("Error saving comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&comment)
}

// getCommentsHandler обрабатывает GET‑запросы для получения комментариев по ID новости
func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем news_id из параметров запроса
	newsID := r.URL.Query().Get("news_id")
	if newsID == "" {
		http.Error(w, "news_id is required", http.StatusBadRequest)
		return
	}

	// Получаем комментарии из БД
	comments, err := commentStorage.GetCommentsByNewsID(newsID)
	if err != nil {
		log.Printf("Error getting comments: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

// startFilterWorker запускает асинхронный процесс проверки комментариев
func startFilterWorker() {
	go func() {
		// В будущем можно добавить периодическую проверку комментариев
		// Сейчас проверка выполняется при сохранении комментария
		log.Println("Filter worker started")
		for {
			time.Sleep(5 * time.Second)
		}
	}()
}

// generateID генерирует уникальный ID для комментария
func generateID() string {
	return "comment_" + time.Now().UTC().String()
}

// initLogging настраивает базовое логирование
func initLogging() {
	log.SetPrefix("[CommentService] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	initLogging()

	// Строка подключения к PostgreSQL
	connectionString := "host=localhost port=5432 user=postgres password=password dbname=comment_db sslmode=disable"

	// Инициализируем хранилище
	var err error
	commentStorage, err = storage.NewDatabaseStorage(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")

	// Инициализируем схему БД
	err = commentStorage.InitializeSchema()
	if err != nil {
		log.Fatalf("Schema initialization failed: %v", err)
	}

	// Проверяем соединение с БД
	err = commentStorage.CheckConnection()
	if err != nil {
		log.Fatalf("DB connection check failed: %v", err)
	}

	// Запускаем фоновый процесс проверки комментариев
	startFilterWorker()

	// Регистрируем обработчики HTTP‑запросов
	http.HandleFunc("/comment", addCommentHandler)
	http.HandleFunc("/comments", getCommentsHandler)

	log.Println("Comment Service is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
