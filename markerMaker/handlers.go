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

var templates = template.Must(template.ParseGlob("templates/*.html"))

type Data struct {
	NavBars []NavBar
	Markers []Marker
	Status  string
	Errors  []string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.Funcs(FuncMap)
	data := Data{NavBars: navs, Status: "None"}
	templates.ExecuteTemplate(w, "base", &data)
}

func markerHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	markers := queryByNames(parseForm(r, "list"), c)
	data := Data{NavBars: navs, Status: "MarkerTable", Markers: markers}
	templates.ExecuteTemplate(w, "base", &data)
}

// a	g	a
// chr1:1-247,249,719
func ucscHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	regions := parseForm(r, "list")
	var markers []Marker
	var errors []string
	for _, region := range regions {
		bed, err := parseBed3(strings.TrimSpace(region))
		if err != nil {
			c.Infof(region + " gave error: " + err.Error())
			errors = append(errors, region+" gave error: "+err.Error())
		} else {
			markerRegion := queryByPosition(bed.Chrom, bed.Start(), bed.End(), c)
			for _, marker := range markerRegion {
				markers = append(markers, marker)
			}
		}
	}
	data := Data{NavBars: navs, Status: "MarkerTable", Markers: markers, Errors: errors}
	templates.ExecuteTemplate(w, "base", &data)
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
