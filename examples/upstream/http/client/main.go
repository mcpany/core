package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("http://localhost:8080/time")
	if err != nil {
		log.Fatalf("Failed to get time: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	fmt.Printf("Current time: %s, Timezone: %s\n", result["current_time"], result["timezone"])
}
