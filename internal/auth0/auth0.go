package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GetMFATokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	MFAToken         string `json:"mfa_token"`
}

func GetMFAToken(
	ctx context.Context,
	username string,
	password string,
	clientID string,
	clientSecret string,
	audience string,
	domain string,
) (string, error) {
	values := url.Values{}
	values.Set("grant_type", "password")
	values.Set("username", username)
	values.Set("password", password)
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("audience", audience)
	values.Set("scope", "openid")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, domain+"/oauth/token", strings.NewReader(values.Encode()))

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusForbidden {
		return "", fmt.Errorf("expected status 403 Forbidden, got %d", res.StatusCode)
	}

	response := GetMFATokenErrorResponse{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	if response.Error != "mfa_required" {
		return "", fmt.Errorf("expected error 'mfa_required', got '%s'", response.Error)
	}

	return response.MFAToken, nil
}

type GenerateSecretRequest struct {
	AuthenticatorTypes []string `json:"authenticator_types"`
}

type GenerateSecretResponse struct {
	AuthenticatorType string   `json:"authenticator_type"`
	Secret            string   `json:"secret"`
	BarcodeURI        string   `json:"barcode_uri"`
	RecoveryCodes     []string `json:"recovery_codes"`
}

func GenerateSecret(ctx context.Context, domain string, mfaToken string) (string, error) {
	request := GenerateSecretRequest{
		AuthenticatorTypes: []string{"otp"},
	}

	buf := bytes.NewBuffer(nil)

	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, domain+"/mfa/associate", buf)
	if err != nil {
		return "", err
	}

	req.Header.Add("authorization", "Bearer "+mfaToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)

		fmt.Println("Response Body:", string(body))

		return "", fmt.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	response := GenerateSecretResponse{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}

	fmt.Println("Authenticator Type:", response.AuthenticatorType)
	fmt.Println("Secret:", response.Secret)
	fmt.Println("Barcode URI:", response.BarcodeURI)
	fmt.Println("Recovery Codes:", response.RecoveryCodes)

	return response.Secret, nil
}

type LoginResponse struct {
	IDToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func Login(
	ctx context.Context,
	mfaToken string,
	otp string,
	clientID string,
	clientSecret string,
	domain string,
) (LoginResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "http://auth0.com/oauth/grant-type/mfa-otp")
	values.Set("client_id", clientID)
	values.Set("mfa_token", mfaToken)
	values.Set("client_secret", clientSecret)
	values.Set("otp", otp)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, domain+"/oauth/token", strings.NewReader(values.Encode()))
	if err != nil {
		return LoginResponse{}, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return LoginResponse{}, fmt.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	response := LoginResponse{}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return LoginResponse{}, err
	}

	return response, nil
}

func GetAuthenticator(
	ctx context.Context,
	domain string,
	mfaToken string,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, domain+"/mfa/authenticators", nil)
	if err != nil {
		return err
	}

	req.Header.Add("authorization", "Bearer "+mfaToken)
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	println(string(body))

	return nil
}
