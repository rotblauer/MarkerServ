package markerMaker

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"appengine"
)

var funcMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

type Data struct {
	NavBars    []NavBar
	Markers    []Marker
	Status     string
	Errors     []error
	SearchType string
}

//Welcome
func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.Funcs(funcMap)
	data := Data{NavBars: navs, Status: "None"}
	templates.ExecuteTemplate(w, "base", &data)
}

//MarkerName
func markerHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	markers, errors := queryAll(parseForm(r, "list"), MARKER_NAME, c)
	printResults(markers, errors, "Results for Probeset ID search:", w)
}

//RS ID
func rsIdHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	markers, errors := queryAll(parseForm(r, "list"), RS_ID, c)
	printResults(markers, errors, "Results for rs ID search:", w)

}

//UCSC
func ucscHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	markers, errors := queryAll(parseForm(r, "list"), UCSC_REGION, c)
	printResults(markers, errors, "Results for UCSC Region search:", w)
}

//Render the results
func printResults(markers []Marker, errors []error, searchType string, w http.ResponseWriter) {
	data := Data{NavBars: navs, Status: "MarkerTable", Markers: markers, Errors: errors, SearchType: searchType}
	templates.ExecuteTemplate(w, "base", &data)
}

// for a json like response - curl friendly
func markerHandlerRaw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	vars := mux.Vars(r)["ids"]
	c := appengine.NewContext(r)
	markers, _ := queryAll(strings.Split(vars, ","), MARKER_NAME, c)
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
	errS := storeMarker(marker, c)
	if errS != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
