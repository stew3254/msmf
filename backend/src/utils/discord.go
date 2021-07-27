package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"msmf/database"
	"net/http"
	"time"
)

type Response struct {
	Message    string
	RetryAfter float64
	global     bool
}

// newRequest makes the request object and reader
func newRequest(url string, body *map[string]string) (req *http.Request, err error) {
	// Create a reader for the body
	reader := bytes.NewReader(ToJSON(body))

	// Create the request
	req, err = http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func sendRequest(client *http.Client, url string, body *map[string]string) error {
	// Loop until we get a good status code
	for {
		req, err := newRequest(url, body)

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		// This is good
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		// If sending things way too fast
		if resp.StatusCode == http.StatusTooManyRequests {
			// Get JSON of body
			data := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				return err
			}

			log.Println("Wait for", data["retry_after"])
			time.Sleep(time.Duration(data["retry_after"].(float64)) * time.Millisecond)
		}
	}
}

// SendWebhook will send over the data to Discord
func SendWebhook(integration database.DiscordIntegration, pipes PipeChans) {
	// Create a new http client
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Forever look for messages
	for {
		body := make(map[string]string)
		// Add avatar if we have it
		if integration.AvatarURL != nil {
			body["avatar_url"] = *integration.AvatarURL
		}
		// Add username if we have it
		if integration.Username != nil {
			body["username"] = *integration.Username
		}

		var data []byte
		var ok bool

		// Get data from the channels
		select {
		case data, ok = <-pipes.StdoutChan:
			body["content"] = string(data)
		case data, ok = <-pipes.StderrChan:
			body["content"] = "ERROR: " + string(data)
		}
		// Our pipes are closed, so break out
		if !ok {
			return
		}

		// Send the request
		err := sendRequest(client, integration.DiscordURL, &body)
		if err != nil {
			log.Println("Webhook error:", err)
			return
		}
	}
}
