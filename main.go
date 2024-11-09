package main

import (
	"fmt"
	"timeTableBot/db"
)

func main() {
	database, err := db.InitDB()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	// Добавление нового пользователя
	err = db.CreateUser(database, "newuser", "securepassword")
	if err != nil {
		fmt.Println("Error creating user:", err)
		return
	}

	fmt.Println("User created successfully!")
}
