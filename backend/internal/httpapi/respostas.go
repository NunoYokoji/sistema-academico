// Package httpapi expõe a API REST consumida pelo front-end, servida pelo Gin.
package httpapi

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"sistema-academico/backend/internal/validate"
)

// erroJSON é o corpo devolvido em qualquer resposta de erro. O campo
// "campos" permite ao front destacar cada input individualmente.
type erroJSON struct {
	Erro   string         `json:"erro"`
	Campos validate.Erros `json:"campos,omitempty"`
}

// responderErro devolve 4xx/5xx no formato padrão da API.
func responderErro(c *gin.Context, status int, mensagem string) {
	c.AbortWithStatusJSON(status, erroJSON{Erro: mensagem})
}

// responderInvalido devolve 400 com o detalhamento por campo (seção 8).
func responderInvalido(c *gin.Context, erros validate.Erros) {
	c.AbortWithStatusJSON(400, erroJSON{
		Erro:   "Existem campos inválidos no formulário.",
		Campos: erros,
	})
}

// lerJSON decodifica e valida o corpo usando o binding do Gin. Em caso de
// falha já responde ao cliente e devolve false.
func lerJSON(c *gin.Context, destino any) bool {
	err := c.ShouldBindJSON(destino)
	if err == nil {
		return true
	}

	// Regras não atendidas ou campo com tipo errado: erro por campo.
	if erros, porCampo := validate.Traduzir(err); porCampo {
		responderInvalido(c, erros)
		return false
	}

	responderErro(c, 400,
		"Corpo da requisição inválido: envie um JSON com os campos esperados.")
	return false
}

// textoAparado é uma string que perde os espaços das pontas já na
// decodificação. Sem isso "   " passaria no binding:"required" e a
// contagem de min/max incluiria os espaços.
type textoAparado string

func (t *textoAparado) UnmarshalJSON(dados []byte) error {
	var bruto string
	if err := json.Unmarshal(dados, &bruto); err != nil {
		return err
	}
	*t = textoAparado(strings.TrimSpace(bruto))
	return nil
}

func (t textoAparado) String() string { return string(t) }
