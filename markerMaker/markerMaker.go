package markerMaker

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
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

<div class="container">
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
</div>

<script type="text/javascript">
$(document).ready(function() {
    $('#markertable').DataTable();
});
</script>

`
)
const (
	FormTemplate = `

<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<meta http-equiv="x-ua-compatible" content="ie=edge">

<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/css/bootstrap.min.css" integrity="sha384-AysaV+vQoT3kOAXZkl02PThvDr8HYKPZhNT5h/CXfBThSRXQ6jW5DO2ekP5ViFdi" crossorigin="anonymous">
<link rel="stylesheet" href="https://cdn.datatables.net/1.10.12/css/dataTables.bootstrap.min.css">
<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js" integrity="sha384-3ceskX3iaEnIogmQchP8opvBy3Mi7Ce34nWjpBIwVTHfGYWQS9jwHDVRnpKKHJg7" crossorigin="anonymous"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/tether/1.3.7/js/tether.min.js" integrity="sha384-XTs3FgkjiBgo8qjEjBk0tGmf3wPrWtA6coPfQDfFEY8AnYJwjalXCiosYRBIBZX8" crossorigin="anonymous"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/js/bootstrap.min.js" integrity="sha384-BLiI7JTZm+JWlgKa0M0kGRpJbF2J8q+qreVrKBC47e3K6BW78kGLrCkeRX6I9RoK" crossorigin="anonymous"></script>
<script src="https://cdn.datatables.net/1.10.12/js/jquery.dataTables.min.js" </script>
<script src="https://cdn.datatables.net/1.10.12/js/dataTables.bootstrap.min.js"</script>

<div class="container">
    <div class="jumbotron">
        <h1>Marker query example</h1>
        <p>Search for markers by Probeset Id, <a href="https://genome.ucsc.edu/FAQ/FAQformat#format1">BED Region</a>, or rsId</p>
        <div class="container">
            <form action="/query/" method="post">
                <div class="row">
                    <div class="col-sm-4">
                        <div class="form-group">
                            <label for="typeSelect">Type of list</label>
                            <select class="form-control" id="typeSelect" name="type">
                                <option>Probeset Id</option>
                                <option>BED Region</option>
                                <option>TODO rsIDs</option>
                            </select>
                        </div>
                    </div>
                </div>
                <div class="row">
                    <div class="col-sm-4">
                        <div class="form-group">
                            <label for="exampleTextarea">List (one per line)</label>
                            <textarea class="form-control" name="list" id="exampleTextarea" rows="3"></textarea>
                        </div>
                        <button type="submit" class="btn btn-primary" value="query">Submit</button>
                    </div>
                </div>
            </form>
        </div>
    </div>
</div>



`
)

//start the url handlers
func init() {

	//json formatted response
	http.HandleFunc("/markerqueryraw/", queryRaw)
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

	if strings.TrimSpace(r.FormValue("list")) != "" {
		var markers []Marker
		switch formType := r.FormValue("type"); formType {
		case "Probeset Id":
			markers = queryByNames(strings.Split(r.FormValue("list"), "\n"), c)
		case "BED Region":
			bed, err := parseBed3(strings.Split(r.FormValue("list"), "\n")[0])
			if err == nil {
				markers = queryByPosition(bed.Chrom, bed.Start(), bed.End(), c)
			}
		default:
			w.Write([]byte("Unhandled " + r.FormValue("type")))
			// id := strings.Split(parts[2], ",")
		}
		template.Must(template.New("Data").Parse(FormTemplate)).Execute(w, nil)

		printMarkers(markers, w)
	} else {
		template.Must(template.New("Data").Parse(FormTemplate)).Execute(w, nil)

	}
}

// 1	10000	500000
func queryByPosition(chr string, start int, stop int, c appengine.Context) []Marker {

	// The Query type and its methods are used to construct a query.
	q := datastore.NewQuery("Markers").
		Filter("Chromosome =", chr).
		Filter("Position <=", stop).
		Filter("Position >=", start)

	var markers []Marker
	t := q.Run(c)
	for {
		var marker Marker
		_, err := t.Next(&marker)
		if err == datastore.Done {
			break
		}
		if err != nil {
			break
		}
		markers = append(markers, marker)
	}
	return markers
}

const (
	chromField = iota
	startField
	endField
)

type Bed3 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
}

func parseBed3(line string) (b *Bed3, err error) {
	const n = 3
	f := strings.Split(line, "\t")
	if len(f) < n {
		return nil, errors.New("bed: bad bed type")
	}
	b = &Bed3{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField], startField),
		ChromEnd:   mustAtoi(f[endField], endField),
	}
	return
}
func mustAtoi(f string, column int) int {
	i, err := strconv.ParseInt(f, 0, 0)
	if err != nil {
		panic(&csv.ParseError{Column: column, Err: err})
	}
	return int(i)
}

func (b *Bed3) Start() int { return b.ChromStart }
func (b *Bed3) End() int   { return b.ChromEnd }
func (b *Bed3) Len() int   { return b.ChromEnd - b.ChromStart }

func printMarkers(markers []Marker, w http.ResponseWriter) {
	data := struct {
		Markers []Marker
	}{
		markers,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, data)
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
	return queryByNames(id, c)

}

//load the info for a particular marker by marker name
func queryByNames(markerNames []string, c appengine.Context) []Marker {
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
