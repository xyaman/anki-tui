package main

import (
	"fmt"
)

func main() {


	client := NewAnkiConnect("http://localhost:8765", 6)
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	app, err := NewApp(config, client)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := app.tviewApp.Run(); err != nil {
		fmt.Println("Error: ", err)
		return
	}
}
