package web

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	_ "github.com/lib/pq"
)

func AddArticle(w http.ResponseWriter, r *http.Request) {
	data := ViewData{}
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	if session.Values["is_auth"] == false {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		data = session.Values["data"].(ViewData)
		data.Auth = "true"
		files := []string{
			path.Join("templates", "add.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	}
}

func AddArticlePost(w http.ResponseWriter, r *http.Request) {
	///text-fields
	title := r.FormValue("title")
	text := r.FormValue("text")

	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	if title == "" {
		data := ViewData{}
		data.Auth = "true"
		data.ErrorType = "1"
		files := []string{
			path.Join("templates", "add.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
			path.Join("templates", "base_not_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	} else {

		//file-field
		r.ParseMultipartForm(10 << 20)
		file, _, err := r.FormFile("pic-file")
		if err != nil {
			fileName := ""
			_, err = Db.Exec(`INSERT INTO "articles" ("title", "text", "picture", "author_id", "date") VALUES ($1, $2, $3, $4, $5)`,
				title, text, fileName, session.Values["data"].(ViewData).User.Id, time.Now())
			CheckError(err)
		} else {
			defer file.Close()
			tempFile, err := ioutil.TempFile("static/", "article-*.png")
			CheckError(err)
			defer tempFile.Close()
			fileBytes, err := ioutil.ReadAll(file)
			CheckError(err)
			tempFile.Write(fileBytes)
			CheckError(err)
			cld, err := cloudinary.NewFromURL(config.AccessURL)
			CheckError(err)
			ctx := context.Background()
			fileName := strings.TrimSuffix(tempFile.Name(), filepath.Ext(tempFile.Name()))
			_, err = cld.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{PublicID: fileName})
			CheckError(err)
			_, err = Db.Exec(`INSERT INTO "articles" ("title", "text", "picture", "author_id", "date") VALUES ($1, $2, $3, $4, $5)`,
				title, text, fileName, session.Values["data"].(ViewData).User.Id, time.Now())
			CheckError(err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}
