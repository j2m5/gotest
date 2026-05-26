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
		email := r.FormValue("email")
		login := r.FormValue("login")
		password := r.FormValue("password")

		if email == "" {
			http.Error(w, "Поле Email обязательно для заполнения", http.StatusUnprocessableEntity)

			return
		}

		if login == "" {
			http.Error(w, "Поле Логин обязательно для заполнения", http.StatusUnprocessableEntity)

			return
		}

		if password == "" {
			http.Error(w, "Поле Пароль обязательно для заполнения", http.StatusUnprocessableEntity)

			return
		}

		user, err := db.CreateUser(db.Pool, email, login, password)

		if err != nil {
			http.Error(w, "Cannot create user", http.StatusInternalServerError)

			return
		}

		token := generateToken()

		err = db.CreateEmailVerification(db.Pool, user.ID, token)

		if err != nil {
			http.Error(w, "Cannot create email verification", http.StatusInternalServerError)

			return
		}

		err = makeEmailMessage(token)

		if err != nil {
			http.Error(w, "Cannot create email message", http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(w, "Missing token", http.StatusUnprocessableEntity)

		return
	}

	emailVerification, err := db.FindEmailVerificationByToken(db.Pool, token)

	if err != nil {
		http.Error(w, "Invalid verification token", http.StatusInternalServerError)

		return
	}

	err = db.UpdateEmailVerifiedAt(db.Pool, emailVerification.UserID)

	if err != nil {
		http.Error(w, "Cannot verify email", http.StatusInternalServerError)

		return
	}

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
