package markerMaker

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/datastore"
)

var FuncMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var templates = template.Must(template.ParseFiles("templates/base.html", "templates/index.html", "templates/searchTypes.html", "templates/dataTable.html"))

type Data struct {
	NavBars []NavBar
	Markers []Marker
	Table   string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.Funcs(FuncMap)
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := Data{}
	data.NavBars = navs
	data.Table = "No"
	templates.ExecuteTemplate(w, "base", data)
}

func markerHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	markers := queryByNames(strings.Split(r.FormValue("list"), "\n"), c)
	data := Data{}
	data.NavBars = navs
	data.Markers = markers
	data.Table = "Yes"

	templates.ExecuteTemplate(w, "base", data)

}

// for a json like response - curl friendly
func markerHandlerRaw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	vars := mux.Vars(r)["ids"]
	c := appengine.NewContext(r)

	markers := queryByNames(strings.Split(vars, ","), c)

	c.Infof("HFD")
	markerJSON, err := json.Marshal(markers) // return bytes, err
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(markerJSON)
	}
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
