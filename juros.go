package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// cliente da api do grupo de juros, que calcula o calendario de parcelas.
type ClienteJuros struct {
	baseURL string
	http    *http.Client
}

func NovoClienteJuros(baseURL string) *ClienteJuros {
	return &ClienteJuros{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GET {baseURL}/calendario - busca o calendario de parcelas da operacao.
func (cj *ClienteJuros) BuscarCalendario(operacao Operacao) ([]NovaParcela, error) {
	if cj.baseURL == "" {
		return nil, errors.New("JUROS_API_URL não configurada")
	}

	query := url.Values{}
	query.Set("valorFinanciado", strconv.FormatFloat(operacao.ValorFinanciado, 'f', 2, 64))
	query.Set("quantidadeParcelas", strconv.Itoa(operacao.QuantidadeParcelas))
	query.Set("taxaJuros", strconv.FormatFloat(operacao.TaxaJuros, 'f', 4, 64))
	query.Set("primeiroVencimento", operacao.PrimeiroVencimento)
	query.Set("sistemaAmortizacao", operacao.SistemaAmortizacao)

	resp, err := cj.http.Get(cj.baseURL + "/calendario?" + query.Encode())

	if err != nil {
		return nil, errors.New("não foi possível consultar a api de juros")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("api de juros retornou status " + strconv.Itoa(resp.StatusCode))
	}

	var body struct {
		Parcelas []struct {
			Numero       int     `json:"numero"`
			Vencimento   string  `json:"vencimento"`
			ValorParcela float64 `json:"valorParcela"`
			Amortizacao  float64 `json:"amortizacao"`
			Juros        float64 `json:"juros"`
			SaldoDevedor float64 `json:"saldoDevedor"`
		} `json:"parcelas"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, errors.New("resposta da api de juros em formato inválido")
	}

	if len(body.Parcelas) == 0 {
		return nil, errors.New("api de juros não retornou parcelas")
	}

	novas := []NovaParcela{}

	for _, parcela := range body.Parcelas {
		if parcela.Numero <= 0 {
			return nil, errors.New("api de juros retornou parcela com numero inválido")
		}
		if _, err := time.Parse("2006-01-02", parcela.Vencimento); err != nil {
			return nil, errors.New("api de juros retornou vencimento fora do formato YYYY-MM-DD")
		}

		novas = append(novas, NovaParcela{
			Numero:       parcela.Numero,
			Vencimento:   parcela.Vencimento,
			ValorParcela: parcela.ValorParcela,
			Amortizacao:  parcela.Amortizacao,
			Juros:        parcela.Juros,
			SaldoDevedor: parcela.SaldoDevedor,
		})
	}

	return novas, nil
}
