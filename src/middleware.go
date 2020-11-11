package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

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
		db.Create(&WebLog{
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