package main

import (
	"fmt"
	"os"
	"os/user"
	"scql/repl"
)

func main() {
	if len(os.Args) > 1 {
		filename := os.Args[1]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("Error opening file %s: %s\n", filename, err)
			os.Exit(1)
		}
		defer file.Close()
		repl.Run(file, os.Stdout)
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the SCQL programming language!\n",
		user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
