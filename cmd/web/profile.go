package web

import (
	"html/template"
	"net/http"
	"path"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	vars := mux.Vars(r)
	session, _ := cookies.Store.Get(r, "session-name")
	var files []string
	data := ViewData{}
	if session.Values["is_auth"] == true {
		data = session.Values["data"].(ViewData)
		data.Auth = "true"
	} else {
		data.Auth = "false"
	}
	row := Db.QueryRow(`SELECT "id", "nick", "name", "surname", "email", "photo", "is_admin" FROM "users" WHERE "nick" = $1`, vars["nickname"])
	err = row.Scan(&data.Profile.Id, &data.Profile.Nick, &data.Profile.Name, &data.Profile.Surname, &data.Profile.Email, &data.Profile.PhotoPath, &data.Profile.IsAdmin)
	myImage, err := cld.Image(data.Profile.PhotoPath)
	CheckError(err)
	url, err := myImage.String()
	CheckError(err)
	data.Profile.PhotoPath = url
	if err != nil {
		Handler404Page(w, r)
	} else {
		cntArtsRow := Db.QueryRow(`SELECT COUNT(*) FROM "articles" WHERE "author_id" = $1`, data.Profile.Id)
		cntComsRow := Db.QueryRow(`SELECT COUNT(*) FROM "comments" WHERE "author_id" = $1`, data.Profile.Id)
		err = cntArtsRow.Scan(&data.CntArticles)
		CheckError(err)
		err = cntComsRow.Scan(&data.CntComments)
		CheckError(err)
		files = []string{
			path.Join("templates", "profile.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
			path.Join("templates", "base_not_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	}
}
