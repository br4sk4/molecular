package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	config        Configuration
	cognitoClient *CognitoClient
	tokenMap      = make(map[string]*CognitoTokens)
)

func main() {
	startServer()
}

func startServer() {
	if err := config.Get(); err != nil {
		log.Fatal(err)
		return
	}

	cognitoClient = &CognitoClient{
		clientID:     config.Cognito.ClientID,
		clientSecret: config.Cognito.ClientSecret,
		client:       cognito.New(cognito.Options{Region: config.Cognito.Region}),
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(indexHandler))

	mux.Handle("/login/ui", http.StripPrefix("/login/ui", http.HandlerFunc(uiLoginHandler)))
	mux.Handle("/login/ui/", http.StripPrefix("/login/ui", http.HandlerFunc(uiLoginHandler)))
	mux.Handle("/login/cli", http.StripPrefix("/login/cli", http.HandlerFunc(cliLoginHandler)))
	mux.Handle("/login/cli/", http.StripPrefix("/login/cli", http.HandlerFunc(cliLoginHandler)))

	mux.Handle("/api/auth", http.StripPrefix("/api/auth", http.HandlerFunc(authEndpointHandler)))
	mux.Handle("/api/auth/", http.StripPrefix("/api/auth", http.HandlerFunc(authEndpointHandler)))
	mux.Handle("/api/token", http.StripPrefix("/api/token", http.HandlerFunc(tokenEndpointHandler)))
	mux.Handle("/api/token/", http.StripPrefix("/api/token", http.HandlerFunc(tokenEndpointHandler)))
	mux.Handle("/api/callback", http.StripPrefix("/api/callback", http.HandlerFunc(callbackEndpointHandler)))
	mux.Handle("/api/callback/", http.StripPrefix("/api/callback", http.HandlerFunc(callbackEndpointHandler)))

	mux.Handle("/dex", http.HandlerFunc(dexProxyHandler))
	mux.Handle("/dex/", http.HandlerFunc(dexProxyHandler))
	mux.Handle("/dex/callback/cli", http.HandlerFunc(dexCognitoAuthProxy))
	mux.Handle("/dex/callback/cli/", http.HandlerFunc(dexCognitoAuthProxy))

	log.Print("Listening on :3000...")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func uiLoginHandler(w http.ResponseWriter, r *http.Request) {
	fileServer := http.FileServer(http.Dir(config.Frontend.Directory))
	fileServer.ServeHTTP(w, r)
}

func cliLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	u := config.Service.MolecularURL + "/dex/auth/cli"
	u += "?redirect_uri=" + url.QueryEscape(config.Service.MolecularURL+"/api/callback")
	u += "&response_type=code"
	u += "&client_id=cognito"
	u += "&state=" + state
	u += "&scope=openid+profile+email+groups+offline_access"

	http.Redirect(w, r, u, http.StatusFound)
}

func authEndpointHandler(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	authResult, err := cognitoClient.Authorize(username, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	authCode := generateAuthCode()
	tokenMap[authCode] = authResult
	time.AfterFunc(5*time.Minute, func() {
		if _, ok := tokenMap[authCode]; ok {
			delete(tokenMap, authCode)
		}
	})

	response := &AuthResponse{Code: authCode}
	responseJson, _ := json.MarshalIndent(response, "", "  ")

	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func tokenEndpointHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tokenRequest, err := getTokenRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch tokenRequest.GrantType {
	case "authorization_code":
		tokens, ok := tokenMap[tokenRequest.Code]
		if ok {
			tokensJson, _ := json.MarshalIndent(tokens, "", "  ")
			delete(tokenMap, tokenRequest.Code)
			w.WriteHeader(http.StatusOK)
			w.Write(tokensJson)
			return
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("authorization code has been claimed by another client or invalidated"))
			return
		}
	case "refresh_token":
		tokens, err := cognitoClient.Refresh(tokenRequest.Username, tokenRequest.RefreshToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tokensJson, _ := json.MarshalIndent(tokens, "", "  ")
		delete(tokenMap, tokenRequest.Code)
		w.WriteHeader(http.StatusOK)
		w.Write(tokensJson)
		return
	default:
		log.Printf("molecular: invalid grant_type: %s", tokenRequest.GrantType)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func callbackEndpointHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	body := []byte("grant_type=authorization_code&client_id=cognito&redirect_uri=" + url.QueryEscape(config.Service.MolecularURL+"/api/callback") + "&code=" + code)
	bodyReader := bytes.NewReader(body)
	request, _ := http.NewRequest(http.MethodPost, config.Service.MolecularURL+"/dex/token", bodyReader)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	originServerResponse, err := client.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	copyHeader(w.Header(), originServerResponse.Header)
	io.Copy(w, originServerResponse.Body)
}

func dexProxyHandler(w http.ResponseWriter, r *http.Request) {
	originServerURL, _ := url.Parse(config.Service.DexURL)

	r.Host = originServerURL.Host
	r.URL.Host = originServerURL.Host
	r.URL.Scheme = originServerURL.Scheme
	r.RequestURI = ""

	proxy := httputil.NewSingleHostReverseProxy(originServerURL)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxy.ServeHTTP(w, r)
}

func dexCognitoAuthProxy(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()
	authCode := generateAuthCode()
	authResult, _ := cognitoClient.Authorize(username, password)

	//r.Header.Del("Authorization")

	tokenMap[authCode] = authResult
	time.AfterFunc(5*time.Minute, func() {
		if _, ok := tokenMap[authCode]; ok {
			delete(tokenMap, authCode)
		}
	})

	originServerURL, _ := url.Parse(config.Service.DexURL + "?" + r.URL.RawQuery + "&code=" + authCode)

	r.Host = originServerURL.Host
	r.URL.Host = originServerURL.Host
	r.URL.Scheme = originServerURL.Scheme
	r.RequestURI = ""

	proxy := httputil.NewSingleHostReverseProxy(originServerURL)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxy.ServeHTTP(w, r)
}

func getTokenRequest(r *http.Request) (*TokenRequest, error) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, err
	}

	var tokenRequest TokenRequest
	if err := json.Unmarshal(requestBody, &tokenRequest); err != nil {
		return nil, err
	}

	return &tokenRequest, nil
}

type AuthResponse struct {
	Code string `json:"code"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Username     string `json:"username,omitempty"`
}
