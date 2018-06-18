package main

import "errors"

// User Model
type User struct {
	ID       int
	Name     string
	Address  string
	MyNumber string
	Votes    int
}

var userMemo map[string]User

func getUser(name string, address string, myNumber string) (user User, err error) {
	var ok bool
	user, ok = userMemo[myNumber]
	if !ok || user.Name != name || user.Address != address {
		err = errors.New("")
	}
	return
}

func userInitialize() {
	if len(userMemo) == 0 {
		userMemo = make(map[string]User)
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			user := User{}
			err = rows.Scan(&user.ID, &user.Name, &user.Address, &user.MyNumber, &user.Votes)
			if err != nil {
				panic(err)
			}
			userMemo[user.MyNumber] = user
		}
	}
}
