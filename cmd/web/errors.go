package web

import (
	"html/template"
	"net/http"
	"path"

	"goproj1/pkg/cookies"

	_ "github.com/lib/pq"
)

func CheckServerError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Handler404Page(w http.ResponseWriter, r *http.Request) {
	session, _ := cookies.Store.Get(r, "session-name")
	data := ViewData{}
	var files []string
	if session.Values["is_auth"] == true {
		data.Auth = "true"
		data = session.Values["data"].(ViewData)
	} else {
		data.Auth = "false"
	}
	files = []string{
		path.Join("templates", "404.tmpl"),
		path.Join("templates", "base_auth.tmpl"),
		path.Join("templates", "base_not_auth.tmpl"),
	}
	tmpl, err := template.ParseFiles(files...)
	CheckServerError(w, err)
	err = tmpl.Execute(w, data)
	CheckServerError(w, err)
}

func PostError(w http.ResponseWriter, r *http.Request, url_path string, error_type string, auth_path string) {
	data := ViewData{}

	session, _ := cookies.Store.Get(r, "session-name")
	if session.Values["is_auth"] == true {
		data = session.Values["data"].(ViewData)
		data.Auth = "true"
	} else {
		data.Auth = "false"
	}
	files := []string{
		path.Join("templates", url_path),
		path.Join("templates", auth_path),
	}
	data.ErrorType = error_type
	tmpl, err := template.ParseFiles(files...)
	CheckServerError(w, err)
	err = tmpl.Execute(w, data)
	CheckServerError(w, err)
}
