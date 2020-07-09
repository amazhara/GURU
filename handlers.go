package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Middleware handler to chek if post request
func postValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(405)
			fmt.Fprintf(w, "Please, use Post method")
			return
		}
	})
}

func addDepositHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	responce := map[string]interface{}{"error": ""}
	var req struct {
		UserID    uint64  `json:"userid"`
		DepositID uint64  `json:"depostid"`
		Amount    float64 `json:"amount"`
		Token     string  `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// Add error to responce
		responce["error"] = err.Error()
	}
	// Validating
	_, exists := deposits[req.DepositID]
	if exists != false {
		responce["error"] = "Deposit ID already exists"
	}
	user, exists := users[req.UserID]
	if exists != true {
		responce["error"] = "No such user exists"
	} else if user.Token != req.Token {
		responce["error"] = "Wrong token"
	}
	// If some error exists
	if responce["error"] != "" {
		jr, _ := json.Marshal(responce)
		w.Write(jr)
		return
	}

	var deposit Deposit
	deposit.userID = req.UserID
	deposit.balanceBefore = user.Balance
	user.Balance += req.Amount
	deposit.balanceAfter = user.Balance
	deposit.date = time.Now()
	deposits[req.DepositID] = deposit
	usersUpdate[req.UserID] = user

	// Save stats
	stat := userStats[req.UserID]
	stat.depositCount++
	stat.depostSum += req.Amount

	responce["balance"] = deposit.balanceAfter
	jr, _ := json.Marshal(responce)
	w.Write(jr)
}

// Add user logic
func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Declare responce
	responce := map[string]string{"error": ""}
	// Get user data from request
	var req struct {
		ID      uint64  `json:"id"`
		Balance float64 `json:"balance"`
		Token   string  `json:"Token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// Add error to responce
		responce["error"] = err.Error()
	} else {
		_, exist := users[req.ID]
		if exist != false {
			responce["error"] = "This user id already exists"
		} else {
			users[req.ID] = &User{req.Balance, req.Token}
			userStats[req.ID] = &UserStat{}
		}
	}
	jr, _ := json.Marshal(responce)
	// Set writer header to json
	w.Header().Set("Content-Type", "application/json")
	w.Write(jr)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ID    uint64 `json:"id"`
		Token string `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jr, _ := json.Marshal([]string{"error", err.Error()})
		w.Write(jr)
		return
	}
	user, exist := users[req.ID]
	if exist != true {
		jr, _ := json.Marshal([]string{"error", "Such user ID not exist"})
		w.Write(jr)
		return
	} else if user.Token != req.Token {
		jr, _ := json.Marshal([]string{"error", "Wrong token for this data"})
		w.Write(jr)
		return
	}
	stats := userStats[req.ID]
	responce := map[string]interface{}{
		"id":           req.ID,
		"balance":      user.Balance,
		"depositCount": stats.depositCount,
		"depostSum":    stats.depostSum,
		"betCount":     stats.betCount,
		"betSum":       stats.betSum,
		"winCount":     stats.winCount,
		"winSum":       stats.winSum,
	}
	jr, _ := json.Marshal(responce)
	w.Write(jr)
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		UserID        uint64  `json:"userid"`
		TransactionID uint64  `json:"transactionid"`
		Type          string  `json:"type"`
		Amount        float64 `json:"amount"`
		Token         string  `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jr, _ := json.Marshal([]string{"error", err.Error()})
		w.Write(jr)
		return
	}
	responce := map[string]interface{}{"error": ""}
	user, exist := users[req.UserID]
	if exist != true {
		responce["error"] = "No such user exist"
	} else if user.Token != req.Token {
		responce["error"] = "Wrong Token"
	}
	_, exist = transactions[req.TransactionID]
	if exist != false {
		responce["error"] = "Such transaction ID alredy exists"
	}
	if responce["error"] != "" {
		jr, _ := json.Marshal(responce)
		w.Write(jr)
		return
	}
	var transaction Transaction
	transaction.balanceBefore = user.Balance
	transaction.Amount = req.Amount
	if req.Type == "Bet" {
		if user.Balance < req.Amount {
			responce["error"] = "Not enough finances to bet"
		} else {
			user.Balance -= transaction.Amount
		}
	} else if req.Type == "Win" {
		user.Balance += transaction.Amount
	} else {
		responce["error"] = "No such transaction type"
	}
	if responce["error"] != "" {
		jr, _ := json.Marshal(responce)
		w.Write(jr)
		return
	}

	transaction.ID = req.TransactionID
	transaction.balanceAfter = user.Balance
	transaction.date = time.Now()
	transactions[req.TransactionID] = transaction
	responce["balance"] = user.Balance
	usersUpdate[req.UserID] = user

	// Save stats
	stat := userStats[req.UserID]
	if req.Type == "Bet" {
		stat.betCount++
		stat.betSum += req.Amount
	} else if req.Type == "Win" {
		stat.winCount++
		stat.winSum += req.Amount
	}

	jr, _ := json.Marshal(responce)
	w.Write(jr)
}

func updateCollection() {
	for _, user := range usersUpdate {
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			log.Fatal(err)
		}
	}
	usersUpdate = make(map[uint64]*User)
}
