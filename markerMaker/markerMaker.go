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
        <tr>
                <th> Marker Name 1</th>
                <th>Rs Id</th>
                <th>Chromosome 1</th>
                <th>Position 1</th>
                <th>A Allele</th>
                <th>B Allele</th>
                <th>Array 1</th>
            </tr>
            <tr>
                <th> Marker Name 2</th>
                <th>Rs Id</th>
                <th>Chromosome 2</th>
                <th>Position</th>
                <th>A Allele</th>
                <th>B Allele</th>
                <th>Array</th>
            </tr>
       
    </table>
</div>
`
)

// {{range $id, $marker := .Markers}}
// <tr>
//     <td>
//         {{HELLO}}
//     </td>
//     <td> {{TWO}}</td>
//     <td>
//         {{.Chromosome}}
//     </td>
//     <td>
//         {{.Position}}
//     </td>
//     <td>
//         {{.A_Allele}}
//     </td>
//     <td>
//         {{.B_Allele}}
//     </td>
//     <td>
//         {{.Arrays}}
//     </td>
// </tr>
// {{end}}
const (
	FormTemplate = `

<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<meta http-equiv="x-ua-compatible" content="ie=edge">
<link rel="stylesheet" href="/bower_components/bootstrap/dist/css/bootstrap.min.css">
<link rel="stylesheet" href="/bower_components/datatables.net-bs/css/dataTables.bootstrap.min.css">
<script src="/bower_components/jquery/dist/jquery.min.js"</script>
<script src="/bower_components/tether/dist/js/tether.min.js"></script>
<script src="/bower_components/bootstrap/dist/js/bootstrap.min.js"></script>
<script src="/bower_components/datatables.net/js/jquery.dataTables.min.js" </script> 
<script src="/bower_components/datatables.net-bs/js/dataTables.bootstrap.min.js"</script>

<script>
$(document).ready(function() {
    $('[data-toggle="popover"]').popover();
});
</script>
<script type="text/javascript">
$(document).ready(function() {
    $('#markertable').DataTable();
});
</script>
<div class="container">
    <div class="jumbotron">
        <h1>SNP Array Search</h1>
        <p>Search for markers by
            <a href="#" title="Probeset Id" data-toggle="popover" data-content="Such as: SNP_A-1782064">Probeset Id</a>,
            <a href="#" title="rsID" data-toggle="popover" data-content="Such as: rs998353">rsID</a>, or
            <a href="#" title="BED Region" data-toggle="popover" data-content="Such as (tab separated): 1   10000   500000">BED Region</a>
        </p>
      {{template "searchBarNav"}}
            
        </div>
    </div>
</div>


`
)

//start the url handlers
func init() {

	// //json formatted response
	// http.HandleFunc("/markerqueryraw/", queryRaw)

	// //load the stuff into the db
	// http.HandleFunc("/populate/", populate)

	// http.HandleFunc("/", query)

	router := NewRouter()
	http.Handle("/", router)
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
