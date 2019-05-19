package main

import (
  "net/http"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Star struct {
  Name string `gorm:"primary_key"`
  Description string
  URL string
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

  // Migrate the schema
  a.DB.AutoMigrate(&Star{})
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
  // Create a test Star.
  a.DB.Create(&Star{Name: "test"})

  // Read from DB.
  var star Star
  a.DB.First(&star, "name = ?", "test")

  // Write to HTTP response.
  w.WriteHeader(200)
  w.Write([]byte(star.Name))

  // Delete.
  a.DB.Delete(&star)
}

func main() {
  a := &App{}
  a.Initialize("sqlite3", "test.db")

  http.HandleFunc("/", a.indexHandler)
  if err := http.ListenAndServe(":8080", nil); err != nil {
    panic(err)
  }

  defer a.DB.Close()
}
