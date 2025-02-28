package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"wallet-service/domain/entity"
	"wallet-service/domain/repository"
	"wallet-service/domain/service"
	"wallet-service/middleware"

	pb "wallet-service/docs/grpc/gen"
)

var secretKey = []byte("jwt_token_example")

type Server struct {
	pb.UnimplementedAuthServiceServer
}

type App struct {
	UserService   service.UserService
	WalletService service.WalletService
}

// --- Генерация токена ---
func generateToken(user entity.User) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// --- Регистрация ---
func (a *App) registrHandler(w http.ResponseWriter, r *http.Request) {
	var regUser entity.User
	err := json.NewDecoder(r.Body).Decode(&regUser)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = a.UserService.Registration(regUser)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// --- Логин ---
func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	var user entity.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = a.UserService.Authorization(user)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// --- Пополнение баланса ---
func (a *App) depositHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount float64 `json:"amount"`
	}

	username := r.Context().Value("username").(string)

	user := entity.User{Username: username}
	userID, err := a.UserService.GetUserID(user)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = a.WalletService.Deposit(userID, req.Amount)
	if err != nil {
		http.Error(w, "Deposit failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Deposit successful"})
}

// --- Перевод средств ---
func (a *App) transferHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToUsername string  `json:"to_username"`
		Amount     float64 `json:"amount"`
	}

	username := r.Context().Value("username").(string)

	user := entity.User{Username: username}
	fromUserID, err := a.UserService.GetUserID(user)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	recipient := entity.User{Username: req.ToUsername}
	toUserID, err := a.UserService.GetUserID(recipient)
	if err != nil {
		http.Error(w, "Recipient not found", http.StatusBadRequest)
		return
	}

	err = a.WalletService.Transfer(fromUserID, toUserID, req.Amount)
	if err != nil {
		http.Error(w, "Transfer failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transfer successful"})
}

// --- gRPC валидация токена ---
func (s *Server) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	username, err := middleware.ValidateToken(req.GetToken())
	if err != nil {
		return nil, err
	}
	return &pb.ValidateResponse{Login: username}, nil
}

// --- Запуск серверов ---
func (a *App) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	// Подключение к БД
	var err error
	dsn := "host=localhost user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable port=5432"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	fmt.Println("Успешное подключение к БД")

	// Инициализация сервисов
	userRepo := repository.NewPostgresUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	a.UserService = *service.NewUserService(userRepo)
	a.WalletService = *service.NewWalletService(walletRepo)

	// HTTP-сервер
	r := mux.NewRouter()
	r.HandleFunc("/register", a.registrHandler).Methods("POST")
	r.HandleFunc("/login", a.loginHandler).Methods("POST")

	// Маршруты с авторизацией
	authRoutes := r.PathPrefix("/api").Subrouter()
	authRoutes.Use(middleware.Auth)

	authRoutes.HandleFunc("/deposit", a.depositHandler).Methods("POST")
	authRoutes.HandleFunc("/transfer", a.transferHandler).Methods("POST")

	go func() {
		log.Println("Starting HTTP server on port :8088")
		if err := http.ListenAndServe(":8088", r); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// gRPC-сервер
	lis, err := net.Listen("tcp", ":50059")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &Server{})
	log.Println("Starting gRPC server on port :50059")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
