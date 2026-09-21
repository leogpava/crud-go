package main

import (
	"database/sql"
	"errors"
	"time"
)

type Parcela struct {
	ID            string  `json:"id"`
	OperacaoID    string  `json:"operacaoID"`
	Numero        int     `json:"numero"`
	Vencimento    string  `json:"vencimento"`
	ValorParcela  float64 `json:"valorParcela"`
	Amortizacao   float64 `json:"amortizacao"`
	Juros         float64 `json:"juros"`
	SaldoDevedor  float64 `json:"saldoDevedor"`
	DataPagamento string  `json:"dataPagamento"`
	Status        string  `json:"status"`
	DiasAtraso    int     `json:"diasAtraso"`
}

type NovaParcela struct {
	Numero       int
	Vencimento   string
	ValorParcela float64
	Amortizacao  float64
	Juros        float64
	SaldoDevedor float64
}

// o status nao fica salvo no banco, ele é sempre derivado da data de hoje.
func calcularSituacao(parcela *Parcela) {
	if parcela.DataPagamento != "" {
		parcela.Status = "PAGA"
		return
	}

	vencimento, err := time.Parse("2006-01-02", parcela.Vencimento)
	if err != nil {
		return
	}

	agora := time.Now()
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, time.UTC)

	if hoje.After(vencimento) {
		parcela.Status = "ATRASADA"
		parcela.DiasAtraso = int(hoje.Sub(vencimento).Hours() / 24)
		return
	}

	parcela.Status = "EM_ABERTO"
}

func (r *Repositorio) BuscarOperacao(operacaoID string) (Operacao, error) {
	row := r.db.QueryRow(`
		SELECT
			id::text,
			operacao_id,
			valor_financiado,
			quantidade_parcelas,
			taxa_juros,
			to_char(primeiro_vencimento, 'YYYY-MM-DD'),
			sistema_amortizacao
		FROM financiamento
		WHERE operacao_id = $1
	`, operacaoID)

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

	if err == sql.ErrNoRows {
		return Operacao{}, errors.New("operação não encontrada")
	}

	if err != nil {
		return Operacao{}, err
	}

	return operacao, nil
}

func (r *Repositorio) ContarParcelas(operacaoID string) (int, error) {
	row := r.db.QueryRow(`
		SELECT count(*)
		FROM parcela
		WHERE operacao_id = $1
	`, operacaoID)

	total := 0

	err := row.Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}

// grava o calendario inteiro de uma vez só, ou nenhuma parcela.
func (r *Repositorio) CriarParcelas(operacaoID string, novas []NovaParcela) error {
	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, nova := range novas {
		_, err := tx.Exec(`
			INSERT INTO parcela (
				id,
				operacao_id,
				numero,
				vencimento,
				valor_parcela,
				amortizacao,
				juros,
				saldo_devedor
			)
			VALUES (
				gen_random_uuid(),
				$1,
				$2,
				$3::date,
				$4,
				$5,
				$6,
				$7
			)
		`,
			operacaoID,
			nova.Numero,
			nova.Vencimento,
			nova.ValorParcela,
			nova.Amortizacao,
			nova.Juros,
			nova.SaldoDevedor,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repositorio) ListarParcelas(operacaoID string) ([]Parcela, error) {
	rows, err := r.db.Query(`
		SELECT
			id::text,
			operacao_id,
			numero,
			to_char(vencimento, 'YYYY-MM-DD'),
			valor_parcela,
			amortizacao,
			juros,
			saldo_devedor,
			to_char(data_pagamento, 'YYYY-MM-DD')
		FROM parcela
		WHERE operacao_id = $1
		ORDER BY numero
	`, operacaoID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	parcelas := []Parcela{}

	for rows.Next() {
		parcela := Parcela{}
		dataPagamento := sql.NullString{}

		err := rows.Scan(
			&parcela.ID,
			&parcela.OperacaoID,
			&parcela.Numero,
			&parcela.Vencimento,
			&parcela.ValorParcela,
			&parcela.Amortizacao,
			&parcela.Juros,
			&parcela.SaldoDevedor,
			&dataPagamento,
		)

		if err != nil {
			return nil, err
		}

		if dataPagamento.Valid {
			parcela.DataPagamento = dataPagamento.String
		}

		calcularSituacao(&parcela)

		parcelas = append(parcelas, parcela)
	}

	return parcelas, nil
}

func (r *Repositorio) PagarParcela(operacaoID string, numero int, dataPagamento string) (Parcela, error) {

	row := r.db.QueryRow(`
		UPDATE parcela
		SET data_pagamento = $3::date
		WHERE operacao_id = $1
			AND numero = $2
			AND data_pagamento IS NULL
		RETURNING
			id::text,
			operacao_id,
			numero,
			to_char(vencimento, 'YYYY-MM-DD'),
			valor_parcela,
			amortizacao,
			juros,
			saldo_devedor,
			to_char(data_pagamento, 'YYYY-MM-DD')
	`,
		operacaoID,
		numero,
		dataPagamento,
	)

	parcela := Parcela{}
	pagamento := sql.NullString{}

	err := row.Scan(
		&parcela.ID,
		&parcela.OperacaoID,
		&parcela.Numero,
		&parcela.Vencimento,
		&parcela.ValorParcela,
		&parcela.Amortizacao,
		&parcela.Juros,
		&parcela.SaldoDevedor,
		&pagamento,
	)

	if err == sql.ErrNoRows {
		return Parcela{}, errors.New("parcela não encontrada ou já está paga")
	}

	if err != nil {
		return Parcela{}, err
	}

	if pagamento.Valid {
		parcela.DataPagamento = pagamento.String
	}

	calcularSituacao(&parcela)

	return parcela, nil
}
