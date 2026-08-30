package main

import "errors"

type Service struct {
	repo *Repositorio
}

func NovoService(repo *Repositorio) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListarTarefas() ([]Tarefa, error) {
	return s.repo.Listar()
}

func (s *Service) CriarTarefa(texto string) (Tarefa, error) {
	if texto == "" {
		return Tarefa{}, errors.New("texto da tarefa não pode ser vazio.")
	}
	return s.repo.Criar(texto)
}

func (s *Service) AtualizarTarefa(id int, feita bool) error {
	ok, err := s.repo.Atualizar(id, feita)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New("tarefa não encontrada")
	}

	return nil
}

func (s *Service) DeletarTarefa(id int) error {
	ok, err := s.repo.Deletar(id)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New("tarefa não encontrada")
	}

	return nil
}
