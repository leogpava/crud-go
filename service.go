package main

import (
	"errors"
	"strings"
	"time"
)

type Service struct {
	repo  *Repositorio
	juros *ClienteJuros
}

func NovoService(repo *Repositorio, juros *ClienteJuros) *Service {
	return &Service{repo: repo, juros: juros}
}

func (s *Service) ListarOperacoes() ([]Operacao, error) {
	return s.repo.ListarOperacoes()
}

func (s *Service) CriarOperacao(nova NovaOperacao) (Operacao, error) {
	nova.OperacaoID = strings.TrimSpace(nova.OperacaoID)
	nova.PrimeiroVencimento = strings.TrimSpace(nova.PrimeiroVencimento)
	nova.SistemaAmortizacao = strings.ToUpper(strings.TrimSpace(nova.SistemaAmortizacao))

	if nova.OperacaoID == "" {
		return Operacao{}, errors.New("operacaoID não pode ser vazio")
	}
	if nova.ValorFinanciado <= 0 {
		return Operacao{}, errors.New("valorFinanciado deve ser maior que zero")
	}
	if nova.QuantidadeParcelas <= 0 {
		return Operacao{}, errors.New("quantidadeParcelas deve ser maior que zero")
	}
	if nova.TaxaJuros < 0 {
		return Operacao{}, errors.New("taxaJuros deve ser maior ou igual a zero")
	}
	if nova.PrimeiroVencimento == "" {
		return Operacao{}, errors.New("primeiroVencimento não pode ser vazio")
	}
	if _, err := time.Parse("2006-01-02", nova.PrimeiroVencimento); err != nil {
		return Operacao{}, errors.New("primeiroVencimento deve estar no formato YYYY-MM-DD")
	}
	if nova.SistemaAmortizacao != "PRICE" && nova.SistemaAmortizacao != "SAC" {
		return Operacao{}, errors.New("sistemaAmortizacao deve ser PRICE ou SAC")
	}

	return s.repo.CriarOperacao(nova)

}

// busca o calendario no grupo de juros e grava as parcelas da operacao.
func (s *Service) ImportarParcelas(operacaoID string) ([]Parcela, error) {
	operacaoID = strings.TrimSpace(operacaoID)

	if operacaoID == "" {
		return nil, errors.New("operacaoID não pode ser vazio")
	}

	operacao, err := s.repo.BuscarOperacao(operacaoID)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.ContarParcelas(operacaoID)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("operação já possui parcelas importadas")
	}

	novas, err := s.juros.BuscarCalendario(operacao)
	if err != nil {
		return nil, err
	}

	// nao confia no calendario de fora, tem que bater com o que foi cadastrado aqui.
	if len(novas) != operacao.QuantidadeParcelas {
		return nil, errors.New("calendário recebido não bate com a quantidadeParcelas da operação")
	}

	if err := s.repo.CriarParcelas(operacaoID, novas); err != nil {
		return nil, err
	}

	return s.repo.ListarParcelas(operacaoID)
}

func (s *Service) ListarParcelas(operacaoID string) ([]Parcela, error) {
	operacaoID = strings.TrimSpace(operacaoID)

	if operacaoID == "" {
		return nil, errors.New("operacaoID não pode ser vazio")
	}

	if _, err := s.repo.BuscarOperacao(operacaoID); err != nil {
		return nil, err
	}

	return s.repo.ListarParcelas(operacaoID)
}

func (s *Service) PagarParcela(operacaoID string, numero int, dataPagamento string) (Parcela, error) {
	operacaoID = strings.TrimSpace(operacaoID)
	dataPagamento = strings.TrimSpace(dataPagamento)

	if operacaoID == "" {
		return Parcela{}, errors.New("operacaoID não pode ser vazio")
	}
	if numero <= 0 {
		return Parcela{}, errors.New("numero da parcela deve ser maior que zero")
	}

	// sem data no corpo, a baixa é de hoje.
	if dataPagamento == "" {
		dataPagamento = time.Now().Format("2006-01-02")
	}

	if _, err := time.Parse("2006-01-02", dataPagamento); err != nil {
		return Parcela{}, errors.New("dataPagamento deve estar no formato YYYY-MM-DD")
	}

	if _, err := s.repo.BuscarOperacao(operacaoID); err != nil {
		return Parcela{}, err
	}

	return s.repo.PagarParcela(operacaoID, numero, dataPagamento)
}
