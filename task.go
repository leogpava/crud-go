package main

import (
	"database/sql"
)

type Operacao struct {
	ID                 string  `json:"id"`
	OperacaoID         string  `json:"operacaoID"`
	ValorFinanciado    float64 `json:"valorFinanciado"`
	QuantidadeParcelas int     `json:"quantidadeParcelas"`
	TaxaJuros          float64 `json:"taxaJuros"`
	PrimeiroVencimento string  `json:"primeiroVencimento"`
	SistemaAmortizacao string  `json:"sistemaAmortizacao"`
}

type NovaOperacao struct {
	OperacaoID         string
	ValorFinanciado    float64
	QuantidadeParcelas int
	TaxaJuros          float64
	PrimeiroVencimento string
	SistemaAmortizacao string
}


type Repositorio struct {
	db *sql.DB
}

// retorna um ponteiro para Repositorio, ou seja, o endereço onde ele está na memória.
func NovoRepositorio(db *sql.DB) *Repositorio {
	//aqui o & pede o endereco do Repositorio. Cria e pega o endereco. Ta criando o ponteiro.
	// cria um Repositorio e retorna um ponteiro para ele.
	return &Repositorio {
		db: db,
	}
}

func (r *Repositorio) ListarOperacoes() ([]Operacao, error) {
	rows, err := r.db.Query(`
		SELECT
			id::text,
			operacao_id,
			valor_financiado,
			quantidade_parcelas,
			taxa_juros,
			to_char(primeiro_vencimento, 'YYYY-MM-DD'),
			sistema_amortizacao
		FROM financiamento
		ORDER BY primeiro_vencimento, operacao_id
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	operacoes := []Operacao{}

	for rows.Next() {
		operacao := Operacao{}

		err := rows.Scan(
			&operacao.ID,
			&operacao.OperacaoID,
			&operacao.ValorFinanciado,
			&operacao.QuantidadeParcelas,
			&operacao.TaxaJuros,
			&operacao.PrimeiroVencimento,
			&operacao.SistemaAmortizacao,
		)

		if err != nil {
			return nil, err
		}

		operacoes = append(operacoes, operacao)
	}

	return operacoes, nil
}

func (r *Repositorio) CriarOperacao(nova NovaOperacao) (Operacao, error) {

	row := r.db.QueryRow(`
		INSERT INTO financiamento (
			id,
			operacao_id,
			valor_financiado,
			quantidade_parcelas,
			taxa_juros,
			primeiro_vencimento,
			sistema_amortizacao
		)
		VALUES (
			gen_random_uuid(),
			$1,
			$2,
			$3,
			$4,
			$5::date,
			$6
		)
		RETURNING
			id::text,
			operacao_id,
			valor_financiado,
			quantidade_parcelas,
			taxa_juros,
			to_char(primeiro_vencimento, 'YYYY-MM-DD'),
			sistema_amortizacao
	`,
		nova.OperacaoID,
		nova.ValorFinanciado,
		nova.QuantidadeParcelas,
		nova.TaxaJuros,
		nova.PrimeiroVencimento,
		nova.SistemaAmortizacao,
	)

	operacao := Operacao{}

	err := row.Scan(
			&operacao.ID,
			&operacao.OperacaoID,
			&operacao.ValorFinanciado,
			&operacao.QuantidadeParcelas,
			&operacao.TaxaJuros,
			&operacao.PrimeiroVencimento,
			&operacao.SistemaAmortizacao,
	)

	if err != nil {
		return Operacao{}, err
	}

	return operacao, nil
}
