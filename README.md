# Desafio Cotação Dólar Go

Este projeto implementa um sistema cliente-servidor em Go para obter a cotação atual do Dólar (USD-BRL), persistir o histórico no servidor e salvar a cotação mais recente no cliente. O projeto demonstra o uso de:

*   Servidor HTTP em Go (`net/http`)
*   Consumo de API externa
*   Manipulação de JSON
*   Banco de dados SQLite (`database/sql`, `github.com/mattn/go-sqlite3`)
*   Gerenciamento de timeouts com `context`
*   Manipulação de arquivos (`os`)

## Funcionalidades

*   **`server.go`**:
    *   Expõe um endpoint `/cotacao` na porta `8080`.
    *   Busca a cotação atual USD-BRL da API `https://economia.awesomeapi.com.br/json/last/USD-BRL`.
        *   Timeout para a API externa: 200ms.
    *   Salva cada cotação buscada (valor `bid`) em um banco de dados SQLite (`cotacoes.db`).
        *   Timeout para persistência no banco: 10ms.
    *   Retorna apenas o valor `bid` da cotação em formato JSON para o cliente.
    *   Loga erros caso os timeouts sejam excedidos.
*   **`client.go`**:
    *   Faz uma requisição GET para `http://localhost:8080/cotacao`.
        *   Timeout total para a requisição: 300ms.
    *   Recebe o valor `bid` do servidor.
    *   Salva a cotação recebida no arquivo `cotacao.txt` no formato: `Dólar: {valor}`.
    *   Loga um erro caso o timeout da requisição seja excedido.

## Pré-requisitos

*   Go (versão 1.18 ou superior recomendado)

## Como Configurar e Executar

1.  **Clone o repositório ou crie os arquivos:**
    Certifique-se de ter os arquivos `server.go`, `client.go` e `go.mod` no mesmo diretório.

2.  **Baixe as dependências:**
    Abra o terminal no diretório do projeto e execute:
    ```bash
    go mod tidy
    ```
    ou
    ```bash
    go get github.com/mattn/go-sqlite3
    ```

3.  **Execute o Servidor:**
    Abra um terminal no diretório do projeto e execute:
    ```bash
    go run server.go
    ```
    O servidor começará a escutar na porta 8080. Mantenha este terminal aberto.

4.  **Execute o Cliente:**
    Abra **outro** terminal no diretório do projeto e execute:
    ```bash
    go run client.go
    ```
    O cliente fará a requisição ao servidor. Se bem-sucedido dentro do timeout de 300ms, ele criará/atualizará o arquivo `cotacao.txt` com a cotação atual.

## Saída Esperada

*   **No terminal do servidor:** Logs indicando inicialização, requisições recebidas e possíveis erros de timeout (especialmente do banco de dados, devido ao limite de 10ms).
*   **No terminal do cliente:** Log indicando que a cotação foi salva com sucesso ou um erro de timeout.
*   **Arquivo `cotacoes.db`:** Será criado no diretório do projeto pelo servidor para armazenar o histórico das cotações (se o salvamento for bem-sucedido).
*   **Arquivo `cotacao.txt`:** Será criado/atualizado no diretório do projeto pelo cliente, contendo a linha `Dólar: X.XXXX`.

## Notas

*   Os timeouts definidos são intencionalmente curtos para demonstrar o funcionamento do `context` e o tratamento de erros de timeout. É esperado que o timeout de escrita no banco de dados (10ms) possa falhar ocasionalmente.
