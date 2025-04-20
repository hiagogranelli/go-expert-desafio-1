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

const (
	serverURL      = "http://localhost:8080/cotacao"
	requestTimeout = 300 * time.Millisecond
	outputFile     = "cotacao.txt"
)

type ServerResponse struct {
	Bid string `json:"bid"`
}

func main() {
	// 1. Contexto para a requisição HTTP (300ms)
	reqCtx, reqCancel := context.WithTimeout(context.Background(), requestTimeout)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", serverURL, nil)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if reqCtx.Err() == context.DeadlineExceeded {
			log.Fatalf("Erro: Timeout ao conectar com o servidor (%v)", requestTimeout)
		}
		log.Fatalf("Erro ao fazer requisição ao servidor: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Erro: Servidor retornou status %d. Resposta: %s", resp.StatusCode, string(bodyBytes))
	}

	var serverResp ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&serverResp); err != nil {
		log.Fatalf("Erro ao decodificar JSON da resposta do servidor: %v", err)
	}

	fileContent := fmt.Sprintf("Dólar: %s", serverResp.Bid)

	err = os.WriteFile(outputFile, []byte(fileContent), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar cotação no arquivo '%s': %v", outputFile, err)
	}

	log.Printf("Cotação salva com sucesso em %s: %s", outputFile, fileContent)
}
