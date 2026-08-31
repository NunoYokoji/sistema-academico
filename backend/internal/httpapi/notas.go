package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"sistema-academico/backend/internal/model"
	"sistema-academico/backend/internal/store"
)

// Ponteiros permitem distinguir "campo ausente" de "campo igual a zero":
// required em ponteiro reprova apenas o nulo, então a nota 0 é válida.
type notaRequisicao struct {
	P1 *float64 `json:"p1" binding:"required,gte=0,lte=10"`
	P2 *float64 `json:"p2" binding:"required,gte=0,lte=10"`
}

// GET /api/alunos/:id/notas
func (a *API) buscarNotasDoAluno(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	nota, err := a.store.BuscarNotaDoAluno(c.Request.Context(), id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		responderErro(c, http.StatusNotFound, "Este aluno ainda não possui notas lançadas.")
		return
	}
	if err != nil {
		a.erroInterno(c, "buscando notas do aluno", err)
		return
	}
	c.JSON(http.StatusOK, nota)
}

// POST /api/alunos/:id/notas — lança as notas de um aluno.
func (a *API) criarNota(c *gin.Context) {
	alunoID, ok := lerID(c)
	if !ok {
		return
	}

	nota, ok := lerNotaValidada(c)
	if !ok {
		return
	}

	existe, err := a.store.AlunoExiste(c.Request.Context(), alunoID)
	if err != nil {
		a.erroInterno(c, "verificando aluno", err)
		return
	}
	if !existe {
		responderErro(c, http.StatusNotFound, "Aluno não encontrado.")
		return
	}

	nota.AlunoID = alunoID
	switch err := a.store.CriarNota(c.Request.Context(), &nota); {
	case errors.Is(err, store.ErrNotaExistente):
		responderErro(c, http.StatusConflict,
			"Este aluno já possui notas lançadas; utilize PUT /api/notas/{id} para alterá-las.")
		return
	case err != nil:
		a.erroInterno(c, "criando nota", err)
		return
	}

	a.invalidarAluno(c, alunoID)
	c.JSON(http.StatusCreated, nota)
}

// PUT /api/notas/:id — altera as notas já lançadas.
func (a *API) atualizarNota(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	nova, ok := lerNotaValidada(c)
	if !ok {
		return
	}

	atual, err := a.store.BuscarNota(c.Request.Context(), id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		responderErro(c, http.StatusNotFound, "Lançamento de notas não encontrado.")
		return
	}
	if err != nil {
		a.erroInterno(c, "buscando nota", err)
		return
	}

	nova.ID = atual.ID
	nova.AlunoID = atual.AlunoID

	if err := a.store.AtualizarNota(c.Request.Context(), &nova); err != nil {
		a.erroInterno(c, "atualizando nota", err)
		return
	}

	a.invalidarAluno(c, nova.AlunoID)
	c.JSON(http.StatusOK, nova)
}

// DELETE /api/notas/:id
func (a *API) excluirNota(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	// Lê antes de excluir para saber qual chave de aluno invalidar.
	nota, err := a.store.BuscarNota(c.Request.Context(), id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		responderErro(c, http.StatusNotFound, "Lançamento de notas não encontrado.")
		return
	}
	if err != nil {
		a.erroInterno(c, "buscando nota", err)
		return
	}

	if err := a.store.ExcluirNota(c.Request.Context(), id); err != nil {
		a.erroInterno(c, "excluindo nota", err)
		return
	}

	a.invalidarAluno(c, nota.AlunoID)
	c.Status(http.StatusNoContent)
}

// lerNotaValidada decodifica e valida P1/P2, já calculando média e situação
// no servidor — os dois nunca são aceitos do cliente.
func lerNotaValidada(c *gin.Context) (model.Nota, bool) {
	var req notaRequisicao
	if !lerJSON(c, &req) {
		return model.Nota{}, false
	}

	nota := model.Nota{P1: *req.P1, P2: *req.P2}
	nota.Aplicar()
	return nota, true
}
