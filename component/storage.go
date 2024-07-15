package component

import (
	"database/sql"
	"fmt"
	"sort"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account, string, string, int64) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int64) (*Account, error)
	TransferMoney(*TransactionRequest, *Account) error
	CreateTransactionBlock(*Transaction) error
	GetAllTransaction(int64) ([]*Transaction, error)
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
	if err := s.CreateAccountTable(); err != nil {
		return err
	}
	if err := s.CreateTransactionTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgressStore) CreateAccountTable() error {
	query := `create table if not exists account(
	id serial primary key,
	first_name varchar(100),
	last_name varchar(100),
	number serial,
	pin varchar(100),
	encrypted_password varchar(100),
	balance serial,
	created_at timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateTransactionTable() error {
	query := `create table if not exists transaction(
	id serial primary key,
	sender serial,
	reciever serial,
	amount serial,
	time timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgressStore) CreateTransactionBlock(block *Transaction) error {
	query := `insert into transaction
	(sender,reciever,amount,time)
	values($1,$2,$3,$4)`

	_, err := s.db.Query(
		query,
		block.Sender,
		block.Reciever,
		block.Amount,
		block.Time,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgressStore) CreateAccount(acc *Account) error {
	query := `insert into account
	(first_name,last_name,number,pin,encrypted_password,balance,created_at)
	values($1,$2,$3,$4,$5,$6,$7)`

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.PIN,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		fmt.Println("hii")
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

func (s *PostgressStore) GetTransactionAsSender(number int64) ([]*Transaction, error) {
	rows, err := s.db.Query("select * from transaction where sender=$1", number)
	if err != nil {
		return nil, err
	}
	transactions := []*Transaction{}

	for rows.Next() {

		transaction, err := scanIntoTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)

	}
	return transactions, nil
}

func (s *PostgressStore) GetTransactionAsReciever(number int64) ([]*Transaction, error) {
	rows, err := s.db.Query("select * from transaction where reciever=$1", number)
	if err != nil {
		return nil, err
	}

	transactions := []*Transaction{}

	for rows.Next() {

		transaction, err := scanIntoTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)

	}
	return transactions, nil
}

func (s *PostgressStore) GetAccountByNumber(number int64) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with number %d not found", number)
}

func scanIntoTransaction(rows *sql.Rows) (*Transaction, error) {
	transaction := new(Transaction)
	err := rows.Scan(
		&transaction.ID,
		&transaction.Sender,
		&transaction.Reciever,
		&transaction.Amount,
		&transaction.Time)
	return transaction, err
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)

	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.PIN,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)
	// fmt.Println("in scan")
	return account, err
}

func (s *PostgressStore) TransferMoney(tranReq *TransactionRequest, acc *Account) error {
	recAcc, err := s.GetAccountByNumber(tranReq.Reciever)
	if err != nil {
		return fmt.Errorf("invalid transaction")
	}
	tx, err := s.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	recAccBalance := recAcc.Balance + tranReq.Amount
	accBalance := acc.Balance - tranReq.Amount

	err = s.UpdateAccount(acc, acc.FirstName, acc.LastName, accBalance)
	if err != nil {
		return fmt.Errorf("error occured inside1 %s", err)
	}
	err = s.UpdateAccount(recAcc, recAcc.FirstName, recAcc.LastName, recAccBalance)
	if err != nil {
		return fmt.Errorf("error occured inside 2 %s", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error occured inside 3 %s", err)
	}

	return nil

}
func (s *PostgressStore) GetAllTransaction(sender int64) ([]*Transaction, error) {
	transactionS, err1 := s.GetTransactionAsSender(sender)

	transactionR, err2 := s.GetTransactionAsReciever(sender)
	if err1 == nil && err2 == nil {

		return combineAndSortTransactions(transactionR, transactionS), nil
	} else if err1 == nil {
		return transactionS, nil
	} else if err2 == nil {
		return transactionR, nil
	}
	return nil, fmt.Errorf("NO transaction")
}

func (s *PostgressStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id=$1", id)
	return err
}
func (s *PostgressStore) UpdateAccount(acc *Account, firstname, lastname string, balance int64) error {
	query := `update account 
	set
	first_name=$1,
	last_name=$2,
	balance=$3
	where number= $4`

	_, err := s.db.Exec(
		query,
		firstname,
		lastname,
		balance,
		acc.Number,
	)
	if err != nil {
		return fmt.Errorf("invalid updation")
	}
	acc.FirstName = firstname
	acc.LastName = lastname
	acc.Balance = balance

	return nil
}

func combineAndSortTransactions(transactions1, transactions2 []*Transaction) []*Transaction {
	// Combine the slices
	combined := append(transactions1, transactions2...)

	// Sort the combined slice by created time
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Time.Before(combined[j].Time)
	})

	return combined
}
