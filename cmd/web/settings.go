package web

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func Settings(w http.ResponseWriter, r *http.Request) {
	data := ViewData{}
	data.Auth = "true"
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	if session.Values["is_auth"] != true {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		data = session.Values["data"].(ViewData)
		files := []string{
			path.Join("templates", "settings.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	}

}

func SettingsPostChText(w http.ResponseWriter, r *http.Request) {
	//text-fields
	data := ViewData{}
	data.User.Name = r.FormValue("name")
	data.User.Surname = r.FormValue("surname")
	data.User.Nick = r.FormValue("nick")
	data.User.Email = r.FormValue("email")
	data.User.IsAdmin = false
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	//Change process
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	rowEmail := Db.QueryRow(`SELECT EXISTS (SELECT * FROM "users" WHERE email = $1)`, data.User.Email)
	rowNick := Db.QueryRow(`SELECT EXISTS (SELECT * FROM "users" WHERE nick = $1)`, data.User.Nick)
	var boolEmail, boolNick bool
	err = rowEmail.Scan(&boolEmail)
	CheckError(err)
	err = rowNick.Scan(&boolNick)
	CheckError(err)
	if password != password2 {
		PostError(w, r, "settings.tmpl", "1", "base_auth.tmpl")
	} else if boolNick == true && data.User.Nick != session.Values["data"].(ViewData).User.Nick {
		PostError(w, r, "settings.tmpl", "2", "base_auth.tmpl")
	} else if boolEmail == true && data.User.Email != session.Values["data"].(ViewData).User.Email {
		PostError(w, r, "settings.tmpl", "3", "base_auth.tmpl")
	} else {
		passwordBytes := []byte(password)
		hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
		CheckError(err)

		_, err = Db.Exec(`UPDATE public.users SET nick=$1, name=$2, surname=$3, email=$4, password=$5 WHERE nick = $6`,
			data.User.Nick, data.User.Name, data.User.Surname, data.User.Email, string(hashedPassword), session.Values["data"].(ViewData).User.Nick)
		CheckError(err)
		rows := Db.QueryRow(`SELECT "id", "photo", "is_admin" FROM "users" WHERE "email" = $1`, data.User.Email)
		CheckError(err)
		err = rows.Scan(&data.User.Id, &data.User.PhotoPath, &data.User.IsAdmin)
		CheckError(err)
		session.Values["data"] = data
		session.Values["is_auth"] = true
		err = session.Save(r, w)
		CheckServerError(w, err)
	}
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func SettingsPostChPhoto(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	ctx := context.Background()
	data := ViewData{}
	//file-field
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("pic-file")
	if err != nil {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	} else {
		defer file.Close()
		tempFile, err := ioutil.TempFile("static/", "upload-*.png")

		CheckError(err)
		defer tempFile.Close()
		fileBytes, err := ioutil.ReadAll(file)
		CheckError(err)
		tempFile.Write(fileBytes)
		CheckError(err)

		session, err := cookies.Store.Get(r, "session-name")
		CheckError(err)

		data.User.PhotoPath = strings.TrimSuffix(tempFile.Name(), filepath.Ext(tempFile.Name()))
		_, err = cld.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{PublicID: data.User.PhotoPath})
		CheckError(err)
		_, err = Db.Exec(`UPDATE public.users SET photo=$1 WHERE nick = $2`, data.User.PhotoPath, session.Values["data"].(ViewData).User.Nick)
		CheckError(err)
		rows := Db.QueryRow(`SELECT "id", "nick", "name", "surname", "email", "photo", "is_admin" FROM "users" WHERE "nick" = $1`, session.Values["data"].(ViewData).User.Nick)
		CheckError(err)
		err = rows.Scan(&data.User.Id, &data.User.Nick, &data.User.Name, &data.User.Surname, &data.User.Email, &data.User.PhotoPath, &data.User.IsAdmin)
		CheckError(err)
		my_image, err := cld.Image(data.User.PhotoPath)
		CheckError(err)
		url, err := my_image.String()
		CheckError(err)
		data.User.PhotoPath = url
		session.Values["data"] = data
		session.Values["is_auth"] = true
		err = session.Save(r, w)
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	}
}
