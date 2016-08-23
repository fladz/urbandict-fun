package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
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

var (
	port string
	term string
	rd   = rand.New(rand.NewSource(time.Now().Unix()))
)
var terms = []string{"fml", "mudkipz", "omg", "lel"}

const path = "http://api.urbandictionary.com/v0/define"

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

	// Start up web server.
	http.ListenAndServe(":"+port, muxRouters())
}

func muxRouters() *mux.Router {
	mx := mux.NewRouter()
	// Default path
	mx.HandleFunc("/", ShowLanding)
	// Others - the term should be in path.
	mx.HandleFunc("/{term}", GetDefinition)
	return mx
}

// This is the first landing page, or the user didn't provide search term.
func ShowLanding(w http.ResponseWriter, r *http.Request) {
	// Get a term from our perdefined list.
	term = getRandomTerm()

	// Get definition of this term.
	def, err := getDefinition(term)
	if err != nil || def == nil {
		shrug(w)
		return
	}

	showDefinition(w, term, def)
}

func GetDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	term := vars["term"]
	if term == "" {
		ShowLanding(w, r)
		return
	}

	// Get definition of this term.
	def, err := getDefinition(term)
	if err != nil || def == nil {
		shrug(w)
		return
	}

	showDefinition(w, term, def)
}

func shrug(w http.ResponseWriter) {
	w.Write([]byte("You don't get any coolio word ¯\\_(ツ)_/¯"))
}

func showDefinition(w http.ResponseWriter, term string, definition *Definition) {
	w.Write([]byte(fmt.Sprintf("Definition for %s\n\n", term)))
	w.Write([]byte(fmt.Sprintf("Definition:\n%s\n\n", definition.Definition)))
	w.Write([]byte(fmt.Sprintf("Example:   \n%s\n", definition.Example)))
}

func getRandomTerm() string {
	return terms[rd.Intn(len(terms)-1)]
}

func getDefinition(word string) (*Definition, error) {
	res, err := http.Get(path + "?term=" + word)
	if err != nil {
		return nil, fmt.Errorf("Error sending request - %s\n", err)
	}
	defer res.Body.Close()

	var jsonResponse = UrbanDictResponse{}
	if err = json.NewDecoder(res.Body).Decode(&jsonResponse); err != nil {
		return nil, fmt.Errorf("Error reading response - %s\n", err)
	}

	if len(jsonResponse.List) == 0 {
		return nil, nil
	}
	if len(jsonResponse.List) == 1 {
		return &jsonResponse.List[0], nil
	}

	return &jsonResponse.List[rd.Intn(len(jsonResponse.List)-1)], nil
}
