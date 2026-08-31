package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sistema-academico/backend/internal/cache"
	"sistema-academico/backend/internal/model"
	"sistema-academico/backend/internal/store"
	"sistema-academico/backend/internal/validate"
)

// As tags binding declaram as regras da seção 8; as mensagens em português
// vêm do pacote validate. Semestre é ponteiro para que "ausente" e "zero"
// sejam distinguíveis: required em ponteiro só reprova o valor nulo.
type alunoRequisicao struct {
	RA       textoAparado `json:"ra"       binding:"required,ra"`
	Nome     textoAparado `json:"nome"     binding:"required,min=3,max=120"`
	Email    textoAparado `json:"email"    binding:"required,email,max=160"`
	Curso    textoAparado `json:"curso"    binding:"required,max=120"`
	Semestre *int         `json:"semestre" binding:"required,min=1,max=10"`
}

func (r alunoRequisicao) paraModelo() model.Aluno {
	return model.Aluno{
		RA:       r.RA.String(),
		Nome:     r.Nome.String(),
		Email:    r.Email.String(),
		Curso:    r.Curso.String(),
		Semestre: *r.Semestre,
	}
}

// GET /api/alunos — consulta com cache (seção 11).
func (a *API) listarAlunos(c *gin.Context) {
	var alunos []model.Aluno

	if a.cache.Buscar(c.Request.Context(), cache.ChaveListaAlunos, &alunos) {
		marcarCache(c, "HIT")
		c.JSON(http.StatusOK, alunos)
		return
	}
	marcarCache(c, "MISS")

	alunos, err := a.store.ListarAlunos(c.Request.Context())
	if err != nil {
		a.erroInterno(c, "listando alunos", err)
		return
	}

	a.cache.Guardar(c.Request.Context(), cache.ChaveListaAlunos, alunos)
	c.JSON(http.StatusOK, alunos)
}

// GET /api/alunos/:id — segunda consulta com cache (seção 11).
func (a *API) buscarAluno(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	var aluno model.Aluno
	if a.cache.Buscar(c.Request.Context(), cache.ChaveAluno(id), &aluno) {
		marcarCache(c, "HIT")
		c.JSON(http.StatusOK, aluno)
		return
	}
	marcarCache(c, "MISS")

	aluno, err := a.store.BuscarAluno(c.Request.Context(), id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		responderErro(c, http.StatusNotFound, "Aluno não encontrado.")
		return
	}
	if err != nil {
		a.erroInterno(c, "buscando aluno", err)
		return
	}

	a.cache.Guardar(c.Request.Context(), cache.ChaveAluno(id), aluno)
	c.JSON(http.StatusOK, aluno)
}

// POST /api/alunos
func (a *API) criarAluno(c *gin.Context) {
	var req alunoRequisicao
	if !lerJSON(c, &req) {
		return
	}

	aluno := req.paraModelo()

	switch err := a.store.CriarAluno(c.Request.Context(), &aluno); {
	case errors.Is(err, store.ErrRADuplicado):
		responderConflitoRA(c)
		return
	case err != nil:
		a.erroInterno(c, "criando aluno", err)
		return
	}

	a.cache.Invalidar(c.Request.Context(), cache.ChaveListaAlunos)
	c.JSON(http.StatusCreated, aluno)
}

// PUT /api/alunos/:id
func (a *API) atualizarAluno(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	var req alunoRequisicao
	if !lerJSON(c, &req) {
		return
	}

	aluno := req.paraModelo()
	aluno.ID = id

	switch err := a.store.AtualizarAluno(c.Request.Context(), &aluno); {
	case errors.Is(err, store.ErrNaoEncontrado):
		responderErro(c, http.StatusNotFound, "Aluno não encontrado.")
		return
	case errors.Is(err, store.ErrRADuplicado):
		responderConflitoRA(c)
		return
	case err != nil:
		a.erroInterno(c, "atualizando aluno", err)
		return
	}

	// Devolve o aluno completo: o cadastro mudou, mas as notas continuam.
	atualizado, err := a.store.BuscarAluno(c.Request.Context(), id)
	if err != nil {
		a.erroInterno(c, "relendo aluno", err)
		return
	}

	a.invalidarAluno(c, id)
	c.JSON(http.StatusOK, atualizado)
}

// DELETE /api/alunos/:id
func (a *API) excluirAluno(c *gin.Context) {
	id, ok := lerID(c)
	if !ok {
		return
	}

	switch err := a.store.ExcluirAluno(c.Request.Context(), id); {
	case errors.Is(err, store.ErrNaoEncontrado):
		responderErro(c, http.StatusNotFound, "Aluno não encontrado.")
		return
	case err != nil:
		a.erroInterno(c, "excluindo aluno", err)
		return
	}

	a.invalidarAluno(c, id)
	c.Status(http.StatusNoContent)
}

func responderConflitoRA(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusConflict, erroJSON{
		Erro:   "Já existe um aluno cadastrado com este RA.",
		Campos: validate.Erros{"ra": "Este RA já está em uso por outro aluno."},
	})
}

// lerID extrai e valida o parâmetro :id da rota.
func lerID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		responderErro(c, http.StatusBadRequest, "Identificador inválido na URL.")
		return 0, false
	}
	return id, true
}
