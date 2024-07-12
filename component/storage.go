package component

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
}

type PostgressStore struct {
	db *sql.DB
}

func NewPostgressStore() (*PostgressStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgressStore{
		db: db,
	}, nil
}

func (s *PostgressStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgressStore) CreateAccountTable() error {
	query := `create table if not exists account(
	id serial primary key,
	first_name varchar(100),
	last_name varchar(100),
	number serial,
	encrypted_password varchar(100),
	balance serial,
	created_at timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateAccount(acc *Account) error {
	query := `insert into account
	(first_name,last_name,number,encrypted_password,balance,created_at)
	values($1,$2,$3,$4,$5,$6)`

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgressStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}
func (s *PostgressStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id=$1", id)
	// fmt.Println("in geta1")
	if err != nil {
		return nil, err
	}
	// fmt.Println("in geta2")
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	// fmt.Println("in geta3")

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgressStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with number %d not found", number)
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)

	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)
	// fmt.Println("in scan")
	return account, err
}

func (s *PostgressStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id=$1", id)
	return err
}
func (s *PostgressStore) UpdateAccount(*Account) error {
	return nil
}
