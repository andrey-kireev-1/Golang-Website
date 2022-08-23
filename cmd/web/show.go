package web

import (
	"html/template"
	"net/http"
	"path"
	"time"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func ArticlePage(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	vars := mux.Vars(r)
	session, _ := cookies.Store.Get(r, "session-name")
	var files []string
	var tmpTime time.Time
	data := ViewData{}
	if session.Values["is_auth"] == true {
		data = session.Values["data"].(ViewData)
		data.Auth = "true"
	} else {
		data.Auth = "false"
	}

	rows, err := Db.Query(`SELECT id, text, article_id, author_id, date,
	(SELECT name FROM public.users 
        WHERE public.users.Id = public.comments.author_id) AS name,
	(SELECT surname FROM public.users 
        WHERE public.users.Id = public.comments.author_id) AS surname,
	(SELECT nick FROM public.users 
        WHERE public.users.Id = public.comments.author_id) AS nick,
	(SELECT photo FROM public.users 
        WHERE public.users.Id = public.comments.author_id) AS photo
	FROM public.comments WHERE article_id = $1`, vars["article_id"])
	CheckError(err)
	defer rows.Close()
	comments := []Comment{}

	for rows.Next() {
		p := Comment{}
		var tmpTime time.Time
		err := rows.Scan(&p.Id, &p.Text, &p.ArticleId, &p.AuthorId, &tmpTime, &p.AuthorName, &p.AuthorSurname, &p.AuthorNick, &p.AuthorPhotoPath)
		if session.Values["is_auth"] == true {
			if data.User.Id == p.AuthorId {
				p.IsAuthor = true
			}
		} else {
			p.IsAuthor = false
		}
		myImageComm, err := cld.Image(p.AuthorPhotoPath)
		CheckError(err)
		urlComm, err := myImageComm.String()
		CheckError(err)
		p.AuthorPhotoPath = urlComm
		p.Date = tmpTime.Format("15:04 01/02/2006")
		CheckError(err)
		comments = append(comments, p)
	}
	data.Comments = comments
	row := Db.QueryRow(`SELECT "id", "title", "text", "picture", "author_id", "date",
	(SELECT name FROM public.users 
        WHERE public.users.Id = public.articles.author_id) AS name,
	(SELECT surname FROM public.users 
        WHERE public.users.Id = public.articles.author_id) AS surname,
	(SELECT nick FROM public.users 
		WHERE public.users.Id = public.articles.author_id) AS nick
		 FROM "articles" WHERE "id" = $1`, vars["article_id"])
	err = row.Scan(&data.OneArticle.Id, &data.OneArticle.Title, &data.OneArticle.Text, &data.OneArticle.Picture, &data.OneArticle.AuthorId, &tmpTime, &data.Profile.Name, &data.Profile.Surname, &data.Profile.Nick)
	myImage, err := cld.Image(data.OneArticle.Picture)
	CheckError(err)
	url, err := myImage.String()
	CheckError(err)
	data.OneArticle.Picture = url
	data.OneArticle.Date = tmpTime.Format("15:04 01/02/2006")
	if session.Values["is_auth"] == true {
		if data.User.Id == data.OneArticle.AuthorId {
			data.OneArticle.IsAuthor = true
		}
	} else {
		data.OneArticle.IsAuthor = false
	}

	if err != nil {
		Handler404Page(w, r)
	} else {
		files = []string{
			path.Join("templates", "article.tmpl"),
			path.Join("templates", "base_auth.tmpl"),
			path.Join("templates", "base_not_auth.tmpl"),
		}
		tmpl, err := template.ParseFiles(files...)
		CheckServerError(w, err)
		err = tmpl.Execute(w, data)
		CheckServerError(w, err)
	}
}
