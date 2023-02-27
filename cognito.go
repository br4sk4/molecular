package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type CognitoTokens struct {
	TokenType    string
	ExpiresIn    int32
	AccessToken  string
	IdToken      string
	RefreshToken string
}

type CognitoClient struct {
	clientID     string
	clientSecret string
	client       *cognito.Client
}

func (c *CognitoClient) Authorize(username, password string) (*CognitoTokens, error) {
	authParams := make(map[string]string)
	authParams["USERNAME"] = username
	authParams["PASSWORD"] = password
	authParams["SECRET_HASH"] = c.generateSecretHash(username)

	response, err := c.client.InitiateAuth(context.Background(), &cognito.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       &c.clientID,
		AuthParameters: authParams,
	})

	if err != nil {
		return nil, err
	}

	return &CognitoTokens{
		TokenType:    *response.AuthenticationResult.TokenType,
		ExpiresIn:    response.AuthenticationResult.ExpiresIn,
		AccessToken:  *response.AuthenticationResult.AccessToken,
		IdToken:      *response.AuthenticationResult.IdToken,
		RefreshToken: *response.AuthenticationResult.RefreshToken,
	}, nil
}

func (c *CognitoClient) Refresh(username, refreshToken string) (*CognitoTokens, error) {
	authParams := make(map[string]string)
	authParams["REFRESH_TOKEN"] = refreshToken
	authParams["SECRET_HASH"] = c.generateSecretHash(username)

	response, err := c.client.InitiateAuth(context.Background(), &cognito.InitiateAuthInput{
		AuthFlow:       "REFRESH_TOKEN_AUTH",
		ClientId:       &c.clientID,
		AuthParameters: authParams,
	})

	if err != nil {
		return nil, err
	}

	return &CognitoTokens{
		TokenType:    *response.AuthenticationResult.TokenType,
		ExpiresIn:    response.AuthenticationResult.ExpiresIn,
		AccessToken:  *response.AuthenticationResult.AccessToken,
		IdToken:      *response.AuthenticationResult.IdToken,
		RefreshToken: refreshToken,
	}, nil
}

func (c *CognitoClient) generateSecretHash(username string) string {
	h := hmac.New(sha256.New, []byte(c.clientSecret))
	h.Write([]byte(username + c.clientID))
	secretHash := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return secretHash
}
