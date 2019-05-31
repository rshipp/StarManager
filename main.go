package main

import (
  "fmt"
  "net/http"
  "net/url"
  "encoding/json"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
  "github.com/gorilla/mux"
)

type Star struct {
  Name string `gorm:"primary_key" json:"name"`
  Description string `json:"description"`
  URL string `json:"url"`
}

type App struct {
  DB *gorm.DB
}

func (a *App) Initialize(dbDriver string, dbURI string) {
  db, err := gorm.Open(dbDriver, dbURI)
  if err != nil {
    panic("failed to connect database")
  }
  a.DB = db

  // Migrate the schema.
  a.DB.AutoMigrate(&Star{})
}

func (a *App) ListHandler(w http.ResponseWriter, r *http.Request) {
  var stars []Star

  // Select all stars and convert to JSON.
  a.DB.Find(&stars)
  starsJSON, _ := json.Marshal(stars)

  // Write to HTTP response.
  w.WriteHeader(200)
  w.Write([]byte(starsJSON))
}

func (a *App) ViewHandler(w http.ResponseWriter, r *http.Request) {
  var star Star
  vars := mux.Vars(r)

  // Select the star with the given name, and convert to JSON.
  a.DB.First(&star, "name = ?", vars["name"])
  starJSON, _ := json.Marshal(star)

  // Write to HTTP response.
  w.WriteHeader(200)
  w.Write([]byte(starJSON))
}

func (a *App) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // Parse the POST body to populate r.PostForm.
  if err := r.ParseForm(); err != nil {
    panic("failed in ParseForm() call")
  }

  // Create a new star from the request body.
  star := &Star{
    Name: r.PostFormValue("name"),
    Description: r.PostFormValue("description"),
    URL: r.PostFormValue("url"),
  }
  a.DB.Create(star)

  // Form the URL of the newly created star.
  u, err := url.Parse(fmt.Sprintf("/stars/%s", star.Name))
  if err != nil {
    panic("failed to form new Star URL")
  }
  base, err := url.Parse(r.URL.String())
  if err != nil {
    panic("failed to parse request URL")
  }

  // Write to HTTP response.
  w.Header().Set("Location", base.ResolveReference(u).String())
  w.WriteHeader(201)
}

func (a *App) UpdateHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  // Parse the POST body to populate r.PostForm.
  if err := r.ParseForm(); err != nil {
    panic("failed in ParseForm() call")
  }

  // Set new star values from the request body.
  star := &Star{
    Name: r.PostFormValue("name"),
    Description: r.PostFormValue("description"),
    URL: r.PostFormValue("url"),
  }

  // Update the star with the given name.
  a.DB.Model(&star).Where("name = ?", vars["name"]).Updates(&star)

  // Write to HTTP response.
  w.WriteHeader(204)
}

func (a *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  // Delete the star with the given name.
  a.DB.Where("name = ?", vars["name"]).Delete(Star{})

  // Write to HTTP response.
  w.WriteHeader(204)
}

func main() {
  a := &App{}
  a.Initialize("sqlite3", "test.db")

  r := mux.NewRouter()

  r.HandleFunc("/stars", a.ListHandler).Methods("GET")
  r.HandleFunc("/stars/{name:.+}", a.ViewHandler).Methods("GET")
  r.HandleFunc("/stars", a.CreateHandler).Methods("POST")
  r.HandleFunc("/stars/{name:.+}", a.UpdateHandler).Methods("PUT")
  r.HandleFunc("/stars/{name:.+}", a.DeleteHandler).Methods("DELETE")

  http.Handle("/", r)
  if err := http.ListenAndServe(":8080", nil); err != nil {
    panic(err)
  }

  defer a.DB.Close()
}
