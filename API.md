# FinOps — API de operações de crédito

Módulo de FinOps da esteira de crédito. Guarda o cadastro da operação, o calendário de parcelas e a situação de pagamento de cada uma (inclusive atraso).

- **Base URL:** `http://localhost:8080`
- **Content-Type:** `application/json` (UTF-8)
- **CORS:** liberado para qualquer origem (`GET`, `POST`)
- **Autenticação:** nenhuma nesta etapa

Datas sempre trafegam como texto no formato `YYYY-MM-DD`. Valores monetários são números com 2 casas decimais (não vêm como string). A taxa de juros é mensal em percentual — `1.5` significa 1,5% a.m.

## Formato de erro

Toda falha devolve o mesmo envelope, com mensagem em português pronta para exibir:

```json
{ "erro": "operação não encontrada" }
```

Códigos usados: `200` OK, `201` criado, `400` qualquer erro de validação ou de regra de negócio (inclusive recurso inexistente), `500` falha inesperada ao listar.

---

## 1. Cadastrar operação

```
POST /operacoes
```

Registra o cabeçalho da operação de crédito. O `operacaoID` é o identificador de negócio e é único — é por ele que todos os outros endpoints são chamados.

**Corpo**

| Campo | Tipo | Regra |
|---|---|---|
| `operacaoID` | string | obrigatório, único |
| `valorFinanciado` | number | maior que zero |
| `quantidadeParcelas` | integer | maior que zero |
| `taxaJuros` | number | maior ou igual a zero, % ao mês |
| `primeiroVencimento` | string | `YYYY-MM-DD` |
| `sistemaAmortizacao` | string | `PRICE` ou `SAC` (aceita minúsculo) |

```json
{
    "operacaoID" : "OP-0001",
    "valorFinanciado" : 10000.00,
    "quantidadeParcelas" : 12,
    "taxaJuros" : 1.5,
    "primeiroVencimento" : "2026-10-20",
    "sistemaAmortizacao" : "PRICE"
}
```

**Resposta `201`** — a operação criada, com o `id` (UUID) gerado pelo banco.

```json
{
    "id" : "550e8400-e29b-41d4-a716-446655440000",
    "operacaoID" : "OP-0001",
    "valorFinanciado" : 10000,
    "quantidadeParcelas" : 12,
    "taxaJuros" : 1.5,
    "primeiroVencimento" : "2026-10-20",
    "sistemaAmortizacao" : "PRICE"
}
```

**Erros `400`:** `corpo da requisição inválido`, `operacaoID não pode ser vazio`, `valorFinanciado deve ser maior que zero`, `quantidadeParcelas deve ser maior que zero`, `taxaJuros deve ser maior ou igual a zero`, `primeiroVencimento não pode ser vazio`, `primeiroVencimento deve estar no formato YYYY-MM-DD`, `sistemaAmortizacao deve ser PRICE ou SAC`.

---

## 2. Listar operações

```
GET /operacoes
```

Devolve `200` com um array de operações (mesmo objeto do item 1), ordenado por `primeiroVencimento` e `operacaoID`. Sem operações cadastradas, devolve `[]`.

---

## 3. Importar o calendário de parcelas

```
POST /operacoes/{operacaoID}/parcelas/importar
```

