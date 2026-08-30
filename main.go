package main

import 	"github.com/gin-gonic/gin"

func main() {
	db, err := ConectarBanco()

	if err != nil {
		panic(err)
	}

	defer db.Close()

	repo := NovoRepositorio(db)
	service := NovoService(repo)
	rotas := NovaRotas(service)

	router := gin.Default()

	router.Use(func (c *gin.Context)  {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})

	router.GET("/tarefas", rotas.Listar)
	router.POST("/tarefas", rotas.Criar)
	router.PUT("/tarefas/:id", rotas.Atualizar)
	router.DELETE("/tarefas/:id", rotas.Deletar)

	router.Static("/public", "./public")
	router.StaticFile("/", "./public/index.html")

	router.Run(":8080")
}
