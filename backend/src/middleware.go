package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"msmf/database"
	"msmf/utils"
)

// printPath prints the HTTP method and request URI to the screen
func printPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// logRequest Logs the request to a database
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonOut, err := json.Marshal(r.URL.Query())
		if err != nil {
			log.Println(err)
			return
		}
		queryParams := string(jsonOut)

		_ = r.ParseForm()
		jsonOut, err = json.Marshal(r.PostForm)
		if err != nil {
			log.Println(err)
			return
		}
		postData := string(jsonOut)

		jsonOut, err = json.Marshal(r.Cookies())
		if err != nil {
			log.Println(err)
			return
		}
		cookies := string(jsonOut)

		// Not complete yet
		database.DB.Create(&database.WebLog{
			Time:        time.Now(),
			IP:          strings.Split(r.RemoteAddr, ":")[0],
			Method:      r.Method,
			QueryParams: queryParams,
			PostData:    postData,
			Cookies:     cookies,
		})

		next.ServeHTTP(w, r)
	})
}

// checkValidUnauthenticatedRoutes simple function to return whether a route needs auth or not
func checkValidUnauthenticatedRoutes(url string) bool {
	if strings.HasSuffix(url, ".css") ||
		strings.HasSuffix(url, ".css") ||
		strings.HasSuffix(url, ".js") ||
		strings.HasSuffix(url, ".map") ||
		(strings.HasPrefix(url, "/api/refer/") && len(url) == len("/api/refer/12345678")) ||
		url == "/" ||
		url == "/login" {
		return true
	}
	return false
}

// checkAuthenticated Checks to see if a user is authenticated to a page before displaying
func checkAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		// Cookie not found
		if err != nil && !checkValidUnauthenticatedRoutes(r.URL.String()) {
			// http.Redirect(w, r, "/login", http.StatusFound)
			http.Error(w, "401 unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.String() == "/login" {
			// All good, no need to log in again
			if err != nil || !utils.ValidateToken(tokenCookie.Value) {
				next.ServeHTTP(w, r)
				return
			} else {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		} else if checkValidUnauthenticatedRoutes(r.URL.String()) {
			next.ServeHTTP(w, r)
			return
		} else if !utils.ValidateToken(tokenCookie.Value) {
			// http.Redirect(w, r, "/login", http.StatusFound)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Increase token expiration if it needs it
		var user database.User
		database.DB.Where("token = ?", tokenCookie.Value).Find(&user)
		now := time.Now()

		// See if the token expires in less than 2 hours
		if now.Add(2 * time.Hour).After(user.TokenExpiration) {
			// Give them another 6 hours from now if that's the case
			user.TokenExpiration = now.Add(6 * time.Hour)
			database.DB.Save(&user)
			http.SetCookie(w, &http.Cookie{
				Path:     "/",
				Name:     "token",
				Value:    user.Token,
				Expires:  user.TokenExpiration,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}

		next.ServeHTTP(w, r)
	})
}
