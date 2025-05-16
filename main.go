package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

func main() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		panic("CLIENT_ID and CLIENT_SECRET must be set")
	}

	loginServer := "login.eveonline.com"
	scopes := []string{
		"esi-markets.read_character_orders.v1",
		"esi-characters.read_blueprints.v1",
		"esi-assets.read_assets.v1",
		"esi-universe.read_structures.v1",
		"esi-planets.manage_planets.v1",
		"esi-wallet.read_character_wallet.v1",
	}
	redirectURL := "http://localhost:8080/oauth/callback"
	state := uuid.New().String()

	go func() {
		// run web server to listen for callback
		http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
			callbackCode := r.URL.Query().Get("code")
			if callbackCode == "" {
				http.Error(w, "code not found", http.StatusBadRequest)
				return

			}
			callbackState := r.URL.Query().Get("state")
			if callbackState == "" {
				http.Error(w, "state not found", http.StatusBadRequest)
				return
			}

			if state != callbackState {
				http.Error(w, "state mismatch", http.StatusBadRequest)
				return
			}

			// exchange code for token
			tokenURL := fmt.Sprintf("https://%s/v2/oauth/token", loginServer)
			requestBody := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s", callbackCode))
			req, err := http.NewRequest("POST", tokenURL, requestBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req.SetBasicAuth(clientID, clientSecret)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Printf("token response: %s\n", string(body))

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("failed to get token: %d\n", resp.StatusCode)
				http.Error(w, "failed to get token", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "token received\n")

			tokenResponse := struct {
				AccessToken  string `json:"access_token"`
				TokenType    string `json:"token_type"`
				ExpiresIn    int    `json:"expires_in"`
				RefreshToken string `json:"refresh_token"`
			}{}

			err = json.Unmarshal(body, &tokenResponse)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Printf("access token: %s\n", tokenResponse.AccessToken)
			fmt.Printf("refresh token: %s\n", tokenResponse.RefreshToken)

			// get character info
			characterURL := "https://esi.evetech.net/verify/"
			req, err = http.NewRequest("GET", characterURL, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))

			resp, err = client.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("failed to get character: %d\n", resp.StatusCode)
				http.Error(w, "failed to get character", http.StatusInternalServerError)
				return
			}

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			characterResponse := struct {
				CharacterID        int    `json:"CharacterID"`
				CharacterName      string `json:"CharacterName"`
				ExpiresOn          string `json:"ExpiresOn"`
				Scopes             string `json:"Scopes"`
				TokenType          string `json:"TokenType"`
				CharacterOwnerHash string `json:"CharacterOwnerHash"`
			}{}

			err = json.Unmarshal(body, &characterResponse)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Printf("character id: %d\n", characterResponse.CharacterID)
			fmt.Printf("character name: %s\n", characterResponse.CharacterName)
			fmt.Printf("expires on: %s\n", characterResponse.ExpiresOn)
			fmt.Printf("scopes: %s\n", characterResponse.Scopes)
			fmt.Printf("token type: %s\n", characterResponse.TokenType)
			fmt.Printf("character owner hash: %s\n", characterResponse.CharacterOwnerHash)

			w.Write([]byte("character info received"))
		})

		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Printf("In a browser open the following url:\n")
	fmt.Printf("https://%s/v2/oauth/authorize?response_type=code&redirect_uri=%s&client_id=%s&scope=%s&state=%s",
		loginServer, redirectURL, clientID, strings.Join(scopes, " "), state)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	for {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Println("sigint")
			return
		case syscall.SIGTERM:
			fmt.Println("sigterm")
			return
		}
	}
}
