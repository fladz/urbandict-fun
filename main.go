package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type UrbanDictResponse struct {
	Tags       []string     `json:"tags"`
	ResultType string       `json:"result_type"`
	List       []Definition `json:"list"`
	Sounds     []string     `json:"sounds"`
}
type Definition struct {
	Definition string `json:"definition"`
	Author     string `json:"author"`
	Permalink  string `json:"permalink"`
	Example    string `json:"example"`
	ThumbsUp   int    `json:"thumbs_up"`
	ThumbsDown int    `json:"thumbs_down"`
}

const (
	errTpl = `
<html>
	<body>
	<b>NO COOLIO DEFINITION FOR {{.Term}} ¯\_(ツ)_/¯</b><br>
	</body>
</html>`
	resTpl = `
<html>
	<head>
		<meta charset="UTF-8">
		<title>Definition for {{.Term}}</title>
	</head>
	<body>
		<b>Definition for {{.Term}}</b> (thumbs↑: {{.ThumbsUp}}, thumbs↓: {{.ThumbsDown}})<br><br>
		<b>DEFINITION:</b><br>
		{{.Definition.Definition}}
		<br>
		(Author: {{.Author}})<br><br>
		<b>EXAMPLE:</b><br>
		{{.Example}}<br><br>
		<b>TAG:</b>
		{{range .Tags}}<div><a href="/{{.}}/">{{.}}</a></div>{{end}}
	</body>
</html>`
)

var (
	port    string
	term    string
	rd      = rand.New(rand.NewSource(time.Now().Unix()))
	errTmpl *template.Template
	resTmpl *template.Template
	terms   = []string{"fml", "mudkipz", "omg", "lel"}
)

const path = "http://api.urbandictionary.com/v0/define"

type noRes struct {
	Term string
}
type def struct {
	Term string
	Tags []string
	Definition
}

func init() {
	flag.StringVar(&term, "term", "", "Term you want to know definition for.")
}

func main() {
	flag.Parse()
	if port == "" {
		// Default 8081
		fmt.Println("Port not provided, using 8081")
		port = "8081"
	}

	// Prepare html templates.
	var err error
	if errTmpl, err = template.New("err").Parse(errTpl); err != nil {
		fmt.Println("Error parsing template")
		return
	}
	if resTmpl, err = template.New("res").Parse(resTpl); err != nil {
		fmt.Println("Error parsing template")
		return
	}

	// Start up web server.
	http.ListenAndServe(":"+port, muxRouters())
}

func muxRouters() *mux.Router {
	mx := mux.NewRouter()
	// Default path
	mx.HandleFunc("/", ShowLanding)
	// Others - the term should be in path.
	mx.HandleFunc("/{term}/", GetDefinition)
	return mx
}

// This is the first landing page, or the user didn't provide search term.
func ShowLanding(w http.ResponseWriter, r *http.Request) {
	// Get a term from our perdefined list.
	term = getRandomTerm()

	// Get definition of this term.
	dict, err := getDefinition(term)
	if err != nil || len(dict.List) == 0 {
		shrug(w, noRes{Term: term})
		return
	}

	showDefinition(w, def{term, dict.Tags, dict.List[rd.Intn(len(dict.List)-1)]})
}

func GetDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if term = vars["term"]; term == "" {
		ShowLanding(w, r)
		return
	}

	// Get definition of this term.
	dict, err := getDefinition(term)
	if err != nil || len(dict.List) == 0 {
		shrug(w, noRes{Term: term})
		return
	}

	showDefinition(w, def{term, dict.Tags, dict.List[rd.Intn(len(dict.List)-1)]})
}

func shrug(w http.ResponseWriter, data noRes) {
	if err := errTmpl.Execute(w, data); err != nil {
		w.Write([]byte("¯\\_(ツ)_/¯"))
	}
}

func showDefinition(w http.ResponseWriter, definition def) {
	if err := resTmpl.Execute(w, definition); err != nil {
		w.Write([]byte("¯\\_(ツ)_/¯"))
	}
	//	w.Write([]byte(fmt.Sprintf("Definition for %s\n\n", term)))
	//	w.Write([]byte(fmt.Sprintf("Definition:\n%s\n\n", definition.Definition)))
	//	w.Write([]byte(fmt.Sprintf("Example:   \n%s\n", definition.Example)))
}

func getRandomTerm() string {
	return terms[rd.Intn(len(terms)-1)]
}

func getDefinition(word string) (UrbanDictResponse, error) {
	vals := url.Values{}
	vals.Add("term", word)

	res, err := http.Get(path + "?" + vals.Encode())
	if err != nil {
		return UrbanDictResponse{}, fmt.Errorf("Error sending request - %s\n", err)
	}
	defer res.Body.Close()

	var jsonResponse = UrbanDictResponse{}
	if err = json.NewDecoder(res.Body).Decode(&jsonResponse); err != nil {
		return UrbanDictResponse{}, fmt.Errorf("Error reading response - %s\n", err)
	}

	// Make sure there's no dupe in tags.
	tags := []string{}
	var dupe bool
	for _, tag := range jsonResponse.Tags {
		dupe = false
		if len(tags) == 0 {
			tags = append(tags, tag)
			continue
		}
		for _, t := range tags {
			if t == tag {
				dupe = true
				break
			}
		}
		if !dupe {
			tags = append(tags, tag)
		}
	}
	jsonResponse.Tags = tags

	return jsonResponse, nil
}
