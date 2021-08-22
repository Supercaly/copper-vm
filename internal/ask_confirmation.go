package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Asks the user a given question until it responds yes or no.
// Returns true if it responded yes and false if no.
func AskConfirmation(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", question)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			fmt.Println("Please answer y or n.")
		}
	}
}
