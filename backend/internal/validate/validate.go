// Package validate adapta o validador do Gin (go-playground/validator) às
// regras da seção 8 do enunciado: registra as validações customizadas e
// traduz os erros para mensagens em português, por campo.
//
// As validações rodam no back-end mesmo já existindo no front-end: a
// validação do cliente não é considerada suficiente.
package validate

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RA: apenas dígitos, entre 5 e 8 posições (aceita tanto o 12345 do
// exercício de front quanto o 123456 do exemplo do enunciado final).
var raRegex = regexp.MustCompile(`^\d{5,8}$`)

// Erros mapeia nome do campo -> mensagem, no formato que o front-end usa
// para destacar cada input individualmente.
type Erros map[string]string

// Vazio informa se a validação passou.
func (e Erros) Vazio() bool { return len(e) == 0 }

// mensagens traz o texto de cada combinação (campo, regra). O fallback por
// regra cobre o que não estiver listado explicitamente.
var mensagens = map[string]map[string]string{
	"ra": {
		"required": "O RA é obrigatório.",
		"ra":       "RA inválido: informe de 5 a 8 dígitos.",
	},
	"nome": {
		"required": "O nome é obrigatório.",
		"min":      "O nome deve ter ao menos 3 caracteres.",
		"max":      "O nome deve ter no máximo 120 caracteres.",
	},
	"email": {
		"required": "O e-mail é obrigatório.",
		"email":    "Formato de e-mail inválido (ex.: joao@email.com).",
	},
	"curso": {
		"required": "O curso é obrigatório.",
		"max":      "O curso deve ter no máximo 120 caracteres.",
	},
	"semestre": {
		"required": "O semestre é obrigatório.",
		"min":      "O semestre deve ser um número inteiro entre 1 e 10.",
		"max":      "O semestre deve ser um número inteiro entre 1 e 10.",
		"tipo":     "O semestre deve ser um número inteiro entre 1 e 10.",
	},
	"p1": {
		"required": "A nota da P1 é obrigatória.",
		"gte":      "A nota da P1 deve estar entre 0 e 10.",
		"lte":      "A nota da P1 deve estar entre 0 e 10.",
		"tipo":     "A nota da P1 deve ser um número válido.",
	},
	"p2": {
		"required": "A nota da P2 é obrigatória.",
		"gte":      "A nota da P2 deve estar entre 0 e 10.",
		"lte":      "A nota da P2 deve estar entre 0 e 10.",
		"tipo":     "A nota da P2 deve ser um número válido.",
	},
}

// Registrar configura o validador singleton do Gin. Deve ser chamado uma
// única vez, antes de o servidor começar a atender.
func Registrar() error {
	motor, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("validador do Gin em formato inesperado")
	}

	// Faz o validador reportar o nome vindo da tag json ("ra", "semestre")
	// em vez do nome do campo Go ("RA", "Semestre"). São essas chaves que o
	// front-end procura nos atributos data-erro do HTML.
	motor.RegisterTagNameFunc(func(campo reflect.StructField) string {
		nome := strings.Split(campo.Tag.Get("json"), ",")[0]
		if nome == "-" || nome == "" {
			return campo.Name
		}
		return nome
	})

	// Regra do RA em um lugar só, reaproveitada pela tag binding:"ra".
	return motor.RegisterValidation("ra", func(campo validator.FieldLevel) bool {
		return raRegex.MatchString(campo.Field().String())
	})
}

// Traduzir converte o erro devolvido por ShouldBindJSON em mensagens por
// campo. O segundo retorno indica se o erro pôde ser atribuído a campos
// específicos; quando é false, o corpo da requisição está malformado e o
// chamador deve responder com a mensagem genérica.
func Traduzir(err error) (Erros, bool) {
	// Regras de validação não atendidas.
	var invalidos validator.ValidationErrors
	if errors.As(err, &invalidos) {
		erros := Erros{}
		for _, campo := range invalidos {
			erros[campo.Field()] = mensagem(campo.Field(), campo.Tag())
		}
		return erros, true
	}

	// Campo com o tipo errado no JSON (ex.: "semestre": "quatro"). Aponta o
	// campo específico em vez de recusar o corpo inteiro sem explicação.
	var tipoInvalido *json.UnmarshalTypeError
	if errors.As(err, &tipoInvalido) && tipoInvalido.Field != "" {
		campo := ultimoSegmento(tipoInvalido.Field)
		return Erros{campo: mensagem(campo, "tipo")}, true
	}

	return nil, false
}

func mensagem(campo, regra string) string {
	if porCampo, ok := mensagens[campo]; ok {
		if texto, ok := porCampo[regra]; ok {
			return texto
		}
	}
	if regra == "required" {
		return "Este campo é obrigatório."
	}
	return "Valor inválido para este campo."
}

// ultimoSegmento extrai "p1" de caminhos aninhados como "notas.p1".
func ultimoSegmento(caminho string) string {
	if i := strings.LastIndex(caminho, "."); i >= 0 {
		return caminho[i+1:]
	}
	return caminho
}
