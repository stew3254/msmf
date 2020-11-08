package main

import (
  "encoding/json"
  "log"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/jinzhu/gorm"
	"github.com/rs/cors"

  _ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
  router := mux.NewRouter() 
	db, err = gorm.Open("postgres", "host=db port=5432 user=postgres dbname=postgres sslmode=disable password=postgres") 

  if err != nil { 
    panic("failed to connect database") 
  } 
	defer db.Close() 

  db.AutoMigrate(&Driver{}) 
	db.AutoMigrate(&Car{}) 

  for index := range cars { 
      db.Create(&cars[index]) 
  }

  for index := range drivers {
      db.Create(&drivers[index])
  }

  router.HandleFunc("/cars", GetCars).Methods("GET")
  router.HandleFunc("/cars/{id}", GetCar).Methods("GET")
  router.HandleFunc("/drivers/{id}", GetDriver).Methods("GET")
	router.HandleFunc("/cars/{id}", DeleteCar).Methods("DELETE")

	handler := cors.Default().Handler(router)

  log.Fatal(http.ListenAndServe(":8080", handler))
}