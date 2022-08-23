package web

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"goproj1/pkg/config"
	"goproj1/pkg/cookies"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func Search(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	v := r.URL.Query()
	search := v.Get("search")
	data := ViewData{}
	query1 := `SELECT id, title, text, picture, author_id, date,
	(SELECT COUNT(*) FROM public.article_likes 
		WHERE public.article_likes.article_id = public.articles.id) AS cntlike,
	(SELECT EXISTS (SELECT * FROM public.article_likes
		WHERE public.article_likes.user_id = $1 AND public.article_likes.article_id = public.articles.id)) AS liked
	FROM public.articles WHERE "title" LIKE '%` + search + `%' ORDER BY date DESC`
	query2 := `SELECT id, title, text, picture, author_id, date,
	(SELECT COUNT(*) FROM public.article_likes 
		WHERE public.article_likes.article_id = public.articles.id) AS cntlike
	FROM public.articles WHERE "title" LIKE '%` + search + `%' ORDER BY date DESC`
	articles := []Article{}

	session, _ := cookies.Store.Get(r, "session-name")
	var files []string
	if session.Values["is_auth"] == true {
		rows, err := Db.Query(query1, session.Values["data"].(ViewData).User.Id)
		CheckError(err)
		defer rows.Close()

		for rows.Next() {
			p := Article{}
			var tmpTime time.Time
			err := rows.Scan(&p.Id, &p.Title, &p.Text, &p.Picture, &p.AuthorId, &tmpTime, &p.CntLikes, &p.Liked)
			myImage, err := cld.Image(p.Picture)
			CheckError(err)
			url, err := myImage.String()
			CheckError(err)
			p.Picture = url
			p.Date = tmpTime.Format("15:04 01/02/2006")
			CheckError(err)
			articles = append(articles, p)
		}
		data = session.Values["data"].(ViewData)
		data.Articles = articles

		data.Auth = "true"
	} else {
		rows, err := Db.Query(query2)
		CheckError(err)
		defer rows.Close()

		for rows.Next() {
			p := Article{}
			var tmpTime time.Time
			err := rows.Scan(&p.Id, &p.Title, &p.Text, &p.Picture, &p.AuthorId, &tmpTime, &p.CntLikes)
			myImage, err := cld.Image(p.Picture)
			CheckError(err)
			url, err := myImage.String()
			CheckError(err)
			p.Picture = url
			p.Liked = false
			p.Date = tmpTime.Format("15:04 01/02/2006")
			CheckError(err)
			articles = append(articles, p)
		}

		data.Articles = articles

		data.Auth = "false"
	}
	files = []string{
		path.Join("templates", "index.tmpl"),
		path.Join("templates", "base_auth.tmpl"),
		path.Join("templates", "base_not_auth.tmpl"),
	}
	tmpl, err := template.ParseFiles(files...)
	CheckServerError(w, err)
	err = tmpl.Execute(w, data)
	CheckServerError(w, err)
}

func Like(w http.ResponseWriter, r *http.Request) {
	var str string
	var id int
	var idCntl []string
	ajaxPostData := r.FormValue("ajax_post_data")
	fmt.Println("Receive ajax post data string", ajaxPostData)
	session, _ := cookies.Store.Get(r, "session-name")
	if session.Values["is_auth"] != true {
		return
	} else {
		rows, err := Db.Query(`SELECT "article_id" from "article_likes" WHERE "user_id" = $1`, session.Values["data"].(ViewData).User.Id)
		CheckError(err)
		defer rows.Close()
		var check bool = false
		var articleId int
		idCntl = strings.Split(ajaxPostData, "_")
		id, _ = strconv.Atoi(idCntl[0])
		for rows.Next() {
			err := rows.Scan(&articleId)
			CheckError(err)
			if articleId == id {
				check = true
				break
			}
		}
		if check == true {
			res, _ := strconv.Atoi(idCntl[1])
			res--
			resStr := strconv.Itoa(res)
			_, err = Db.Exec(`DELETE FROM "article_likes" WHERE "article_id" = $1 AND "user_id" = $2`, id, session.Values["data"].(ViewData).User.Id)
			CheckError(err)
			str = `<button type="button" class="btn like mt-2" data-id="` + idCntl[0] + "_" + resStr + `"><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" id="Capa_1" x="0px" y="0px" width="20px" height="20px" viewBox="0 0 45.743 45.743" style="fill:#C3073F;" xml:space="preserve">
	<g>
	  <path d="M34.199,3.83c-3.944,0-7.428,1.98-9.51,4.997c0,0-0.703,1.052-1.818,1.052c-1.114,0-1.817-1.052-1.817-1.052   c-2.083-3.017-5.565-4.997-9.51-4.997C5.168,3.83,0,8.998,0,15.376c0,1.506,0.296,2.939,0.82,4.258   c3.234,10.042,17.698,21.848,22.051,22.279c4.354-0.431,18.816-12.237,22.052-22.279c0.524-1.318,0.82-2.752,0.82-4.258   C45.743,8.998,40.575,3.83,34.199,3.83z"/>
	</g>
</svg> ` + resStr + `</button>`
			w.Write([]byte(str))
		} else {
			res, _ := strconv.Atoi(idCntl[1])
			res++
			resStr := strconv.Itoa(res)
			_, err = Db.Exec(`INSERT INTO public.article_likes (article_id, user_id) VALUES ($1, $2)`, id, session.Values["data"].(ViewData).User.Id)
			CheckError(err)
			str = `<button type="button" class="btn like-pressed mt-2" data-id="` + idCntl[0] + "_" + resStr + `"><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" id="Capa_1" x="0px" y="0px" width="20px" height="20px" viewBox="0 0 45.743 45.743" style="fill:#28282b;" xml:space="preserve">
	<g>
	  <path d="M34.199,3.83c-3.944,0-7.428,1.98-9.51,4.997c0,0-0.703,1.052-1.818,1.052c-1.114,0-1.817-1.052-1.817-1.052   c-2.083-3.017-5.565-4.997-9.51-4.997C5.168,3.83,0,8.998,0,15.376c0,1.506,0.296,2.939,0.82,4.258   c3.234,10.042,17.698,21.848,22.051,22.279c4.354-0.431,18.816-12.237,22.052-22.279c0.524-1.318,0.82-2.752,0.82-4.258   C45.743,8.998,40.575,3.83,34.199,3.83z"/>
	</g>
</svg> ` + resStr + `</button>`
			w.Write([]byte(str))
		}
	}
}

