package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)

		return
	}

	if errs := validateRegister(request.Email, request.Login, request.Password); len(errs) > 0 {
		jsonResponse(w, map[string]any{"errors": errs}, http.StatusUnprocessableEntity)

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	if err != nil {
		h.serverError(w, r, err)

		return
	}

	user, err := h.storage.CreateUser(request.Email, request.Login, string(hash))

	if err != nil {
		h.serverError(w, r, err)

		return
	}

	token := generateToken()

	if err = h.storage.CreateEmailVerification(user.ID, token); err != nil {
		h.serverError(w, r, err)

		return
	}

	if err = makeEmailMessage(token); err != nil {
		h.serverError(w, r, err)

		return
	}

	jsonResponse(w, map[string]string{
		"message": "Успешная регистрация, проверьте указанный email для верификации аккаунта",
	}, http.StatusCreated)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	token := r.URL.Query().Get("token")

	if token == "" {
		jsonError(w, "Токен не найден", http.StatusBadRequest)

		return
	}

	emailVerification, err := h.storage.FindEmailVerificationByToken(token)

	if err != nil {
		jsonError(w, "Токен недействителен", http.StatusBadRequest)

		return
	}

	if err = h.storage.UpdateEmailVerifiedAt(emailVerification.UserID); err != nil {
		h.serverError(w, r, err)

		return
	}

	if err = h.storage.DeleteEmailVerification(emailVerification.Token); err != nil {
		h.serverError(w, r, err)

		return
	}

	jsonResponse(w, map[string]string{
		"message": "Аккаунт успешно верифицирован",
	}, http.StatusOK)
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
