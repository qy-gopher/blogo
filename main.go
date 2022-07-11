package main

import (
	"fmt"
	"net/http"
)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fmt.Fprint(w, "<h1>hello blogo</h1>")
	case "/about":
		fmt.Fprint(w, "<h1>blogo可以用来分享信息</h1>")
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>页面未找到: %s</h1>", r.Host+r.URL.Path)
	}
}

func main() {
	http.HandleFunc("/", handlerFunc)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		panic(err)
	}
}
