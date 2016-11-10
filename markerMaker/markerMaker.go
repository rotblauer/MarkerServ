package markerMaker

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"appengine"
	"appengine/datastore"
)

// current information returned for query
// queries can be set up to handle pos<X&&chr==2, etc. This would allow getting all markers for an interval across any/all arrays
// could also do RS ID searches, etc
type Marker struct {
	MarkerName string   `json:"markerName"`
	RSId       string   `json:"rsID"`
	Chromosome string   `json:"chromosome"`
	Position   int      `json:"position"`
	A_Allele   string   `json:"a_allele"`
	B_Allele   string   `json:"b_allele"`
	Arrays     []string `json:"arrays"` //which arrays this marker is present on
}

//template html for displaying results
const (
	PrintTemplate = `
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/css/bootstrap.min.css" integrity="sha384-AysaV+vQoT3kOAXZkl02PThvDr8HYKPZhNT5h/CXfBThSRXQ6jW5DO2ekP5ViFdi" crossorigin="anonymous">
<link rel="stylesheet" href="https://cdn.datatables.net/1.10.12/css/dataTables.bootstrap.min.css">
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<meta http-equiv="x-ua-compatible" content="ie=edge">

<h1>Marker Query -sortable table</h1>
<table class="table table-striped table-bordered" id="markertable">
    <thead>
        <tr>
            <th> Marker Name</th>
            <th>Rs Id</th>
            <th>Chromosome</th>
            <th>Position</th>
            <th>A Allele</th>
            <th>B Allele</th>
            <th>Array</th>
        </tr>
    </thead>
    {{range $id, $marker := .Markers}}
    <tr>
        <td>
            {{.MarkerName}}
        </td>
        <td> {{.RSId}}</td>
        
        <td>
            {{.Chromosome}}
        </td>
        <td>
            {{.Position}}
        </td>
        <td>
            {{.A_Allele}}
        </td>
        <td>
            {{.B_Allele}}
        </td>
        <td>
            {{.Arrays}}
        </td>
    </tr>
    {{end}}
</table>
<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js" integrity="sha384-3ceskX3iaEnIogmQchP8opvBy3Mi7Ce34nWjpBIwVTHfGYWQS9jwHDVRnpKKHJg7" crossorigin="anonymous"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.3.7/js/tether.min.js" integrity="sha384-XTs3FgkjiBgo8qjEjBk0tGmf3wPrWtA6coPfQDfFEY8AnYJwjalXCiosYRBIBZX8" crossorigin="anonymous"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/js/bootstrap.min.js" integrity="sha384-BLiI7JTZm+JWlgKa0M0kGRpJbF2J8q+qreVrKBC47e3K6BW78kGLrCkeRX6I9RoK" crossorigin="anonymous"></script>
<script src="https://cdn.datatables.net/1.10.12/js/jquery.dataTables.min.js" </script>
<script src="https://cdn.datatables.net/1.10.12/js/dataTables.bootstrap.min.js" </script>
<script type="text/javascript">
$(document).ready(function() {
    $('#markertable').DataTable();
});
</script>

`
)
const (
	FormTemplate = `
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/css/bootstrap.min.css" integrity="sha384-AysaV+vQoT3kOAXZkl02PThvDr8HYKPZhNT5h/CXfBThSRXQ6jW5DO2ekP5ViFdi" crossorigin="anonymous">
<link rel="stylesheet" href="https://cdn.datatables.net/1.10.12/css/dataTables.bootstrap.min.css">
<form action="/query/" method="post">
    <div class="form-group">
        <label for="typeSelect">Type of list</label>
        <select class="form-control" id="typeSelect" name="type">
          	<option>Probeset Id</option>
            <option>TODO UCSC Regions</option>
            <option>TODO rsIDs</option>
        </select>
    </div>
    <div class="form-group">
        <label for="exampleTextarea">List (one per line)</label>
        <textarea class="form-control" name="list" id="exampleTextarea" rows="3"></textarea>
    </div>
    <button type="submit" class="btn btn-primary" value="query">Submit</button>
</form>
`
)

//start the url handlers
func init() {

	//json formatted response
	http.HandleFunc("/markerqueryraw/", queryRaw)
	//html response
	http.HandleFunc("/markerquery/", queryPrint)
	http.HandleFunc("/query/", query)

	//load the stuff into the db
	http.HandleFunc("/", populate)
}

// for a json like response - curl friendly
func queryRaw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	markers := queryMarker(w, r)
	markerJSON, err := json.Marshal(markers) // return bytes, err
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(markerJSON)
	}
}
func query(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(FormTemplate)).Execute(w, nil)

	if strings.TrimSpace(r.FormValue("list")) != "" {
		var markers []Marker
		switch formType := r.FormValue("type"); formType {
		case "Probeset Id":
			markerNames := strings.Split(r.FormValue("list"), "\n")
			markers = getMarkerInfo(markerNames, c)
		default:
			w.Write([]byte("Unhandled " + r.FormValue("type")))
			// id := strings.Split(parts[2], ",")
		}
		printMarkers(markers, w)
	}
}

func printMarkers(markers []Marker, w http.ResponseWriter) {
	data := struct {
		Markers []Marker
	}{
		markers,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, data)
}

// for a web page like response
func queryPrint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	markers := queryMarker(w, r)

	data := struct {
		Markers []Marker
	}{
		markers,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, data)

}

//populator
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

// forms the marker key
func markerKey(c appengine.Context, markerName string) *datastore.Key {
	return datastore.NewKey(c, "Markers", strings.TrimSpace(markerName), 0, nil)
}

// parse the request, return all results
func queryMarker(w http.ResponseWriter, r *http.Request) []Marker {
	c := appengine.NewContext(r)
	parts := strings.Split(r.URL.Path, "/")
	id := strings.Split(parts[2], ",")
	return getMarkerInfo(id, c)

}

//load the info for a particular marker
func getMarkerInfo(markerNames []string, c appengine.Context) []Marker {
	var markers []Marker
	for _, markerName := range markerNames {
		if strings.TrimSpace(markerName) != "" {

			marker := Marker{}
			err := datastore.Get(c, markerKey(c, markerName), &marker)
			if err != nil {
				marker.MarkerName = markerName + "(" + err.Error() + ")"
			}
			markers = append(markers, marker)
		}
	}
	return markers
}
