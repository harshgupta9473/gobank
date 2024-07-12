package component

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}


type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type TransactionRequest struct {
	Reciever int64 `json:"reciver"`
	Amount   int64 `json:"amount"`
}

type Transaction struct {
	Sender   int64     `json:"sender"`
	Reciever int64     `json:"reciever"`
	Amount   int64     `json:"amount"`
	Time     time.Time `json:"time"`
}

type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstname"`
	LastName          string    `json:"lastname"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"createdat"`
}

func (a *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(pw)) == nil
}

func NewAccount(firstName, lastname string, password string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		// ID:        rand.Intn(1000),
		FirstName:         firstName,
		LastName:          lastname,
		Number:            int64(rand.Intn(100000000)),
		EncryptedPassword: string(encpw),
		Balance:           0,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func NewTransaction(sender,reciever,amount int64)(*Transaction){
	return &Transaction{
		Sender: sender,
		Reciever: reciever,
		Amount: amount,
		Time: time.Now().UTC(),
	}
}
