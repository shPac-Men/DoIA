package main

import (
	"AwsProj/internal/api"
	"fmt"
	"log"
	"os"
)

func main() {
	dir, _ := os.Getwd()
	fmt.Println("Current directory:", dir)

	// Проверяем существует ли папка
	if _, err := os.Stat("../../resources"); os.IsNotExist(err) {
		fmt.Println("ERROR: resources folder not found!")
	} else {
		fmt.Println("Resources folder found")
	}

	log.Println("app start")
	api.StartServer()
	log.Println("app down")
}
