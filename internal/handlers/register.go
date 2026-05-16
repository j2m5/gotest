package handlers

import (
	"gotest/internal/db"
	"gotest/internal/templates"
	"net/http"
	"os"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		files := []string{
			"templates/layout.html",
			"templates/register.html",
		}

		data := map[string]any{
			"Title": "Register",
		}

		templates.Render(w, files, data)

		return
	}

	if r.Method == http.MethodPost {
		email := r.PostFormValue("email")
		login := r.FormValue("login")
		password := r.FormValue("password")

		user, err := db.CreateUser(db.Pool, email, login, password)

		if err != nil {
			return
		}

		token := generateToken()

		err = db.CreateEmailVerification(db.Pool, user.ID, token)

		err = makeEmailMessage(token)

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")

	emailVerification, err := db.FindEmailVerificationByToken(db.Pool, token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	err = db.UpdateEmailVerifiedAt(db.Pool, emailVerification.UserID)
	err = db.DeleteEmailVerification(db.Pool, emailVerification.Token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}

func makeEmailMessage(token string) error {
	file, err := os.Create("./Email-Verification-Code.txt")

	if err != nil {
		return err
	}

	defer file.Close()

	message := `
		Добро пожаловать!

		Осталось только подтвердить адрес электронной почты
		
		Чтобы закончить регистрацию, перейдите по ссылке: http://localhost:8081/verify?token=` + token

	_, err = file.WriteString(message)

	if err != nil {
		return err
	}

	return err
}
