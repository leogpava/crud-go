CREATE TABLE financiamento (
    id UUID PRIMARY KEY,
    operacao_id VARCHAR(50) NOT NULL UNIQUE,
    valor_financiado NUMERIC(15,2) NOT NULL,
    quantidade_parcelas INTEGER NOT NULL,
    taxa_juros NUMERIC(10,4) NOT NULL,
    primeiro_vencimento DATE NOT NULL,
    sistema_amortizacao VARCHAR(20) NOT NULL
);

CREATE TABLE parcela (
    id UUID PRIMARY KEY,
    operacao_id VARCHAR(50) NOT NULL REFERENCES financiamento(operacao_id),
    numero INTEGER NOT NULL,
    vencimento DATE NOT NULL,
    valor_parcela NUMERIC(15,2) NOT NULL,
    amortizacao NUMERIC(15,2) NOT NULL,
    juros NUMERIC(15,2) NOT NULL,
    saldo_devedor NUMERIC(15,2) NOT NULL,
    data_pagamento DATE,
    UNIQUE (operacao_id, numero)
);
