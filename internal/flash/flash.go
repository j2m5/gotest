package flash

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type Flash struct {
	store *sessions.CookieStore
}

func NewFlash(secret string) *Flash {
	return &Flash{store: sessions.NewCookieStore([]byte(secret))}
}

func (f *Flash) Set(w http.ResponseWriter, r *http.Request, kind string, message string) {
	session, _ := f.store.Get(r, "flash")
	session.AddFlash(message, kind)
	session.Save(r, w)
}

func (f *Flash) Get(w http.ResponseWriter, r *http.Request, kind string) []string {
	session, _ := f.store.Get(r, "flash")
	flashes := session.Flashes(kind)
	session.Save(r, w)

	messages := make([]string, len(flashes))
	for i, message := range flashes {
		messages[i] = message.(string)
	}

	return messages
}
