package main

import (
	"fmt"
	"log"
	"net/http" 
)

var (
	address string
)

func startHTTPServer() {
	http.HandleFunc("/hello", httpHandler)

	log.Fatal(http.ListenAndServe(address, nil))

}


func httpHandler(w http.ResponseWriter, request *http.Request) {
	err := request.ParseForm() 
	if err != nil{
	       panic(err)
	}
	log.Println(request.Form)
	params := request.PostFormValue("params")
	fmt.Fprintf(w, "params, %q", params)
}