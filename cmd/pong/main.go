package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var errNoPong = errors.New("no pong")

type PingResponse struct {
	Pong bool `json:"pong"`
}

func main() {
	server := "http://localhost:8000"

	if os.Getenv("SERVER") != "" {
		server = os.Getenv("SERVER")
	}

	log.Printf("pick server: %s\n", server)

	for i := 1; i < 5; i++ {
		err := sendPing(server, i)

		if err != nil {
			log.Printf("unable to send ping: %v\n", err)
			os.Exit(1)
		}

		log.Printf("[%d/5] PING sent, received PONG: OK\n", i)

		time.Sleep(time.Second)
	}

	err := sendPing(server, 5)

	if errors.Is(err, errNoPong) {
		log.Printf("[5/5] PING sent, received close signal: OK\n")
		os.Exit(0)
	}

	if err == nil {
		log.Printf("[5/5] PING sent, received PONG: FAIL\n")
	} else {
		log.Printf("unable to send ping: %v\n", err)
	}

	os.Exit(1)
}

func sendPing(server string, n int) error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ping", server), nil)
	if err != nil {
		return fmt.Errorf("failed to initiate request: %w", err)
	}

	request.Header.Set("Request-ID", strconv.Itoa(n))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("unexpected content type: %s", response.Header.Get("Content-Type"))
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	var pong PingResponse

	if err := json.Unmarshal(data, &pong); err != nil {
		return fmt.Errorf("unable to unmarshal response body: %w", err)
	}

	if !pong.Pong {
		return errNoPong
	}

	return nil
}
