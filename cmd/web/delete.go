package web

import (
	"net/http"
	"strconv"

	"goproj1/pkg/cookies"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func DeleteArticlePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := Db.Exec(`DELETE FROM public.article_likes WHERE article_id = $1`, vars["article_id"])
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.comments WHERE article_id = $1`, vars["article_id"])
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.articles WHERE id = $1`, vars["article_id"])
	CheckError(err)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func DeleteCommentPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var articleId int
	rows := Db.QueryRow(`SELECT "article_id" FROM "comments" WHERE "id" = $1`, vars["comment_id"])
	err := rows.Scan(&articleId)
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.comments WHERE id = $1`, vars["comment_id"])
	CheckError(err)

	redirecting(w, r, strconv.Itoa(articleId))
}

func DeleteProfilePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := Db.Exec(`DELETE FROM public.article_likes WHERE user_id = $1`, vars["user_id"])
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.comments WHERE author_id = $1`, vars["user_id"])
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.articles WHERE author_id = $1`, vars["user_id"])
	CheckError(err)
	_, err = Db.Exec(`DELETE FROM public.users WHERE id = $1`, vars["user_id"])
	CheckError(err)
	session, err := cookies.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	CheckServerError(w, err)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
