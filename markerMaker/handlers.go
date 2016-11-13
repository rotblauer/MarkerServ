package markerMaker

import (
	"encoding/json"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

var FuncMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var templates = template.Must(template.ParseFiles("templates/searchBarNav.html", "templates/index.html"))

func Index(w http.ResponseWriter, r *http.Request) {
	templates.Funcs(FuncMap)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecuteTemplate(w, "index", navBars)
}

//populator, temp for adding data
func populate(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var marker Marker

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&marker)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	key := markerKey(c, marker.MarkerName)
	if _, err := datastore.Put(c, key, &marker); err != nil { //store it
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
