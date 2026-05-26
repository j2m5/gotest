package flash

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret"))

func Set(w http.ResponseWriter, r *http.Request, kind string, message string) {
	session, _ := store.Get(r, "flash")
	session.AddFlash(message, kind)
	session.Save(r, w)
}

func Get(w http.ResponseWriter, r *http.Request, kind string) []string {
	session, _ := store.Get(r, "flash")
	flashes := session.Flashes(kind)
	session.Save(r, w)

	messages := make([]string, len(flashes))
	for i, message := range flashes {
		messages[i] = message.(string)
	}

	return messages
}
