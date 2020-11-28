package src

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"msmf/database"
	"msmf/utils"
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
func checkValidUnauthenticatedRoutes(url string) bool {
	return (strings.HasSuffix(url, ".css") || strings.HasSuffix(url, ".js") || strings.HasSuffix(url, ".map") || url == "/" || url == "/login")
}

// Checks to see if a user is authenticated to a page before displaying
// If they aren't authenticated, they will be redirected to the login page
func checkAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		// Cookie not found
		if err != nil && !checkValidUnauthenticatedRoutes(r.URL.String()) {
			// http.Redirect(w, r, "/login", http.StatusFound)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if r.URL.String() == "/login" {
			// All good, no need to log in again
			if err != nil || !utils.ValidateToken(tokenCookie.Value) {
				next.ServeHTTP(w, r)
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
		next.ServeHTTP(w, r)
	})
}
