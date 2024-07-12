package main

import (
	// "fmt"
	// "flag"
	// "fmt"
	"log"

	"github.com/harshgupta9473/goBank/component"
)


// func seedAccount(store component.Storage,fname,lname,pw string)*component.Account{
// 	acc,err:=component.NewAccount(fname,lname,pw)
// 	if err!=nil{
// 		log.Fatal(err)
// 	}
// 	if err:=store.CreateAccount(acc); err!=nil{
// 		log.Fatal(err)
// 	}
// 	return acc

// }

// func seedAccounts(s component.Storage){
// 	seedAccount(s,"harshu","gupta","kash232323")
// }


func main() {
	// seed:=flag.Bool("seed",false,"seed the db")
	// flag.Parse()
	store, err := component.NewPostgressStore()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", store)
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
// if *seed {
// 	fmt.Println("seeding the database")
// 	seedAccounts(store)
// }
	// seed stufff
	

	server := component.NewAPIServer(":3000", store)
	server.Run()
}
