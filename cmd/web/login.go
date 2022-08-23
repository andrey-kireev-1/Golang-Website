package web

import (
	"html/template"
	"net/http"
	"path"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	data := ViewData{}
	data.Auth = "false"
	data.ErrorType = ""
	files := []string{
		path.Join("templates", "login.tmpl"),
		path.Join("templates", "base_not_auth.tmpl"),
	}
	tmpl, err := template.ParseFiles(files...)
	CheckServerError(w, err)
	err = tmpl.Execute(w, data)
	CheckServerError(w, err)
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	data := ViewData{}
	email := r.FormValue("email")
	password := r.FormValue("password")

	session, err := cookies.Store.Get(r, "session-name")
	CheckServerError(w, err)

	var emailFromDb, passwordFromDb string
	rows := Db.QueryRow(`SELECT "email", "password" FROM "users" WHERE "email" = $1`, email)
	CheckError(err)
	err = rows.Scan(&emailFromDb, &passwordFromDb)
	if err != nil {
		data.Auth = "false"
		data.ErrorType = "1"
		files := []string{
			path.Join("templates", "login.tmpl"),
			path.Join("templates", "base_not_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	} else {

		if email == emailFromDb && bcrypt.CompareHashAndPassword([]byte(passwordFromDb), []byte(password)) == nil {
			session.Values["is_auth"] = true
			row := Db.QueryRow(`SELECT "id", "nick", "name", "surname", "email", "photo", "is_admin" FROM "users" WHERE "email" = $1`, email)
			err = row.Scan(&data.User.Id, &data.User.Nick, &data.User.Name, &data.User.Surname, &data.User.Email, &data.User.PhotoPath, &data.User.IsAdmin)
			CheckError(err)
			myImage, err := cld.Image(data.User.PhotoPath)
			CheckError(err)
			url, err := myImage.String()
			CheckError(err)
			data.User.PhotoPath = url
			session.Values["data"] = data
			err = session.Save(r, w)
			CheckServerError(w, err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			data.Auth = "false"
			data.ErrorType = "1"
			files := []string{
				path.Join("templates", "login.tmpl"),
				path.Join("templates", "base_not_auth.tmpl"),
			}
			tmpl, err := template.ParseFiles(files...)
			CheckServerError(w, err)
			err = tmpl.Execute(w, data)
			CheckServerError(w, err)
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := cookies.Store.Get(r, "session-name")
	CheckServerError(w, err)
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	CheckServerError(w, err)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
