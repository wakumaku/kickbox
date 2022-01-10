package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"wakumaku/kickbox"
)

const apiKey = "test_my_apikey"

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(os.Args) != 2 {
		log.Println("expecting at least 1 parameter")
		os.Exit(1)
	}

	kbClient, err := kickbox.New(apiKey)
	if err != nil {
		log.Printf("cannot create the verifier instance: %v\n", err)
	}

	email := strings.TrimSpace(os.Args[1])
	stats, resp, err := kbClient.Verify(ctx, email)
	if err != nil {
		log.Printf("error verifying the email '%s': %v\n", email, err)
		os.Exit(1)
	}

	log.Printf("Balance: %d", stats.Balance)
	log.Printf("Time: %d", stats.ResponseTime)
	log.Printf("HTTP Status: %d", stats.HttpStatus)

	response, err := json.MarshalIndent(resp, "    ", "    ")
	if err != nil {
		log.Printf("cannot marshal the response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(response))
}
