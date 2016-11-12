package markerMaker

import (
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(FormTemplate)).Execute(w, nil)
	template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, nil)
}
