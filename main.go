package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	tokenMap = make(map[string]*AuthenticationResult)
	config   Configuration
)

func main() {
	startServer()
}

func startServer() {
	if err := config.Get(); err != nil {
		log.Fatal(err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/login/", http.StripPrefix("/login", http.FileServer(http.Dir(config.Frontend.Directory))))
	mux.Handle("/api/authorize/", http.StripPrefix("/api/authorize", http.HandlerFunc(authorizeHandler)))
	mux.Handle("/api/token/", http.StripPrefix("/api/token", http.HandlerFunc(tokenHandler)))
	mux.Handle("/api/callback/", http.StripPrefix("/api/callback", http.HandlerFunc(callbackHandler)))

	log.Print("Listening on :3000...")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	authCode := generateAuthCode()
	authResult, _ := authorize(username, password)

	r.Header.Del("Authorization")

	tokenMap[authCode] = authResult
	time.AfterFunc(5*time.Minute, func() {
		if _, ok := tokenMap[authCode]; ok {
			delete(tokenMap, authCode)
		}
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"code\": \"" + authCode + "\"}"))
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	authCode := r.URL.Query().Get("code")

	tokens, ok := tokenMap[authCode]

	if ok {
		tokensJson, _ := json.MarshalIndent(tokens, "", "  ")
		delete(tokenMap, authCode)
		w.WriteHeader(http.StatusOK)
		w.Write(tokensJson)
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("authorization code has been claimed by another client or invalidated"))
	}
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	body := []byte("grant_type=authorization_code&client_id=cognito&client_secret=123456&redirect_uri=https%3A%2F%2Fsso.k8s.naffets.local%2Fapi%2Fcallback&code=" + code)
	bodyReader := bytes.NewReader(body)
	request, _ := http.NewRequest(http.MethodPost, "https://dex.k8s.naffets.local/dex/token", bodyReader)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	originServerResponse, err := client.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, err)
		return
	}

	copyHeader(w.Header(), originServerResponse.Header)
	io.Copy(w, originServerResponse.Body)
}
