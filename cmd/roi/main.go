package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"sort"

	"dev.2lfilm.com/2l/roi"
)

var dev bool
var templates *template.Template

func parseTemplate() {
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	if dev {
		parseTemplate()
	}
	templates.ExecuteTemplate(w, name, data)
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	prj := "test"

	shots, err := roi.SelectShots(db, prj)
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(shots, func(i int, j int) bool {
		if shots[i].Project < shots[j].Project {
			return true
		}
		if shots[i].Project > shots[j].Project {
			return false
		}
		if shots[i].Scene < shots[j].Scene {
			return true
		}
		if shots[i].Scene > shots[j].Scene {
			return false
		}
		return shots[i].Name <= shots[j].Name
	})

	recipt := struct {
		Project string
		Shots   []roi.Shot
	}{
		Project: prj,
		Shots:   shots,
	}
	executeTemplate(w, "index.html", recipt)
}

func main() {
	dev = true
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	roi.CreateTableIfNotExists(db, "projects", roi.ProjectTableFields)
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:7070", mux))
}