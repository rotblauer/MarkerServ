package markerMaker

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		indexHandler,
	},
	Route{
		"Index",
		"POST",
		"/",
		indexHandler,
	},
	Route{
		"MarkerQuery",
		"POST",
		"/probesetq/",
		markerHandler,
	},
	Route{
		"UCSCQuery",
		"POST",
		"/ucscq/",
		ucscHandler,
	},
	Route{
		"MarkerQueryRaw",
		"GET",
		"/probesetqRaw/{ids}",
		markerHandlerRaw,
	},
	Route{
		"MarkerPopulate",
		"POST",
		"/populate/",
		populate,
	},

	// Route{
	// 	"TodoShow",
	// 	"GET",
	// 	"/todos/{todoId}",
	// 	TodoShow,
	// },
}
