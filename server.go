package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	apiTimeout = 200 * time.Millisecond
	dbTimeout  = 10 * time.Millisecond
	dbFile     = "./cotacoes.db"
	serverPort = ":8080"
	dbSchema   = `CREATE TABLE IF NOT EXISTS cotacoes (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        bid TEXT,
                        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
                      );`
)

type ApiResponse struct {
	USDBRL Usdbrl `json:"USDBRL"`
}

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(dbSchema)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func fetchCotacao(ctx context.Context) (*Usdbrl, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout ao buscar cotação na API externa")
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Erro: API externa retornou status %d", resp.StatusCode)
		return nil, err
	}

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Println("Erro ao decodificar JSON da API:", err)
		return nil, err
	}

	return &apiResp.USDBRL, nil
}

func saveCotacao(db *sql.DB, ctx context.Context, cotacao *Usdbrl) error {
	stmt, err := db.Prepare("INSERT INTO cotacoes (bid) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.Bid)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout ao salvar cotação no banco de dados")
		}
		return err
	}
	return nil
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	apiCtx, apiCancel := context.WithTimeout(r.Context(), apiTimeout)
	defer apiCancel()

	cotacaoData, err := fetchCotacao(apiCtx)
	if err != nil {
		log.Printf("Erro ao buscar cotação: %v", err)
		http.Error(w, "Erro ao buscar cotação externa", http.StatusInternalServerError)
		return
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbTimeout)
	defer dbCancel()

	err = saveCotacao(db, dbCtx, cotacaoData)
	if err != nil {
		log.Printf("Erro ao salvar cotação no banco: %v", err)
	}

	response := CotacaoResponse{Bid: cotacaoData.Bid}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Erro ao enviar resposta JSON ao cliente: %v", err)
	}

	log.Printf("Cotação solicitada e processada: Bid = %s", cotacaoData.Bid)
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Erro ao inicializar banco de dados: %v", err)
	}
	defer db.Close()
	log.Println("Banco de dados inicializado com sucesso.")

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacaoHandler(w, r, db)
	})

	log.Printf("Servidor escutando na porta %s", serverPort)
	err = http.ListenAndServe(serverPort, mux)
	if err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
