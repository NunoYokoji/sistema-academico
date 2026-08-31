// Package model define as entidades do sistema acadêmico e as regras de
// negócio que derivam da nota: média e situação acadêmica.
package model

import "math"

// Situações acadêmicas possíveis (seção 5 do enunciado).
const (
	SituacaoAprovado  = "Aprovado"
	SituacaoExame     = "Exame"
	SituacaoReprovado = "Reprovado"
)

// Aluno é o cadastro do aluno. Nota é opcional: um aluno recém-cadastrado
// ainda não possui notas lançadas.
type Aluno struct {
	ID       int64  `json:"id"`
	RA       string `json:"ra"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Curso    string `json:"curso"`
	Semestre int    `json:"semestre"`
	Nota     *Nota  `json:"nota"`
}

// Nota guarda as avaliações do aluno na disciplina. Média e situação são
// persistidas (o enunciado as exige como colunas) mas sempre recalculadas
// pelo back-end a partir de P1 e P2 — nunca aceitas do cliente.
//
// As tags db nomeiam as colunas para o goqu, que lê esta struct diretamente
// do banco; sem elas AlunoID seria procurado como "alunoid".
type Nota struct {
	ID       int64   `json:"id"       db:"id"`
	AlunoID  int64   `json:"aluno_id" db:"aluno_id"`
	P1       float64 `json:"p1"       db:"p1"`
	P2       float64 `json:"p2"       db:"p2"`
	Media    float64 `json:"media"    db:"media"`
	Situacao string  `json:"situacao" db:"situacao"`
}

// CalcularMedia devolve (P1 + P2) / 2 arredondada em duas casas, que é a
// precisão da coluna numeric(4,2) do banco. Arredondar antes de decidir a
// situação garante que a média exibida seja exatamente a média julgada.
func CalcularMedia(p1, p2 float64) float64 {
	return math.Round(((p1+p2)/2)*100) / 100
}

// DefinirSituacao aplica as faixas do enunciado: >= 6 aprovado,
// >= 4 e < 6 exame, < 4 reprovado.
func DefinirSituacao(media float64) string {
	switch {
	case media >= 6:
		return SituacaoAprovado
	case media >= 4:
		return SituacaoExame
	default:
		return SituacaoReprovado
	}
}

// Aplicar recalcula média e situação a partir das notas atuais.
func (n *Nota) Aplicar() {
	n.Media = CalcularMedia(n.P1, n.P2)
	n.Situacao = DefinirSituacao(n.Media)
}