func StartPage(w http.ResponseWriter, r *http.Request) {
	cld, err := cloudinary.NewFromURL(config.AccessURL)
	CheckError(err)
	var exist bool
	vars := mux.Vars(r)
	_, ok := vars["page_id"]
	data := ViewData{}
	if ok == false {
		vars["page_id"] = "0"
	}

	pageIdInt, err := strconv.Atoi(vars["page_id"])
	pageIdInt5 := pageIdInt * 5
	pageIdStr := strconv.Itoa(pageIdInt5)
	CheckError(err)
	articles := []Article{}

	session, _ := cookies.Store.Get(r, "session-name")
	var files []string
	if session.Values["is_auth"] == true {
		rows, err := Db.Query(`SELECT id, title, text, picture, author_id, date,
		(SELECT COUNT(*) FROM public.article_likes 
			WHERE public.article_likes.article_id = public.articles.id) AS cntlike,
		(SELECT EXISTS (SELECT * FROM public.article_likes
			WHERE public.article_likes.user_id = $1 AND public.article_likes.article_id = public.articles.id)) AS liked
		FROM public.articles ORDER BY date DESC LIMIT 5 OFFSET `+pageIdStr, session.Values["data"].(ViewData).User.Id)
		CheckError(err)
		defer rows.Close()

		for rows.Next() {
			p := Article{}
			var tmpTime time.Time
			err := rows.Scan(&p.Id, &p.Title, &p.Text, &p.Picture, &p.AuthorId, &tmpTime, &p.CntLikes, &p.Liked)
			myImage, err := cld.Image(p.Picture)
			CheckError(err)
			url, err := myImage.String()
			CheckError(err)
			p.Picture = url
			p.Date = tmpTime.Format("15:04 01/02/2006")
			CheckError(err)
			articles = append(articles, p)
		}
		data = session.Values["data"].(ViewData)
		pageIdInt55 := pageIdInt5 + 5
		pageIdStr = strconv.Itoa(pageIdInt55)
		row := Db.QueryRow(`SELECT EXISTS (SELECT * FROM public.articles ORDER BY date DESC LIMIT 5 OFFSET ` + pageIdStr + `)`)
		err = row.Scan(&exist)
		CheckError(err)
		data.PageId = pageIdInt
		if exist == true {
			data.PageIdNext = pageIdInt + 1
		} else {
			data.PageIdNext = -1
		}
		data.PageIdBack = pageIdInt - 1

		data.Articles = articles
		data.Auth = "true"
	} else {
		rows, err := Db.Query(`SELECT id, title, text, picture, author_id, date,
		(SELECT COUNT(*) FROM public.article_likes 
			WHERE public.article_likes.article_id = public.articles.id) AS cntlike
		FROM public.articles ORDER BY date DESC LIMIT 5 OFFSET ` + pageIdStr)
		CheckError(err)
		defer rows.Close()

		for rows.Next() {
			p := Article{}
			var tmpTime time.Time
			err := rows.Scan(&p.Id, &p.Title, &p.Text, &p.Picture, &p.AuthorId, &tmpTime, &p.CntLikes)
			myImage, err := cld.Image(p.Picture)
			CheckError(err)
			url, err := myImage.String()
			CheckError(err)
			p.Picture = url
			p.Liked = false
			p.Date = tmpTime.Format("15:04 01/02/2006")
			CheckError(err)
			articles = append(articles, p)
		}
		pageIdInt55 := pageIdInt5 + 5
		pageIdStr = strconv.Itoa(pageIdInt55)
		row := Db.QueryRow(`SELECT EXISTS (SELECT * FROM public.articles ORDER BY date DESC LIMIT 5 OFFSET ` + pageIdStr + `)`)
		err = row.Scan(&exist)
		CheckError(err)
		data.PageId = pageIdInt
		if exist == true {
			data.PageIdNext = pageIdInt + 1
		} else {
			data.PageIdNext = -1
		}
		data.PageIdBack = pageIdInt - 1
		data.Articles = articles
		data.Auth = "false"
	}
	files = []string{
		path.Join("templates", "index.tmpl"),
		path.Join("templates", "base_auth.tmpl"),
		path.Join("templates", "base_not_auth.tmpl"),
	}
	tmpl, err := template.ParseFiles(files...)
	CheckServerError(w, err)
	err = tmpl.Execute(w, data)
	CheckServerError(w, err)
}
