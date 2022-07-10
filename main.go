package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello blogo!")
}

func main() {
	http.HandleFunc("/", handlerFunc)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		panic(err)
	}
}
