package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	secret := flag.String("secret", "", "TOTP Secret (required)")

	flag.Parse()

	if *secret == "" {
		flag.Usage()
		return
	}

	fmt.Printf("TOTP Generator started with secret\n")
	fmt.Printf("Generating tokens every 30 seconds...\n\n")

	generateAndPrintToken(*secret)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		generateAndPrintToken(*secret)
	}
}

func generateAndPrintToken(secret string) {
	token, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		log.Printf("Error generating token: %v\n", err)
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, token)
}
