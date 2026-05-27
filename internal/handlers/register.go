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
			"Title":   "Register",
			"Success": h.flash.GetOne(w, r, "success"),
			"Errors":  h.flash.Get(w, r, "error"),
			"Old": map[string]string{
				"Email": h.flash.GetOne(w, r, "old_email"),
				"Login": h.flash.GetOne(w, r, "old_login"),
			},
		}

		templates.Render(w, files, data)

		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		login := r.FormValue("login")
		password := r.FormValue("password")

		if errs := validateRegister(email, login, password); len(errs) > 0 {
			for _, e := range errs {
				h.flash.Set(w, r, "error", e)
			}
			h.flash.Set(w, r, "old_email", email)
			h.flash.Set(w, r, "old_login", login)
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

		h.flash.Set(w, r, "success", "Успешная регистрация, проверьте указанный email для верификации аккаунта")
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

	h.flash.Set(w, r, "success", "Аккаунт успешно верифицирован")
	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}

func validateRegister(email string, login string, password string) []string {
	var errors []string

	if email == "" {
		errors = append(errors, "Поле Email обязательно для заполнения")
	}

	if login == "" {
		errors = append(errors, "Поле Логин обязательно для заполнения")
	}

	if password == "" {
		errors = append(errors, "Поле Пароль обязательно для заполнения")
	}

	return errors
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
