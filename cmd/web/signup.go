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

func Signup(w http.ResponseWriter, r *http.Request) {
	data := ViewData{}
	data.Auth = "false"
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	if session.Values["is_auth"] == true {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		files := []string{
			path.Join("templates", "signup.tmpl"),
			path.Join("templates", "base_not_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	}
}

func SignupPost(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	ctx := context.Background()
	//text-fields
	data := ViewData{}
	data.User.Name = r.FormValue("name")
	data.User.Surname = r.FormValue("surname")
	data.User.Nick = r.FormValue("nick")
	data.User.Email = r.FormValue("email")
	data.User.IsAdmin = false
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	//Sign up process
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
		PostError(w, r, "signup.tmpl", "1", "base_not_auth.tmpl")
	} else if boolNick == true {
		PostError(w, r, "signup.tmpl", "2", "base_not_auth.tmpl")
	} else if boolEmail == true {
		PostError(w, r, "signup.tmpl", "3", "base_not_auth.tmpl")
	} else {
		passwordBytes := []byte(password)
		hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
		CheckError(err)

		//file-field
		r.ParseMultipartForm(10 << 20)
		file, _, err := r.FormFile("myFile")
		if err != nil {
			data.User.PhotoPath = "static/default_avatar_zlwddd"
			_, err = Db.Exec(`INSERT INTO "users" ("nick", "name", "surname", "email", "password", "photo", "is_admin") VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				data.User.Nick, data.User.Name, data.User.Surname, data.User.Email, string(hashedPassword), data.User.PhotoPath, data.User.IsAdmin)
			CheckError(err)
		} else {
			defer file.Close()
			tempFile, err := ioutil.TempFile("static/", "upload-*.png")
			CheckError(err)
			defer tempFile.Close()
			fileBytes, err := ioutil.ReadAll(file)
			CheckError(err)
			tempFile.Write(fileBytes)
			CheckError(err)
			fileName := strings.TrimSuffix(tempFile.Name(), filepath.Ext(tempFile.Name()))
			_, err = cld.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{PublicID: fileName})
			CheckError(err)
			data.User.PhotoPath = fileName
			_, err = Db.Exec(`INSERT INTO "users" ("nick", "name", "surname", "email", "password", "photo", "is_admin") VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				data.User.Nick, data.User.Name, data.User.Surname, data.User.Email, string(hashedPassword), data.User.PhotoPath, data.User.IsAdmin)
			CheckError(err)
		}
		rows := Db.QueryRow(`SELECT id FROM public.users WHERE email = $1`, data.User.Email)
		CheckError(err)
		err = rows.Scan(&data.User.Id)
		CheckError(err)
		myImage, err := cld.Image(data.User.PhotoPath)
		CheckError(err)
		url, err := myImage.String()
		CheckError(err)
		data.User.PhotoPath = url
		session.Values["data"] = data
		session.Values["is_auth"] = true
		err = session.Save(r, w)
		CheckServerError(w, err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
