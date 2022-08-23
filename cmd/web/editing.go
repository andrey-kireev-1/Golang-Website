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
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func redirecting(w http.ResponseWriter, r *http.Request, str string) {
	rPath := "/article/" + str
	http.Redirect(w, r, rPath, http.StatusSeeOther)
}

func EditArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data := ViewData{}
	data.Auth = "true"
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	var id, authorId int
	var title, text string
	row := Db.QueryRow(`SELECT "id", "title", "text", "author_id" FROM "articles" WHERE "id" = $1`, vars["article_id"])
	err = row.Scan(&id, &title, &text, &authorId)
	CheckError(err)
	if session.Values["is_auth"] != true {
		redirecting(w, r, vars["article_id"])
	} else if session.Values["data"].(ViewData).User.Id == authorId {
		data = session.Values["data"].(ViewData)
		data.OneArticle.Id = id
		data.OneArticle.Title = title
		data.OneArticle.Text = text
		files := []string{
			path.Join("templates", "edit_article.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	} else {
		redirecting(w, r, vars["article_id"])
	}
}

func EditArticlePost(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	ctx := context.Background()
	vars := mux.Vars(r)
	///text-fields
	title := r.FormValue("title")
	text := r.FormValue("text")

	if title == "" {
		PostError(w, r, "edit_article.tmpl", "1", "base_auth.tmpl")
	} else {
		//file-field
		r.ParseMultipartForm(10 << 20)
		file, _, err := r.FormFile("pic-file")
		if err != nil {
			_, err = Db.Exec(`UPDATE public.articles SET title=$1, text=$2 WHERE id = $3`, title, text, vars["article_id"])
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
			file_name := strings.TrimSuffix(tempFile.Name(), filepath.Ext(tempFile.Name()))
			_, err = cld.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{PublicID: file_name})
			CheckError(err)
			_, err = Db.Exec(`UPDATE public.articles SET title=$1, text=$2, picture=$3 WHERE id = $4`, title, text, file_name, vars["article_id"])
			CheckError(err)
		}
		redirecting(w, r, vars["article_id"])
	}
}
