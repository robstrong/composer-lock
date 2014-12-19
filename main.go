package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	bindAddress  = kingpin.Flag("bind-address", "Bind address").Default("localhost").String()
	port         = kingpin.Flag("port", "Port").Default("5799").Int()
	composerPath string
)

func main() {
	initialize()
	startServer()
}

func startServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//we will always return JSON
		w.Header().Set("Content-Type", "application/json")

		//must be a POST request
		if r.Method != "POST" {
			jsonError(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		//read in the JSON from body
		js, err := ioutil.ReadAll(r.Body)
		if err != nil || !isJSON(js) {
			jsonError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		//generate the lock file
		lock, err := GetLockJson(js)
		if err != nil {
			jsonError(w, "Could not generate JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		//return the lock file in the response
		_, err = w.Write(lock)
		if err != nil {
			log.Print(err.Error())
		}
	})
	log.Printf("Starting Server on %s:%d", *bindAddress, *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func jsonError(w http.ResponseWriter, msg string, httpCode int) {
	resp := ErrorResponse{
		Status: "error",
		Detail: msg,
	}
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(js)
	if err != nil {
		log.Print(err.Error())
	}
}

type ErrorResponse struct {
	Status string
	Detail string
}

func initialize() {
	//setup CLI
	kingpin.Version("0.0.1")
	kingpin.Parse()

	//find composer binary path
	var err error
	composerPath, err = exec.LookPath("composer")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Using Composer binary: %s", composerPath)
}

func isJSON(j []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(j, &js) == nil
}

func GetLockJson(c []byte) ([]byte, error) {
	path, err := ioutil.TempDir("", "composer-lock")
	if err != nil {
		return []byte{}, err
	}
	defer os.RemoveAll(path)
	log.Printf("Using path %s", path)

	ioutil.WriteFile(path+"/composer.json", c, 0777)

	//run composer install
	log.Print("Start composer install")
	var out bytes.Buffer
	cmd := exec.Command(composerPath, "install")
	cmd.Dir = path
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return []byte{}, errors.New(out.String())
		} else {
			return []byte{}, err
		}
	}
	log.Printf("Composer install done")

	//read composer.lock
	return ioutil.ReadFile(path + "/composer.lock")
}
