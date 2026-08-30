package main

import (
	"database/sql"
)

type Tarefa struct {
	ID 	  int    `json:"id"`
	Texto string `json:"texto"`
	Feita bool   `json:"feita"`
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

func (r *Repositorio) Listar() ([]Tarefa, error) {
	rows, err := r.db.Query("SELECT id, texto, feita FROM tarefas")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tarefas := []Tarefa{}

	for rows.Next() {
		tarefa := Tarefa{}

		err := rows.Scan(
		&tarefa.ID,
        &tarefa.Texto,
        &tarefa.Feita,
		)

		if err != nil {
			return nil, err
		}

		tarefas = append(tarefas, tarefa)
	}

	return tarefas, nil
}

func (r *Repositorio) Criar( texto string ) (Tarefa, error) {

	row := r.db.QueryRow(
		"INSERT INTO tarefas (texto) VALUES ($1) RETURNING id, texto, feita", texto,
	)

	tarefa := Tarefa{}

	err := row.Scan(
			&tarefa.ID,
			&tarefa.Texto,
			&tarefa.Feita,
	)

	if err != nil {
		return Tarefa{}, err
	}

	return tarefa, nil
}

func (r *Repositorio) Atualizar(id int, feita bool) (bool, error) {
	att, err := r.db.Exec("UPDATE tarefas SET feita = $1 WHERE id = $2",
	feita,
	id,
	)

	if err != nil {
		return false, err
	}

	linhas, err := att.RowsAffected()

	if err != nil {
		return false, err
	}

	if linhas == 0 {
		return false, nil
	}

	return true, nil
}

func (r *Repositorio) Deletar(id int) (bool, error) {
	del, err := r.db.Exec("DELETE FROM tarefas WHERE id = $1", id)

	if err != nil {
		return false, err
	}

	linhas, err := del.RowsAffected()

	if err != nil {
			return false, err
	}

	if linhas == 0 {
		return false, nil
	}

	return true, nil
}
