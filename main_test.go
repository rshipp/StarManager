package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func setup() *App {
	// Initialize an in-memory database for full integration testing.
	app := &App{}
	app.Initialize("sqlite3", ":memory:")
	return app
}

func teardown(app *App) {
	// Closing the connection discards the in-memory database.
	app.DB.Close()
}

func StarFormValues(star Star) *strings.Reader {
	// Transforms Star record into *strings.Reader suitable for use in HTTP POST forms.
	data := url.Values{
		"name":        {star.Name},
		"description": {star.Description},
		"url":         {star.URL},
	}

	return strings.NewReader(data.Encode())
}

func TestCreateHandler(t *testing.T) {
	app := setup()

	testStar := &Star{
		ID:          1,
		Name:        "test/name",
		Description: "test desc",
		URL:         "test url",
	}

	// Set up a new request.
	req, err := http.NewRequest("POST", "/stars", StarFormValues(*testStar))
	if err != nil {
		t.Fatal(err)
	}
	// Our API expects a form body, so set the content-type header to make sure it's treated as one.
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	http.HandlerFunc(app.CreateHandler).ServeHTTP(rr, req)

	// Test that the status code is correct.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Status code is invalid. Expected %d. Got %d instead", http.StatusCreated, status)
	}

	// Test that the Location header is correct.
	expectedURL := fmt.Sprintf("/stars/%s", testStar.Name)
	if location := rr.Header().Get("Location"); location != expectedURL {
		t.Errorf("Location header is invalid. Expected %s. Got %s instead", expectedURL, location)
	}

	// Test that the created star is correct.
	// Note: There is only one star in the database.
	createdStar := Star{}
	app.DB.First(&createdStar)
	if createdStar != *testStar {
		t.Errorf("Created star is invalid. Expected %+v. Got %+v instead", testStar, createdStar)
	}

	teardown(app)
}

func TestUpdateHandler(t *testing.T) {
	app := setup()

	// Create a star for us to update.
	testStar := &Star{
		ID:          1,
		Name:        "test/name",
		Description: "test desc",
		URL:         "test url",
	}
	app.DB.Create(testStar)

	// Set up a test table.
	starTests := []struct {
		original Star
		update   Star
	}{
		{original: *testStar,
			update: Star{ID: 1, Name: "test/name", Description: "updated desc", URL: "test URL"},
		},
		{original: Star{ID: 1, Name: "test/name", Description: "updated desc", URL: "test URL"},
			update: Star{ID: 1, Name: "updated name", Description: "updated desc", URL: "test URL"},
		},
	}

	for _, tt := range starTests {
		// Set up a new request.
		req, err := http.NewRequest("PUT", fmt.Sprintf("/stars/%s", tt.original.Name), StarFormValues(tt.update))
		if err != nil {
			t.Fatal(err)
		}
		// Our API expects a form body, so set the content-type header appropriately.
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		// We need a mux router in order to pass in the `name` variable.
		r := mux.NewRouter()

		r.HandleFunc("/stars/{name:.*}", app.UpdateHandler).Methods("PUT")
		r.ServeHTTP(rr, req)

		// Test that the status code is correct.
		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Status code is invalid. Expected %d. Got %d instead", http.StatusNoContent, status)
		}

		// Test that the updated star is correct.
		// Note: There is only one star in the database.
		updatedStar := Star{}
		app.DB.First(&updatedStar)
		if updatedStar != tt.update {
			t.Errorf("Updated star is invalid. Expected %+v. Got %+v instead", tt.update, updatedStar)
		}
	}

	teardown(app)
}

func TestViewHandler(t *testing.T) {
	app := setup()

	// Set up a test table.
	starTests := []Star{
		Star{ID: 1, Name: "test/name", Description: "test desc", URL: "test URL"},
		Star{ID: 2, Name: "test/another_name", Description: "test desc 2", URL: "http://example.com/"},
	}

	for _, star := range starTests {
		// Create a star for us to view.
		app.DB.Create(star)

		// Set up a new request.
		req, err := http.NewRequest("GET", fmt.Sprintf("/stars/%s", star.Name), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		// We need a mux router in order to pass in the `name` variable.
		r := mux.NewRouter()

		r.HandleFunc("/stars/{name:.*}", app.ViewHandler).Methods("GET")
		r.ServeHTTP(rr, req)

		// Test that the status code is correct.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Status code is invalid. Expected %d. Got %d instead", http.StatusOK, status)
		}

		// Read the response body.
		data, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Fatal(err)
		}

		// Test that the updated star is correct.
		returnedStar := Star{}
		if err := json.Unmarshal(data, &returnedStar); err != nil {
			t.Errorf("Returned star is invalid JSON. Got: %s", data)
		}
		if returnedStar != star {
			t.Errorf("Returned star is invalid. Expected %+v. Got %+v instead", star, returnedStar)
		}
	}

	teardown(app)
}

func TestListHandler(t *testing.T) {
	app := setup()

	// Create a couple stars to list.
	stars := []Star{
		Star{ID: 1, Name: "test/name", Description: "test desc", URL: "test URL"},
		Star{ID: 2, Name: "test/another_name", Description: "test desc 2", URL: "http://example.com/"},
	}

	for _, star := range stars {
		app.DB.Create(star)
	}

	// Set up a new request.
	req, err := http.NewRequest("GET", "/stars", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	http.HandlerFunc(app.ListHandler).ServeHTTP(rr, req)

	// Test that the status code is correct.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code is invalid. Expected %d. Got %d instead", http.StatusOK, status)
	}

	// Read the response body.
	data, err := ioutil.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// Test that our stars list is the same as what was returned.
	returnedStars := []Star{}
	if err := json.Unmarshal(data, &returnedStars); err != nil {
		t.Errorf("Returned star list is invalid JSON. Got: %s", data)
	}
	if len(returnedStars) != len(stars) {
		t.Errorf("Returned star list is an invalid length. Expected %d. Got %d instead", len(stars), len(returnedStars))
	}
	for index, returnedStar := range returnedStars {
		if returnedStar != stars[index] {
			t.Errorf("Returned star is invalid. Expected %+v. Got %+v instead", stars[index], returnedStar)
		}
	}

	teardown(app)
}

func TestDeleteHandler(t *testing.T) {
	app := setup()

	// Set up a test table.
	starTests := []struct {
		star Star
	}{
		{star: Star{ID: 1, Name: "test/name", Description: "test desc", URL: "test URL"}},
		{star: Star{ID: 2, Name: "test/another_name", Description: "test desc 2", URL: "http://example.com/"}},
	}

	for _, tt := range starTests {
		// Create a star for us to delete.
		app.DB.Create(tt.star)

		// Set up a new request.
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/stars/%s", tt.star.Name), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		// We need a mux router in order to pass in the `name` variable.
		r := mux.NewRouter()

		r.HandleFunc("/stars/{name:.*}", app.DeleteHandler).Methods("DELETE")
		r.ServeHTTP(rr, req)

		// Test that the status code is correct.
		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Status code is invalid. Expected %d. Got %d instead", http.StatusNoContent, status)
		}

		// Test that the star is no longer in the db.
		deletedStar := Star{}
		app.DB.Where("name = ?", tt.star.Name).First(&deletedStar)
		if deletedStar != (Star{}) {
			t.Errorf("Star still exists in db: %+v", tt.star)
		}
	}

	teardown(app)
}
