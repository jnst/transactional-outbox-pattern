// Package main provides the HTTP API server for the transactional outbox pattern.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jnst/transactional-outbox-pattern/internal/config"
	"github.com/jnst/transactional-outbox-pattern/internal/logger"
	"github.com/jnst/transactional-outbox-pattern/internal/model"
	"github.com/jnst/transactional-outbox-pattern/internal/repository"
	"github.com/jnst/transactional-outbox-pattern/internal/service"
)

const (
	contentTypeJSON        = "Content-Type"
	applicationJSON        = "application/json"
	failedToEncodeResponse = "failed to encode response"
	decimalBase            = 10
	int64BitSize           = 64
	signalBufferSize       = 1
	exitCode               = 1
)

// APIServer handles HTTP requests for user management.
type APIServer struct {
	userService service.UserService
}

// NewAPIServer creates a new API server instance.
func NewAPIServer(userService service.UserService) *APIServer {
	return &APIServer{
		userService: userService,
	}
}

// CreateUser handles POST /users endpoint for user creation.
func (s *APIServer) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params model.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := s.userService.CreateUser(r.Context(), &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(contentTypeJSON, applicationJSON)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, failedToEncodeResponse, http.StatusInternalServerError)
		return
	}
}

// GetUser handles GET /users/get endpoint for user retrieval.
func (s *APIServer) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, decimalBase, int64BitSize)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	user, err := s.userService.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set(contentTypeJSON, applicationJSON)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, failedToEncodeResponse, http.StatusInternalServerError)
		return
	}
}

// HealthCheck handles GET /health endpoint for service health check.
func (*APIServer) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(contentTypeJSON, applicationJSON)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, failedToEncodeResponse, http.StatusInternalServerError)
		return
	}
}

func main() {
	// 環境変数読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(exitCode)
	}

	// ログ設定
	loggerInstance := logger.Setup(cfg.LogLevel)
	slog.SetDefault(loggerInstance)

	// データベース接続
	dbURL := cfg.DatabaseURL

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(exitCode)
	}
	defer dbPool.Close()

	// 依存関係注入
	userRepo := repository.NewUserRepositoryImpl(dbPool)
	outboxRepo := repository.NewOutboxRepositoryImpl(dbPool)
	transactionMgr := repository.NewTransactionManagerImpl(dbPool)
	userService := service.NewUserServiceImpl(userRepo, outboxRepo, transactionMgr)

	// APIサーバー初期化
	server := NewAPIServer(userService)

	// ルート定義
	http.HandleFunc("/users", server.CreateUser)
	http.HandleFunc("/users/get", server.GetUser)
	http.HandleFunc("/health", server.HealthCheck)

	// サーバー起動
	port := cfg.Port

	slog.Info("starting API server", slog.String("service", "api"), slog.String("port", port))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("failed to start server", slog.String("error", err.Error()))
		return
	}
}
