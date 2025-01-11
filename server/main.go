package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CotacaoRecebida struct {
	USDBRL struct {
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
		Create_date string `json:"create_date"`
	}
}

type Cotacao struct {
	ID         int `gorm:"primaryKey"`
	Code       string
	Codein     string
	Name       string
	High       string
	Low        string
	VarBid     string
	PctChange  string
	Bid        string
	Ask        string
	Timestamp  string
	CreateDate string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)

	fmt.Println("Servidor up")
	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := BuscaCotacao(ctx)
	if err != nil {
		http.Error(w, "Erro ao buscar cotação", http.StatusInternalServerError)
		return
	}

	ctxDB, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = SalvaDB(ctxDB, cotacao)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao salvar no banco: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(cotacao)
	if err != nil {
		log.Printf("Erro ao codificar JSON: %v", err)
		http.Error(w, "Erro ao processar resposta", http.StatusInternalServerError)
	}
}

func BuscaCotacao(ctx context.Context) (*Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacaoRecebida CotacaoRecebida
	err = json.Unmarshal(body, &cotacaoRecebida)
	if err != nil {
		return nil, err
	}

	cotacao := Cotacao{
		Code:       cotacaoRecebida.USDBRL.Code,
		Codein:     cotacaoRecebida.USDBRL.Codein,
		Name:       cotacaoRecebida.USDBRL.Name,
		High:       cotacaoRecebida.USDBRL.High,
		Low:        cotacaoRecebida.USDBRL.Low,
		VarBid:     cotacaoRecebida.USDBRL.VarBid,
		PctChange:  cotacaoRecebida.USDBRL.PctChange,
		Bid:        cotacaoRecebida.USDBRL.Bid,
		Ask:        cotacaoRecebida.USDBRL.Ask,
		Timestamp:  cotacaoRecebida.USDBRL.Timestamp,
		CreateDate: cotacaoRecebida.USDBRL.Create_date,
	}

	return &cotacao, nil
}

func SalvaDB(ctx context.Context, cotacao *Cotacao) error {
	dsn := "root:root@tcp(localhost:3306)/DB_desafio1"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Erro ao abrir o BD: %v", err)
		return err
	}

	db.AutoMigrate(&Cotacao{})

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Timeout do banco de dados: %v", ctx.Err())
		return ctx.Err()
	}

	err = db.WithContext(ctx).Create(cotacao).Error
	if err != nil {
		log.Printf("Erro ao salvar cotação: %v", err)
		return err
	}

	return nil
}
