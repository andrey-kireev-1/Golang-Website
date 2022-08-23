package web

import (
	"net/http"
	"time"

	"goproj1/pkg/cookies"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func ArticlePagePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	///text-fields
	text := r.FormValue("text")
	session, err := cookies.Store.Get(r, "session-name")
	CheckError(err)
	if text == "" {
		redirecting(w, r, vars["article_id"])
	} else {
		_, err = Db.Exec(`INSERT INTO "comments" ("text", "article_id", "author_id", "date") VALUES ($1, $2, $3, $4)`,
			text, vars["article_id"], session.Values["data"].(ViewData).User.Id, time.Now())
		CheckError(err)
		redirecting(w, r, vars["article_id"])
	}
}
