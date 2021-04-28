package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"veolet/lib"
)

type answer struct {
	Text string
}

func handler(w http.ResponseWriter, r *http.Request) {
	p := answer{Text: "test"}
	json, _ := json.Marshal(p)
	w.Write(json)
}

func oldmain() {

	// Read Flow configurations
	file, _ := os.Open("flow.json")
	defer file.Close()
	byteFile, _ := ioutil.ReadAll(file)
	var configuration lib.FlowConfiguration
	json.Unmarshal(byteFile, &configuration)
	config := lib.GetConfig(configuration, "emulator")
	log.Println(config)

	// API configurations
	port := 80

	http.HandleFunc("/", handler)

	log.Printf("Server listening on port %d", port)
	log.Print(http.ListenAndServe(":80", nil))
}
