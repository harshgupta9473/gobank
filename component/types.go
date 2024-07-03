package component

import (
	"math/rand"
	"time"
)

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"createdat"`
}

func NewAccount(firstName, lastname string) *Account {
	return &Account{
		// ID:        rand.Intn(1000),
		FirstName: firstName,
		LastName:  lastname,
		Number:    int64(rand.Intn(100000000)),
		Balance:   0,
		CreatedAt: time.Now().UTC(),
	}
}
