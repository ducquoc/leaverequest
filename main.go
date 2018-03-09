package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"os"
)

func PostLeaveRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	fmt.Println(string(body), "POST done")
}

func GetLeaveRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Println("POST done")
}

func HandleRequest() {
	r := mux.NewRouter()
	port := os.Getenv("PORT")
	r.HandleFunc("/lq", PostLeaveRequestHandler).Methods("POST")
	r.HandleFunc("/lq", GetLeaveRequestHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":" + port, r))
}

func main() {
	HandleRequest()
}
