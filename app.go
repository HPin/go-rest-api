// app.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	// contains database credentials
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString) // try to open database
	if err != nil {
		log.Fatal(err)
	}

	// instantiate router
	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

// opens and starts the connection on the specified port
// addr contains a colon followed by the port number, e.g. ":8010"
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// handler to get a product from the database
func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // get variables of the request
	id, err := strconv.Atoi(vars["id"]) // convert to int
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id} // instantiate a product with the id from the request
	// try if getting this product fails
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// if getting the product does not fail, return the requested product
	respondWithJSON(w, http.StatusOK, p)
}

// convenience function to send an http error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// convenience function to send an http response with included JSON payload
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// http handler...returns a number of products at once
func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	// convert response variables to integers
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	// check if variables are out of bounds
	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	// get products from the database
	products, err := getProducts(a.DB, start, count)

	// check if database request was unsuccessful
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

// http handler...creates a new product in the database
func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	// create an uninitialzed product
	var p product

	// use a json decoder to map the request body to the product object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close() // make sure to properly close it, even in case of an error

	// try to insert this object into the database
	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

// http handler...updates entries for a given product in the database
func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // get request variables
	id, err := strconv.Atoi(vars["id"]) // convert to int
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// create an uninitialzed product
	var p product

	// use a json decoder to map the request body to the product object
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close() // make sure to properly close it, even in case of an error

	// set the id of the new template product object to the db entry we want to update
	p.ID = id

	// try to update the product in the db
	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// http handler...deletes a product from the database
func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // get all request variables
	id, err := strconv.Atoi(vars["id"]) // convert id to int
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	// create an empty product which only contains the id of the entry we want to delete
	p := product{ID: id}

	// try to delete the product from the database
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// Defines routes for http requests and maps them to handler functions
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}
