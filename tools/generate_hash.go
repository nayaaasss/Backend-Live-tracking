package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "102070"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash password admin:", string(hash))
}
