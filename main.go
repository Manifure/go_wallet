package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
)

type WalletRequest struct {
	WalletID      uuid.UUID `json:"walletId"`
	OperationType string    `json:"operationType"`
	Amount        float64   `json:"amount"`
}

type Wallet struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  float64   `json:"balance"`
}

var db *sql.DB

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	// Подключение к базе данных
	var err error
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка ping к базе данных: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet", handleWalletOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{walletId}", getWalletBalance).Methods("GET")

	fmt.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var req WalletRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	if req.OperationType != "DEPOSIT" && req.OperationType != "WITHDRAW" {
		http.Error(w, "Неверный тип операции", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var currentBalance float64

	err = tx.QueryRow("SELECT balance FROM wallets WHERE wallet_id = $1", req.WalletID).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		if req.OperationType == "DEPOSIT" {
			_, err = tx.Exec("INSERT INTO wallets (wallet_id, balance) VALUES ($1, $2)", req.WalletID, req.Amount)
			if err != nil {
				http.Error(w, "Ошибка создания кошелька", http.StatusInternalServerError)
				return
			}
			tx.Commit()
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Кошелек создан и пополнен")
			return
		} else {
			http.Error(w, "Кошелек не найден", http.StatusNotFound)
			return
		}
	} else if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	if req.OperationType == "WITHDRAW" {
		if currentBalance < req.Amount {
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
			return
		}
		currentBalance -= req.Amount
	} else if req.OperationType == "DEPOSIT" {
		currentBalance += req.Amount
	}

	_, err = tx.Exec("UPDATE wallets SET balance = $1 WHERE wallet_id = $2", currentBalance, req.WalletID)
	if err != nil {
		http.Error(w, "Ошибка обновления баланса", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Ошибка сохранения транзакции", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Операция успешно выполнена. Новый баланс: %.2f", currentBalance)
}

func getWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		http.Error(w, "Некорректный UUID кошелька", http.StatusBadRequest)
		return
	}

	var balance float64
	err = db.QueryRow("SELECT balance FROM wallets WHERE wallet_id = $1", walletID).Scan(&balance)
	if err == sql.ErrNoRows {
		http.Error(w, "Кошелек не найден", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	wallet := Wallet{
		WalletID: walletID,
		Balance:  balance,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}
