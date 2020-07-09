package main

import "time"

type User struct {
	Balance float64
	Token   string
}

type Deposit struct {
	balanceBefore float64
	balanceAfter  float64
	date          time.Time
	userID        uint64
}

type Transaction struct {
	ID            uint64
	Amount        float64
	balanceBefore float64
	balanceAfter  float64
	date          time.Time
	userID        uint64
}

type UserStat struct {
	depositCount uint64
	depostSum    float64
	betCount     uint64
	betSum       float64
	winCount     uint64
	winSum       float64
}
