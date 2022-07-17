package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, "<h1>Hello blogo</h1>")
}

func aboutHanher(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "test/html; charset=utf-8")

	fmt.Fprint(w, "<h1>blogo可以用来记录和分享信息</h1>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.WriteHeader(http.StatusNotFound)

	fmt.Fprint(w, "<h1>请求页面未找到</h1>")
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	fmt.Fprintf(w, "文章ID: %s", id)
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "文章列表页")
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "创建文章页")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHanher).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	err := http.ListenAndServe(":9090", router)
	if err != nil {
		panic(err)
	}
}
