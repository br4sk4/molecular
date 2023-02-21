package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/cristalhq/jwt/v4"
	"io"
	"net/http"
	"time"
)

type JWK struct {
	Keys []struct {
		KeyType   string `json:"kty"`
		KeyID     string `json:"kid"`
		Algorythm string `json:"alg"`
		E         string `json:"e"`
		N         string `json:"n"`
		Usage     string `json:"use"`
	} `json:"keys"`
}

type AuthenticationResult struct {
	TokenType    string
	ExpiresIn    int32
	AccessToken  string
	IdToken      string
	RefreshToken string
}

func authorize(username, password string) (*AuthenticationResult, error) {
	cidp := cognito.New(cognito.Options{Region: "eu-central-1"})
	clientId := "4rnf4s7j8c1tm7rjal6al951as"
	authParams := make(map[string]string)
	authParams["USERNAME"] = username
	authParams["PASSWORD"] = password

	response, err := cidp.InitiateAuth(context.Background(), &cognito.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       &clientId,
		AuthParameters: authParams,
	})

	if err != nil {
		return nil, err
	}

	return &AuthenticationResult{
		TokenType:    *response.AuthenticationResult.TokenType,
		ExpiresIn:    response.AuthenticationResult.ExpiresIn,
		AccessToken:  *response.AuthenticationResult.AccessToken,
		IdToken:      *response.AuthenticationResult.IdToken,
		RefreshToken: *response.AuthenticationResult.RefreshToken,
	}, nil
}

func refresh(refreshToken string) (*AuthenticationResult, error) {
	cidp := cognito.New(cognito.Options{Region: "eu-central-1"})
	clientId := "4rnf4s7j8c1tm7rjal6al951as"
	authParams := make(map[string]string)
	authParams["REFRESH_TOKEN"] = refreshToken

	response, err := cidp.InitiateAuth(context.Background(), &cognito.InitiateAuthInput{
		AuthFlow:       "REFRESH_TOKEN_AUTH",
		ClientId:       &clientId,
		AuthParameters: authParams,
	})

	if err != nil {
		return nil, err
	}

	return &AuthenticationResult{
		TokenType:    *response.AuthenticationResult.TokenType,
		ExpiresIn:    response.AuthenticationResult.ExpiresIn,
		AccessToken:  *response.AuthenticationResult.AccessToken,
		IdToken:      *response.AuthenticationResult.IdToken,
		RefreshToken: refreshToken,
	}, nil
}

func isValidToken(token string) (bool, error) {
	verifiedToken, err := parseToken(token)
	if err != nil {
		return false, err
	}

	var claims jwt.RegisteredClaims
	if unmarshalError := json.Unmarshal(verifiedToken.Claims(), &claims); unmarshalError != nil {
		return false, unmarshalError
	}

	return claims.IsValidAt(time.Now()), nil
}

func parseToken(token string) (*jwt.Token, error) {
	parsedToken, parseError := jwt.ParseNoVerify([]byte(token))
	if parseError != nil {
		return nil, parseError
	}

	var claims jwt.RegisteredClaims
	if unmarshalError := json.Unmarshal(parsedToken.Claims(), &claims); unmarshalError != nil {
		return nil, unmarshalError
	}

	response, httpError := http.DefaultClient.Get(claims.Issuer + "/.well-known/jwks.json")
	if httpError != nil {
		return nil, httpError
	}

	body, httpError := io.ReadAll(response.Body)
	if httpError != nil {
		return nil, httpError
	}

	var jwk JWK
	if unmarshalError := json.Unmarshal(body, &jwk); unmarshalError != nil {
		return nil, unmarshalError
	}

	keyMap := make(map[string]*rsa.PublicKey, 0)

	for i, v := range jwk.Keys {
		publicKey := convertKey(v.E, v.N)
		keyMap[jwk.Keys[i].KeyID] = publicKey
	}

	verifier, verificationError := jwt.NewVerifierRS(jwt.RS256, keyMap[parsedToken.Header().KeyID])
	if verificationError != nil {
		return nil, verificationError
	}

	verifiedToken, verificationError := jwt.Parse([]byte(token), verifier)
	if verificationError != nil {
		return nil, verificationError
	}

	return verifiedToken, nil
}
