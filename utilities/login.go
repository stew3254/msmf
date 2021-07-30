package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Auth struct {
	Username string
	Password string
}

func Login(client *http.Client, loginUrl string, auth Auth) (token *http.Cookie, err error) {
	// Create the form
	form := url.Values{}
	form.Add("username", auth.Username)
	form.Add("password", auth.Password)

	// Create new post request
	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	// Add relevant headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	// Look at the cookies
	cookies := resp.Cookies()
	// Set the cookies to the jar in case the client gets used further
	if client.Jar != nil {
		client.Jar.SetCookies(req.URL, cookies)
	}

	for _, cookie := range cookies {
		// We got the token cookie
		if cookie.Name == "token" {
			return cookie, nil
		}
	}

	return nil, errors.New("token cookie not found")
}
