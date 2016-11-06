package markerMaker

import (
	"bufio"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

// current information returned for query
// queries can be set up to handle pos<X&&chr==2, etc. This would allow getting all markers for an interval across an array
type Marker struct {
	MarkerName string `json:"markerName"`
	Chromosome int    `json:"chromosome"`
	Position   int    `json:"position"`
	A_Allele   string `json:"a_allele"`
	B_Allele   string `json:"b_allele"`
}

//For easy transfer
const (
	markerPositionsURL = "http://genvisis.org/rsrc/Arrays/AffySnp6/hg19_markerPositions.txt"
)

const (
	PrintTemplate = `
    <style>
      table { border-collapse:collapse; }
      table, th, td { border: 1px solid black; padding: .2em; }
      td { vertical-align:top; }
    </style>
    Marker Query:
    <table>
     {{range $id, $marker := .markers}}
       <tr>
         <td>{{$id}}</td>
         <td>
             {{.MarkerName}}, {{.Chromosome}}
             <p/>
         </td>
       </tr>
     {{end}}
    </table>
  `
)

func init() {
	//json formatted response
	http.HandleFunc("/markerqueryraw/", queryRaw)
	//html response
	http.HandleFunc("/markerquery/", queryPrint)
	//load the url into the db
	http.HandleFunc("/loaddata/", loadData)
}

func loadData(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("Started data import")
	client := urlfetch.Client(c)
	resp, err := client.Get(markerPositionsURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		count++
		if count < 1000 {
			tmp, err := parseMarker(strings.Split(scanner.Text(), "\t"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Write([]byte(tmp.MarkerName))
				key := markerKey(c, tmp.MarkerName)
				if _, err := datastore.Put(c, key, &tmp); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}
}

// forms the marker key
func markerKey(c appengine.Context, markerName string) *datastore.Key {
	return datastore.NewKey(c, "Markers", markerName, 0, nil)
}

//parse a marker from a string array
func parseMarker(line []string) (Marker, error) {
	chr, err := strconv.Atoi(line[1])
	pos, err := strconv.Atoi(line[2])
	marker := Marker{
		MarkerName: line[0],
		Chromosome: chr,
		Position:   pos,
	}
	return marker, err
}

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
		template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, data)
	}
}

// parse the request, write the results
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
