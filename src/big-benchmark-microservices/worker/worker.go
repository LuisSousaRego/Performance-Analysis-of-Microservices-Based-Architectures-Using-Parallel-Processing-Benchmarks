package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const pingLimit = 5000000

type PingMessage struct {
	Id int `json:"id"`
}

func ping(w http.ResponseWriter, req *http.Request) {
	var p PingMessage
	err := json.NewDecoder(req.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(p)
}

func pong(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "pong received\n")
}

func main() {

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/pong", pong)

	http.ListenAndServe(":8090", nil)
}
