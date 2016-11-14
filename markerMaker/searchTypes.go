package markerMaker

type NavBar struct {
	Href           string
	Name           string
	Content        string
	Active         string
	Submit         string
	SuggestedValue string
}

var navs = []NavBar{
	NavBar{
		"probeset",
		"Probeset ID",
		"Search by Probeset ID",
		"Yes",
		"/probesetq/",
		"SNP_A-1780903",
	},
	NavBar{
		"ucsc",
		"UCSC Region",
		"Search by UCSC Region",
		"No",
		"/ucscq/",
		"chr17:7,571,720-7,590,868",
	},
	NavBar{
		"rsid",
		"rs ID",
		"Search by rs ID",
		"No",
		"/rsidq/",
		"rs1998081",
	},
}
