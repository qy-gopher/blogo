package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

type ArticlesFormData struct {
	Title  string
	Body   string
	URL    *url.URL
	Errors map[string]string
}

var router = mux.NewRouter()

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello blogo</h1>")
}

func aboutHanher(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>blogo可以用来记录和分享信息</h1>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
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
	err := r.ParseForm()
	if err != nil {
		fmt.Fprint(w, "表单解析错误")
		return
	}

	title := r.PostForm.Get("title")
	body := r.PostForm.Get("body")

	errs := make(map[string]string)

	titleLen := utf8.RuneCountInString(title)
	bodyLen := utf8.RuneCountInString(body)

	if titleLen <= 2 || titleLen >= 40 {
		errs["title"] = "文章标题要在2到40个字符之间"
	}
	if bodyLen <= 10 {
		errs["body"] = "文章内容要大于10个字符"
	}

	if len(errs) == 0 {
		fmt.Fprintf(w, "title的值为: %v <br>", title)
		fmt.Fprintf(w, "body的值为: %v", body)
	} else {
		storeURL, _ := router.Get("articles.store").URL()
		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errs,
		}

		tmpl, err := template.ParseFiles("resources/views/articles/create.html")
		if err != nil {
			panic(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	}
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	storeURL, _ := router.Get("articles.store").URL()

	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}

	tmpl, err := template.ParseFiles("resources/views/articles/create.html")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 移除首页外的请求路径最后的'/'
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// router := mux.NewRouter()

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHanher).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	router.Use(forceHTMLMiddleware)

	err := http.ListenAndServe(":9090", removeTrailingSlash(router))
	if err != nil {
		panic(err)
	}
}
