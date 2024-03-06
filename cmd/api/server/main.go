package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	url = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

type Price struct {
	Code        string `json:"code"`
	Codein      string `json:"codein"`
	Name        string `json:"name"`
	High        string `json:"high"`
	Low         string `json:"low"`
	VarBid      string `json:"varBid"`
	PctChange   string `json:"pctChange"`
	Bid         string `json:"bid"`
	Ask         string `json:"ask"`
	Timestamp   string `json:"timestamp"`
	CreateDate  string `json:"create_date"`
	RequestDate string `json:"request_date"`
}

type PriceData struct {
	USDBRL Price `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer log.Println("Request finished...")

	log.Println("Request received on server...")

	ctxHttp, ctxHttpCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer ctxHttpCancel()
	req, err := http.NewRequestWithContext(ctxHttp, http.MethodGet, url, nil)
	if err != nil {
		treatError(err, w)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		treatError(err, w)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		treatError(err, w)
		return
	}

	var price PriceData
	if err = json.Unmarshal(body, &price); err != nil {
		treatError(err, w)
		return
	}

	if err := savePrice(price.USDBRL); err != nil {
		treatError(err, w)
		return
	}

	log.Println("Request processed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(price.USDBRL.Bid)
}

func savePrice(price Price) error {
	dsn := "root:root@tcp(localhost:3306)/cotacao"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&Price{})

	price.RequestDate = time.Now().Format("2006-01-02 15:04:05")

	log.Println("Saving price on database...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	tx := db.WithContext(ctx).Create(price)
	return tx.Error
}

func treatError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}
