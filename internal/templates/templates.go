package templates

import (
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, files []string, data any) {
	tmpl := template.Must(template.ParseFiles(files...))

	err := tmpl.ExecuteTemplate(w, "layout.html", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
