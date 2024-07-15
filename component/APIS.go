package component

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	jwt "github.com/golang-jwt/jwt/v5"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {

	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(status)
	// return nil
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/account/user/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountInfo), s.store))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))
	router.HandleFunc("/account/delete/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleDeleteAccount), s.store))
	router.HandleFunc("/account/transfer/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleTransfer), s.store))
	router.HandleFunc("/account/transactions/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleTransactions), s.store))
	log.Println("JSON API server is running on port:", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByNumber(int64(req.Number))
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Number: acc.Number,
		Token:  token,
		ID:     acc.ID,
	}

	return WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleTransactions(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		accNo, err := DecodingJWT(r)
		if err != nil {
			return fmt.Errorf("permission Denied")
		}
		transactions, err := s.store.GetAllTransaction(accNo)
		if err != nil {
			return fmt.Errorf("no transactions available")
		}
		WriteJSON(w, http.StatusOK, transactions)
		return nil
	}
	return fmt.Errorf("method not allowed")
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	// if r.Method == "GET" {
	// 	return s.handleGetAccount(w, r)
	// }
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccountInfo(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		accNo, err := DecodingJWT(r)
		if err != nil {
			return fmt.Errorf("permission Denied")
		}
		acc, err := s.store.GetAccountByNumber(accNo)
		if err != nil {
			return fmt.Errorf("error occured")
		}
		WriteJSON(w, http.StatusOK, acc)
		return nil

	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// get /account /accounts
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err

	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := new(CreateAccountRequest)
	// createAccountRequest:= CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}
	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Password, createAccountRequest.PIN)

	if err != nil {
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "DELETE" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	id, err := getID(r)
	if err != nil {
		return err

	}
	acc, err := s.store.GetAccountByID(id)
	if err != nil {
		return fmt.Errorf("account does not exist")
	}
	if acc.Balance > 0 {
		return fmt.Errorf("you still have %d money left in your account you cannot delete it", acc.Balance)
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransactionRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()
	accountno, err := DecodingJWT(r)
	if err != nil {
		return fmt.Errorf("invalid transaction")
	}
	account, err := s.store.GetAccountByNumber(accountno)
	if err != nil {
		return fmt.Errorf("error")
	}
	if !account.ValidPIN(transferReq.PIN) {
		return fmt.Errorf("invalid transaction")
	}
	if account.Balance < transferReq.Amount {
		return fmt.Errorf("low balance")
	}
	err = s.store.TransferMoney(transferReq, account)
	if err != nil {
		return fmt.Errorf("error occcured in transaction %s", err)
	}
	transection := NewTransaction(accountno, transferReq.Reciever, transferReq.Amount)
	err = s.store.CreateTransactionBlock(transection)
	if err != nil {
		return err
	}
	return nil
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id is given %s", idStr)
	}
	return id, nil
}

// func getAccountNo(r *http.Request) (int64, error) {
// 	accountnostring := mux.Vars(r)["account"]
// 	accountno, err := strconv.Atoi(accountnostring)
// 	if err != nil {
// 		return 0, fmt.Errorf("prohibited %s", accountnostring)
// 	}
// 	return int64(accountno), nil
// }

func DecodingJWT(r *http.Request) (int64, error) {
	tokenString := r.Header.Get("x-jwt-token")
	token, err := validateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	claims := token.Claims.(jwt.MapClaims)
	return int64(claims["accountNumber"].(float64)), nil
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)

		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return
		}
		id, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("error occured"))
		}
		account, err := s.GetAccountByID(id)
		if err != nil {
			permissionDenied(w)
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		accountno := int64(claims["accountNumber"].(float64))
		if account.Number != accountno {
			permissionDenied(w)
			return
		}
		_, err = s.GetAccountByNumber(accountno)
		if err != nil {
			permissionDenied(w)
			return
		}
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET_KEY")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func createJWT(account *Account) (string, error) {

	// Create the Claims
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))

}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}
