package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := ConectarBanco()

	if err != nil {
		panic(err)
	}

	defer db.Close()

	repo := NovoRepositorio(db)
	juros := NovoClienteJuros(os.Getenv("JUROS_API_URL"))
	service := NovoService(repo, juros)
	rotas := NovaRotas(service)

	router := gin.Default()

	router.Use(func (c *gin.Context)  {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})

	router.GET("/operacoes", rotas.Listar)
	router.POST("/operacoes", rotas.Criar)

	router.POST("/operacoes/:operacaoID/parcelas/importar", rotas.ImportarParcelas)
	router.GET("/operacoes/:operacaoID/parcelas", rotas.ListarParcelas)
	router.POST("/operacoes/:operacaoID/parcelas/:numero/pagamento", rotas.PagarParcela)

	router.Static("/public", "./public")
	router.StaticFile("/", "./public/index.html")

	router.Run(":8080")
}
