package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"github.com/qy-gopher/blogo/pkg/route"
)

type ArticlesFormData struct {
	Title  string
	Body   string
	URL    *url.URL
	Errors map[string]string
}

type Article struct {
	Title, Body string
	ID          int64
}

func (a *Article) Delete() (rowsAffected int64, err error) {
	res, err := db.Exec("DELETE FROM articles WHERE id = " + strconv.FormatInt(a.ID, 10))

	if err != nil {
		return 0, err
	}

	if n, _ := res.RowsAffected(); n > 0 {
		return n, nil
	}

	return 0, nil
}

func (a *Article) Link() string {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		checkError(err)
		return ""
	}

	return showURL.String()
}

var router *mux.Router
var db *sql.DB

func initDB() {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "a123456A",
		Addr:                 "192.168.8.107:3306",
		Net:                  "tcp",
		DBName:               "blogo",
		AllowNativePasswords: true,
	}

	db, err = sql.Open("mysql", config.FormatDSN())
	checkError(err)

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Println(err)
	}
}

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
	id := getRouterVariable("id", r)
	article, err := getArticlesByID(id)

	if err != nil {
		getArticlesError(w, err)

	} else {
		tmpl, err := template.New("show.html").Funcs(template.FuncMap{
			"RouteNameToURL": route.NameToURL,
			"Int64ToString":  Int64ToString,
		}).ParseFiles("resources/views/articles/show.html")
		checkError(err)

		err = tmpl.Execute(w, article)
		checkError(err)
	}
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * from articles")
	checkError(err)
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article

		err = rows.Scan(&article.ID, &article.Title, &article.Body)
		checkError(err)

		articles = append(articles, article)
	}

	err = rows.Err()
	checkError(err)

	tmpl, err := template.ParseFiles("resources/views/articles/index.html")
	checkError(err)

	err = tmpl.Execute(w, articles)
	checkError(err)
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprint(w, "表单解析错误")
		return
	}

	title := r.PostForm.Get("title")
	body := r.PostForm.Get("body")

	errors := validateArticleFormData(title, body)

	if len(errors) == 0 {
		LastInsertId, err := saveArticleToDB(title, body)
		if LastInsertId > 0 {
			fmt.Fprint(w, "插入成功，ID为"+strconv.FormatInt(LastInsertId, 10))
		} else {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}

	} else {
		storeURL, _ := router.Get("articles.store").URL()
		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
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

func articlesEditHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouterVariable("id", r)
	article, err := getArticlesByID(id)

	if err != nil {
		getArticlesError(w, err)

	} else {
		updateURL, _ := router.Get("articles.update").URL("id", id)
		data := ArticlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}

		tmpl, err := template.ParseFiles("resources/views/articles/edit.html")
		checkError(err)

		err = tmpl.Execute(w, data)
		checkError(err)
	}
}

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouterVariable("id", r)
	_, err := getArticlesByID(id)

	if err != nil {
		getArticlesError(w, err)

	} else {
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := validateArticleFormData(title, body)

		if len(errors) == 0 {
			query := "UPDATE articles SET title = ?, body = ? WHERE id = ?"

			res, err := db.Exec(query, title, body, id)
			if err != nil {
				checkError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "500 服务器内部错误")
			}

			if n, _ := res.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)

			} else {
				fmt.Fprint(w, "未做任何更改")
			}

		} else {
			updateURL, _ := router.Get("articles.update").URL("id", id)
			data := ArticlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateURL,
				Errors: errors,
			}

			tmpl, err := template.ParseFiles("resources/views/articles/edit.html")
			checkError(err)

			err = tmpl.Execute(w, data)
			checkError(err)
		}
	}
}

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := getRouterVariable("id", r)

	article, err := getArticlesByID(id)
	if err != nil {
		getArticlesError(w, err)

	} else {
		rowsAffected, err := article.Delete()

		if err != nil {
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")

		} else {
			if rowsAffected > 0 {
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)

			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}
}

func getRouterVariable(param string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[param]
}

func getArticlesByID(id string) (Article, error) {
	article := Article{}
	query := "SELECT * FROM articles WHERE id = ?"

	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)

	return article, err
}

func getArticlesError(w http.ResponseWriter, err error) {
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 文章未找到")

	} else {
		checkError(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500 服务器内部错误")
	}
}

func validateArticleFormData(title string, body string) map[string]string {
	errors := make(map[string]string)

	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需在3-40个字符之间"
	}

	if body == "" {
		errors["body"] = "内容不能为空"

	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于10的字符"
	}

	return errors
}

func saveArticleToDB(title string, body string) (int64, error) {
	var (
		id   int64
		err  error
		res  sql.Result
		stmt *sql.Stmt
	)

	stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?,?)")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	res, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	if id, err = res.LastInsertId(); id > 0 {
		return id, nil
	}

	return 0, err
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

func createTables() {
	creatArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
    id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
    title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    body longtext COLLATE utf8mb4_unicode_ci
);`

	_, err := db.Exec(creatArticlesSQL)
	checkError(err)
}

func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

func main() {
	initDB()
	createTables()

	route.Initialize()
	router = route.Router

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHanher).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	router.Use(forceHTMLMiddleware)

	err := http.ListenAndServe(":9090", removeTrailingSlash(router))
	if err != nil {
		panic(err)
	}
}
