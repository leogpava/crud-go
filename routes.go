package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Rotas struct {
	service *Service
}

func NovaRotas(service *Service) *Rotas {
	return &Rotas{service: service}
}

// GET /operacoes - lista todas as operacoes.

func (rt *Rotas) Listar(c *gin.Context) {
	operacoes, err := rt.service.ListarOperacoes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(http.StatusOK, operacoes)
}

// POST /operacoes - cria uma nova operacao.
func (rt *Rotas) Criar(c *gin.Context) {
	var body struct {
		OperacaoID         string  `json:"operacaoID"`
		ValorFinanciado    float64 `json:"valorFinanciado"`
		QuantidadeParcelas int     `json:"quantidadeParcelas"`
		TaxaJuros          float64 `json:"taxaJuros"`
		PrimeiroVencimento string  `json:"primeiroVencimento"`
		SistemaAmortizacao string  `json:"sistemaAmortizacao"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "corpo da requisição inválido"})
		return
	}

	operacao, err := rt.service.CriarOperacao(NovaOperacao{
		OperacaoID:         body.OperacaoID,
		ValorFinanciado:    body.ValorFinanciado,
		QuantidadeParcelas: body.QuantidadeParcelas,
		TaxaJuros:          body.TaxaJuros,
		PrimeiroVencimento: body.PrimeiroVencimento,
		SistemaAmortizacao: body.SistemaAmortizacao,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, operacao)
}

// POST /operacoes/:operacaoID/parcelas/importar - busca o calendario no grupo de juros.
func (rt *Rotas) ImportarParcelas(c *gin.Context) {
	parcelas, err := rt.service.ImportarParcelas(c.Param("operacaoID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, parcelas)
}

// GET /operacoes/:operacaoID/parcelas - lista as parcelas com a situacao de cada uma.
func (rt *Rotas) ListarParcelas(c *gin.Context) {
	parcelas, err := rt.service.ListarParcelas(c.Param("operacaoID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(http.StatusOK, parcelas)
}

// POST /operacoes/:operacaoID/parcelas/:numero/pagamento - da baixa na parcela.
func (rt *Rotas) PagarParcela(c *gin.Context) {
	var body struct {
		DataPagamento string `json:"dataPagamento"`
	}

	// o corpo é opcional, sem ele a baixa é com a data de hoje.
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "corpo da requisição inválido"})
			return
		}
	}

	numero, err := strconv.Atoi(c.Param("numero"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "numero da parcela inválido"})
		return
	}

	parcela, err := rt.service.PagarParcela(c.Param("operacaoID"), numero, body.DataPagamento)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, parcela)
}
