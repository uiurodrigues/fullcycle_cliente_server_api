package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	url            = "http://localhost:8080/cotacao"
	contextTimeout = 300 * time.Second
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	defer log.Println("Request finished...")

	log.Println("Request started...")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error while calling the server API...")
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error... Server returned a status code different from 200")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading server API result...")
		panic(err)
	}
	result := treatResult(string(body))
	log.Printf("Price received:{%v}", result)

	price, err := strconv.ParseFloat(result, 64)
	if err != nil {
		panic(err)
	}

	if err = savePriceOnFile(price); err != nil {
		log.Println("Error while saving prince on file...")
		panic(err)
	}
}

func savePriceOnFile(price float64) error {
	log.Println("Saving price on file...")

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Dólar:{" + strconv.FormatFloat(price, 'g', 5, 64) + "}\n")
	if err != nil {
		return err
	}

	return nil
}

func treatResult(result string) string {
	result = strings.Replace(result, "\n", "", -1)
	result = strings.Replace(result, "\"", "", -1)
	return result
}
