package main

import (
	"fmt"
	"os"

	"github.com/alex-laycalvert/ghtui/app"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gtui <repo> <token>")
		os.Exit(1)
	}
	repoName := os.Args[1]
	token := os.Args[2]

	app, err := app.New(token, repoName)
	checkErr(err)

	err = app.Run()
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
