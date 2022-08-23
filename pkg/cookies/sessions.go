package cookies

import (
	"os"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func SetSessionSettings() {
	Store.Options.MaxAge = 0
	Store.Options.HttpOnly = true
	Store.Options.Secure = true
}
