package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"msmf/database"
)

func printPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonOut, err := json.Marshal(r.URL.Query())
		if err != nil {
			log.Println(err)
			return
		}
		queryParams := string(jsonOut)
		
		r.ParseForm()
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
			Time: time.Now(),
			IP: strings.Split(r.RemoteAddr, ":")[0],
			Method: r.Method,
			QueryParams: queryParams,
			PostData: postData,
			Cookies: cookies,
		})
		
		next.ServeHTTP(w, r)
	})
}

// Good enough for now to check which routes will accept unauthenticated requests
func checkvalidUnauthenticatedRoutes(url string) bool {
	return (!strings.HasSuffix(url, ".css") && !strings.HasSuffix(url, ".js") && !strings.HasSuffix(url, ".map") && url != "/" && url != "/login")
}

// Checks to see if a user is authenticated to a page before displaying
// If they aren't authenticated, they will be redirected to the login page
func checkAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkvalidUnauthenticatedRoutes(r.URL.String()) {
			tokenCookie, err:= r.Cookie("token")
			// Cookie not found
			if err != nil {
				// http.Redirect(w, r, "/login", http.StatusFound)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// See if user exists
			user := database.User{}
			result :=  database.DB.Where("token = ?", tokenCookie.Value).First(&user)
			if result.Error != nil || time.Now().After(user.TokenExpiration) {
				// http.Redirect(w, r, "/login", http.StatusFound)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

		}
		next.ServeHTTP(w, r)
	})
}