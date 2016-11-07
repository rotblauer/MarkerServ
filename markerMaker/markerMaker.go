package markerMaker

import (
	"bufio"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
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

//For easy transfer
const (
	markerPositionsURL = "http://genvisis.org/rsrc/Arrays/AffySnp6/hg19_markerPositions.txt"
)

//template html for displaying results
const (
	PrintTemplate = `
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-alpha.5/css/bootstrap.min.css" integrity="sha384-AysaV+vQoT3kOAXZkl02PThvDr8HYKPZhNT5h/CXfBThSRXQ6jW5DO2ekP5ViFdi" crossorigin="anonymous">
<link rel="stylesheet" href="https://cdn.datatables.net/1.10.12/css/dataTables.bootstrap.min.css">
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<meta http-equiv="x-ua-compatible" content="ie=edge">
<h1>Marker Query</h1>
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

//start the url handlers
func init() {
	//json formatted response
	http.HandleFunc("/markerqueryraw/", queryRaw)
	//html response
	http.HandleFunc("/markerquery/", queryPrint)
	//load the url(s) into the db
	http.HandleFunc("/loaddata/", loadData)
}

// for a json like response - curl friendly
func queryRaw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	markers, err := queryMarker(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		markerJSON, err := json.Marshal(markers) // return bytes, err
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(markerJSON)
		}
	}
}

// for a web page like response
func queryPrint(w http.ResponseWriter, r *http.Request) {
	markers, err := queryMarker(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		data := struct {
			Markers []Marker
		}{
			markers,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, data)
	}
}

//populator
func loadData(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	c.Infof("Started data import")
	//Hard coded a bit
	loadFromUrl(c, w, markerPositionsURL, "Affymetrix SNP 6.0")

}

//populate the database - currently from a remote url
func loadFromUrl(c appengine.Context, w http.ResponseWriter, url string, array string) {
	// client := urlfetch.Client(c)
	duration, _ := time.ParseDuration("1m")

	client := &http.Client{
		Transport: &urlfetch.Transport{
			Context:                       c,
			Deadline:                      duration,
			AllowInvalidServerCertificate: true,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "\t")
		if strings.HasPrefix(line[0], "SNP") {
			count++
			if count%100 == 0 {
				c.Infof("> Loaded: %s : %d", line[0], count)
			}
			// if count < 1000 {
			// time.Sleep(700 * time.Millisecond)
			tmp, err := parseMarker(line)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				a := make([]string, 1)
				a[0] = array
				tmp.Arrays = a

				key := markerKey(c, tmp.MarkerName)
				if _, err := datastore.Put(c, key, &tmp); err != nil { //store it
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// }
			}
		}
	}
}

// forms the marker key
func markerKey(c appengine.Context, markerName string) *datastore.Key {
	return datastore.NewKey(c, "Markers", markerName, 0, nil)
}

// parse the request, return all results
func queryMarker(w http.ResponseWriter, r *http.Request) ([]Marker, error) {
	c := appengine.NewContext(r)
	parts := strings.Split(r.URL.Path, "/")
	id := strings.Split(parts[2], ",")
	c.Infof("> Marker: [%s]", id)
	defer c.Infof("Marker loaded")
	return getMarkerInfo(id, c, w)

}

//load the info for a particular marker
func getMarkerInfo(markerNames []string, c appengine.Context, w http.ResponseWriter) ([]Marker, error) {
	markers := make([]Marker, len(markerNames))
	for i, markerName := range markerNames {
		marker := Marker{}
		err := datastore.Get(c, markerKey(c, markerName), &marker)
		if err != nil {
			return nil, err
		} else {
			markers[i] = marker
		}
	}
	return markers, nil
}

//parse a marker from a string array
func parseMarker(line []string) (Marker, error) {
	pos, err := strconv.Atoi(line[2])
	marker := Marker{
		MarkerName: line[0],
		Chromosome: line[1],
		Position:   pos,
	}
	return marker, err
}
