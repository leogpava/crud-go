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

//GET /tarefas - lista todas

func (rt *Rotas) Listar(c *gin.Context) {
	tarefas, err := rt.service.ListarTarefas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tarefas)
}

// POST /tarefas - cria uma nova tarefa.
func (rt *Rotas) Criar(c *gin.Context) {
	var body struct {
		Texto string `json:"texto"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "corpo da requsicao invalido."})
		return
	}

	tarefa, err := rt.service.CriarTarefa(body.Texto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tarefa)
}

//PUT /tarefas/:id - atualiza a tarefa.

func (rt *Rotas) Atualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id invalido"})
		return
	}

	var body struct {
		Feita bool `json:"feita"`
	}

	c.ShouldBindJSON(&body)

	if err := rt.service.AtualizarTarefa(id, body.Feita); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "atualizado"})
}

//DELETE /tarefas/:id - remova uma tarefa

func (rt *Rotas) Deletar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id invalido"})
		return
	}

	if err := rt.service.DeletarTarefa(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removido"})
}
