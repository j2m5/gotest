package handlers

import (
	"gotest/internal/templates"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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
			h.flash.Set(w, r, "error", "Поле Email обязательно для заполнения")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

			return
		}

		if login == "" {
			h.flash.Set(w, r, "error", "Поле Логин обязательно для заполнения")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

			return
		}

		if password == "" {
			h.flash.Set(w, r, "error", "Поле Пароль обязательно для заполнения")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if err != nil {
			h.serverError(w, r, err)

			return
		}

		user, err := h.storage.CreateUser(email, login, string(hash))

		if err != nil {
			h.serverError(w, r, err)

			return
		}

		token := generateToken()

		err = h.storage.CreateEmailVerification(user.ID, token)

		if err != nil {
			h.serverError(w, r, err)

			return
		}

		err = makeEmailMessage(token)

		if err != nil {
			h.serverError(w, r, err)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.flash.Set(w, r, "error", "")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

		return
	}

	token := r.URL.Query().Get("token")

	if token == "" {
		h.flash.Set(w, r, "error", "Токен не найден")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)

		return
	}

	emailVerification, err := h.storage.FindEmailVerificationByToken(token)

	if err != nil {
		h.serverError(w, r, err)

		return
	}

	err = h.storage.UpdateEmailVerifiedAt(emailVerification.UserID)

	if err != nil {
		h.serverError(w, r, err)

		return
	}

	err = h.storage.DeleteEmailVerification(emailVerification.Token)

	if err != nil {
		h.serverError(w, r, err)

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
