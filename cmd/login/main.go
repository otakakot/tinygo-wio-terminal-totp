package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/otakakot/tinygo-wio-terminal-totp/internal/auth0"
)

func main() {
	clientID := flag.String("client-id", "", "Auth0 Client ID (required)")
	clientSecret := flag.String("client-secret", "", "Auth0 Client Secret (required)")
	audience := flag.String("audience", "", "Auth0 API Audience (required)")
	domain := flag.String("domain", "", "Auth0 Domain (required)")
	username := flag.String("username", "", "Username/Email (required)")
	password := flag.String("password", "", "Password (required)")
	otp := flag.String("otp", "", "6-digit TOTP One-Time Password (required)")

	flag.Parse()

	if *clientID == "" || *clientSecret == "" || *audience == "" || *domain == "" || *username == "" || *password == "" || *otp == "" {
		flag.Usage()
		return
	}

	ctx := context.Background()

	mfaToken, err := auth0.GetMFAToken(
		ctx,
		*username,
		*password,
		*clientID,
		*clientSecret,
		*audience,
		*domain,
	)
	if err != nil {
		panic(err)
	}

	tokens, err := auth0.Login(
		ctx,
		mfaToken,
		*otp,
		*clientID,
		*clientSecret,
		*domain,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("ID Token:", tokens.IDToken)
	fmt.Println("Access Token:", tokens.AccessToken)
}
