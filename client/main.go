package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoRecebida struct {
	Bid string `json:"bid"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: tempo de execução excedeu o limite!")
		} else {
			panic(err)
		}
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var cotacao CotacaoRecebida
	error := json.Unmarshal(body, &cotacao)
	if error != nil {
		panic(error)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Dólar:%s", cotacao.Bid))
	if err != nil {
		panic(err)
	}
}
