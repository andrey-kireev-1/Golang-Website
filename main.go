package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"goproj1/cmd/web"
	"goproj1/pkg/config"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func init() {
	gob.Register(web.ViewData{})
	gob.Register(web.User{})
	gob.Register(web.Article{})
}

func main() {
	var err error
	connStr := config.ConnectDB()
	web.Db, err = sql.Open("postgres", connStr)
	web.CheckError(err)
	defer web.Db.Close()
	err = web.Db.Ping()
	fmt.Println("Connected to database!")
	web.CheckError(err)

	r := mux.NewRouter()
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/srch", web.Search).Methods("GET")
	r.HandleFunc("/", web.Like).Methods("POST")
	r.HandleFunc("/", web.StartPage)
	r.HandleFunc("/page/{page_id}", web.StartPage)
	r.HandleFunc("/login", web.LoginPost).Methods("POST")
	r.HandleFunc("/login", web.Login)
	r.HandleFunc("/logout", web.Logout)
	r.HandleFunc("/signup", web.SignupPost).Methods("POST")
	r.HandleFunc("/signup", web.Signup)
	r.HandleFunc("/profiles/{nickname}", web.ProfilePage)
	r.HandleFunc("/add_article", web.AddArticlePost).Methods("POST")
	r.HandleFunc("/add_article", web.AddArticle)
	r.HandleFunc("/article/{article_id}", web.ArticlePagePost).Methods("POST")
	r.HandleFunc("/article/{article_id}", web.ArticlePage)
	r.HandleFunc("/settings", web.SettingsPostChText).Methods("POST")
	r.HandleFunc("/settings", web.Settings)
	r.HandleFunc("/settings_ph", web.SettingsPostChPhoto).Methods("POST")
	r.HandleFunc("/delete_article/{article_id}", web.DeleteArticlePost).Methods("POST")
	r.HandleFunc("/delete_comment/{comment_id}", web.DeleteCommentPost).Methods("POST")
	r.HandleFunc("/delete_profile/{user_id}", web.DeleteProfilePost).Methods("POST")
	r.HandleFunc("/edit_article/{article_id}", web.EditArticlePost).Methods("POST")
	r.HandleFunc("/edit_article/{article_id}", web.EditArticle)

	h := http.HandlerFunc(web.Handler404Page)
	r.NotFoundHandler = h
	http.Handle("/", r)
	log.Print("Listening on :" + os.Getenv("PORT"))
	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
