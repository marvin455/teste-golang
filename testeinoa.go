package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"time"
)

// Estrutura para armazenar os dados da resposta da API
type AlphaVantageResponse struct {
	Metadata struct {
		Information string `json:"1. Information"`
		Symbol      string `json:"2. Symbol"`
	} `json:"Meta Data"`
	TimeSeries map[string]struct {
		Close string `json:"4. close"`
	} `json:"Time Series (Daily)"`
}

var ultimaRecomendacao string // Guarda a última ação recomendada

// Função para buscar dados da ação
func buscarDadosAcao(symbol string) (*AlphaVantageResponse, error) {
	apiKey := "SUA_CHAVE_API_AQUI" // Insira sua chave de API aqui
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=TSCO.LON&outputsize=full&apikey=demo", symbol, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data AlphaVantageResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func deveComprarVender(dados *AlphaVantageResponse) string {
	var ultimoPreco string
	for _, v := range dados.TimeSeries {
		ultimoPreco = v.Close
		break
	}

	// Verifica se ultimoPreco está vazio
	if ultimoPreco == "" {
		log.Println("Preço vazio recebido da API")
		return "erro"
	}

	// Converte o preço para float
	precoFloat, err := strconv.ParseFloat(ultimoPreco, 64)
	if err != nil {
		log.Printf("Erro ao converter preço: %v", err)
		return "erro"
	}

	if precoFloat > 100 {
		return "compra"
	}
	return "venda"
}

// Função para enviar um e-mail com a recomendação
func enviarEmail(mensagem string) error {
	smtpServer := "smtp.gmail.com"
	smtpPort := "587"
	from := "lmarquinhos31@gmail.com"
	to := "mvlap77@gmail.com"
	subject := "Alerta de Ações"
	body := fmt.Sprintf("Assunto: %s\n\n%s", subject, mensagem)

	auth := smtp.PlainAuth("", from, "zlor mmkg ydci tnfy", smtpServer)

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, from, []string{to}, []byte(body))
	return err
}

// Função para monitorar as ações e enviar alertas por e-mail a cada 20 segundos
func monitorarAcoes(symbol string) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dados, err := buscarDadosAcao(symbol)
		if err != nil {
			log.Printf("Erro ao buscar dados da ação: %v", err)
			continue
		}

		acao := deveComprarVender(dados)

		if acao != ultimaRecomendacao {
			mensagem := ""
			if acao == "compra" {
				mensagem = fmt.Sprintf("É o momento de comprar a ação %s!", symbol)
			} else {
				mensagem = fmt.Sprintf("É o momento de vender a ação %s!", symbol)
			}

			err := enviarEmail(mensagem)
			if err != nil {
				log.Printf("Erro ao enviar e-mail: %v", err)
			} else {
				log.Printf("E-mail enviado: %s", mensagem)
			}

			ultimaRecomendacao = acao
		}
	}
}

// Função que responde a requisições HTTP
func handleRequest(w http.ResponseWriter, r *http.Request) {
	symbol := "FTSE 100 (FTSE)"

	go monitorarAcoes(symbol)

	fmt.Fprintf(w, "Monitorando a ação %s. Você receberá um e-mail quando houver uma mudança.", symbol)
}

func main() {
	ultimaRecomendacao = ""

	http.HandleFunc("/verificar", handleRequest)

	log.Println("Servidor iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
