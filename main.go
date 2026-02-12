package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
)

type PageData struct {
	Error   string
	Success string
}

const (
	keycloakBaseURL = "https://ourden-api.site" // change if needed
	realm           = "auth_distributed"
	clientID        = "registration_system"
	clientSecret    = "MIqONtqJXVplxJISnOG18nRRf4UtZW4i" // <-- replace
)

func main() {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{}

		if r.Method == http.MethodPost {

			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirm := r.FormValue("confirm_password")

			if username == "" || email == "" || password == "" {
				data.Error = "All fields are required"

			} else if password != confirm {
				data.Error = "Passwords do not match"

			} else {

				// 🔥 PUT IT HERE
				token, err := getAdminToken()
				if err != nil {
					data.Error = "Failed to connect to auth server"
				} else {
					err = createUser(token, username, email, password)
					if err != nil {
						data.Error = err.Error()
					} else {
						data.Success = "Account created successfully!"
					}
				}
			}
		}

		tmpl.Execute(w, data)
	})

	log.Println("Server running on port 8080")

	// 🔴 THIS is the key fix
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func getAdminToken() (string, error) {

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.PostForm(
		keycloakBaseURL+"/realms/"+realm+"/protocol/openid-connect/token",
		data,
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get access token")
	}

	return token, nil
}
func createUser(token, username, email, password string) error {

	user := map[string]interface{}{
		"username": username,
		"email":    email,
		"enabled":  true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     password,
				"temporary": false,
			},
		},
	}

	jsonData, _ := json.Marshal(user)

	req, _ := http.NewRequest(
		"POST",
		keycloakBaseURL+"/admin/realms/"+realm+"/users",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)

		var errResp map[string]interface{}
		json.Unmarshal(body, &errResp)

		// Case 1: Email already exists
		if msg, ok := errResp["errorMessage"].(string); ok {
			if msg == "User exists with same email" {
				return fmt.Errorf("Email is already registered")
			}
		}

		// Case 2: Field validation error
		if field, ok := errResp["field"].(string); ok {
			if errMsg, ok := errResp["errorMessage"].(string); ok {
				if errMsg == "error-invalid-length" {
					if field == "username" {
						return fmt.Errorf("username must be at least 3 characters, no space")
					}
				}
				return fmt.Errorf("%s must be at least 3 characters, no space", field)
			}
		}

		return fmt.Errorf("Registration failed")
	}

	return nil
}