**Este é o ponto de integração com o grupo de Juros.** Não tem corpo. Ao ser chamado, o FinOps consulta a API de Juros (ver [Dependência externa](#dependência-externa-api-do-grupo-de-juros)), valida o calendário recebido e persiste as parcelas em uma única transação — ou grava todas, ou nenhuma.

**Resposta `201`** — o array de parcelas gravadas, no mesmo formato do item 4.

**Erros `400`**

| Mensagem | Quando |
|---|---|
| `operação não encontrada` | `operacaoID` não cadastrado |
| `operação já possui parcelas importadas` | import é uma única vez por operação |
| `calendário recebido não bate com a quantidadeParcelas da operação` | a API de Juros devolveu um número de parcelas diferente do cadastrado |
| `não foi possível consultar a api de juros` | timeout (10s) ou API fora do ar |
| `api de juros retornou status 500` | a API de Juros respondeu diferente de `200` |
| `resposta da api de juros em formato inválido` | JSON malformado |
| `api de juros não retornou parcelas` | lista vazia |
| `api de juros retornou parcela com numero inválido` | `numero` menor ou igual a zero |
| `api de juros retornou vencimento fora do formato YYYY-MM-DD` | data inválida |
| `JUROS_API_URL não configurada` | variável de ambiente ausente no FinOps |

---

## 4. Listar parcelas

```
GET /operacoes/{operacaoID}/parcelas
```

Devolve `200` com as parcelas ordenadas por `numero`. Antes do import, devolve `[]`.

```json
[
    {
        "id" : "9f1c2b7a-3d4e-4f5a-8b6c-7d8e9f0a1b2c",
        "operacaoID" : "OP-0001",
        "numero" : 1,
        "vencimento" : "2026-10-20",
        "valorParcela" : 916.8,
        "amortizacao" : 766.8,
        "juros" : 150,
        "saldoDevedor" : 9233.2,
        "dataPagamento" : "2026-10-20",
        "status" : "PAGA",
        "diasAtraso" : 0
    },
    {
        "id" : "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
        "operacaoID" : "OP-0001",
        "numero" : 2,
        "vencimento" : "2026-11-20",
        "valorParcela" : 916.8,
        "amortizacao" : 778.3,
        "juros" : 138.5,
        "saldoDevedor" : 8454.9,
        "dataPagamento" : "",
        "status" : "ATRASADA",
        "diasAtraso" : 47
    }
]
```

### Situação da parcela

`status` e `diasAtraso` **não são armazenados** — são calculados a cada leitura, comparando o vencimento com a data de hoje. Não é preciso rodar nenhuma rotina para a parcela "virar" inadimplente.

| `status` | Significado |
|---|---|
| `PAGA` | tem `dataPagamento` preenchida |
| `ATRASADA` | em aberto e o vencimento já passou; `diasAtraso` traz os dias corridos de atraso |
| `EM_ABERTO` | em aberto e ainda dentro do prazo (`diasAtraso` é `0`) |

`dataPagamento` vem como string vazia (`""`) enquanto a parcela não foi paga — nunca `null`.

**Erro `400`:** `operação não encontrada`.

---

## 5. Dar baixa em uma parcela

```
POST /operacoes/{operacaoID}/parcelas/{numero}/pagamento
```

Registra o pagamento. Só existe baixa total — não há pagamento parcial nesta etapa.

**Corpo (opcional).** Sem corpo, a baixa é registrada com a data de hoje.

```json
{ "dataPagamento" : "2026-10-20" }
```

**Resposta `200`** — a parcela atualizada, já com `status: "PAGA"` e `diasAtraso: 0`.

**Erros `400`:** `numero da parcela inválido` (não é número), `numero da parcela deve ser maior que zero`, `dataPagamento deve estar no formato YYYY-MM-DD`, `operação não encontrada`, `parcela não encontrada ou já está paga`.

A baixa não é reversível e não sobrescreve: chamar duas vezes na mesma parcela devolve `parcela não encontrada ou já está paga`, preservando o primeiro pagamento.

---

## Dependência externa: API do grupo de Juros

Para o endpoint de import funcionar, o FinOps precisa da variável `JUROS_API_URL` apontando para a API de Juros, que deve expor:

```
GET {JUROS_API_URL}/calendario
```

**Query params enviados por nós** (derivados da operação cadastrada):

| Param | Exemplo |
|---|---|
| `valorFinanciado` | `10000.00` |
| `quantidadeParcelas` | `12` |
| `taxaJuros` | `1.5000` |
| `primeiroVencimento` | `2026-10-20` |
| `sistemaAmortizacao` | `PRICE` |

**Resposta esperada — `200` com:**

```json
{
    "parcelas" : [
        { "numero" : 1, "vencimento" : "2026-10-20", "valorParcela" : 916.80, "amortizacao" : 766.80, "juros" : 150.00, "saldoDevedor" : 9233.20 },
        { "numero" : 2, "vencimento" : "2026-11-20", "valorParcela" : 916.80, "amortizacao" : 778.30, "juros" : 138.50, "saldoDevedor" : 8454.90 }
    ]
}
```

Regras que validamos na entrada: a lista não pode ser vazia, `numero` tem que ser maior que zero, `vencimento` tem que estar em `YYYY-MM-DD`, e a quantidade de itens tem que ser igual ao `quantidadeParcelas` da operação. Timeout de 10 segundos. Campos extras no JSON são ignorados.

Exemplo completo de payload em [examples/calendarioParcelasPayload.json](examples/calendarioParcelasPayload.json).

---

## Fluxo de integração ponta a ponta

```
1. POST /operacoes                                   → cadastra OP-0001
2. POST /operacoes/OP-0001/parcelas/importar         → FinOps chama a API de Juros e grava as 12 parcelas
3. GET  /operacoes/OP-0001/parcelas                  → acompanha status e dias de atraso
4. POST /operacoes/OP-0001/parcelas/1/pagamento      → baixa a parcela 1
```

### curl

```bash
curl -X POST http://localhost:8080/operacoes \
  -H 'Content-Type: application/json' \
  -d @examples/inboundDecisionPayload.json

curl -X POST http://localhost:8080/operacoes/OP-0001/parcelas/importar

curl http://localhost:8080/operacoes/OP-0001/parcelas

curl -X POST http://localhost:8080/operacoes/OP-0001/parcelas/1/pagamento \
  -H 'Content-Type: application/json' \
  -d '{"dataPagamento":"2026-10-20"}'
```

---

## Rodando local

```bash
# 1. Postgres com as tabelas
psql $DATABASE_URL -f examples/table.sql

# 2. .env na raiz
DATABASE_URL=postgres://usuario:senha@localhost:5432/finops?sslmode=disable
JUROS_API_URL=http://localhost:8081

# 3. sobe na :8080 (front de demonstração em http://localhost:8080/)
go run .
```

## Fora do escopo desta versão

Carteira consolidada / aging por faixa de atraso, multa e juros de mora, PDD, dados do cliente, pagamento parcial, estorno de baixa e autenticação.
