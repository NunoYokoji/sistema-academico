package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sistema-academico/backend/internal/cache"
	"sistema-academico/backend/internal/store"
	"sistema-academico/backend/internal/validate"
)

// tamanhoMaximoCorpo limita o corpo das requisições a 1 MiB.
const tamanhoMaximoCorpo = 1 << 20

// API reúne as dependências dos handlers.
type API struct {
	store *store.Store
	cache *cache.Cache
}

func Nova(s *store.Store, c *cache.Cache) *API {
	return &API{store: s, cache: c}
}

// Rotas monta o roteador Gin com todos os endpoints da seção 9 do enunciado.
func (a *API) Rotas() (http.Handler, error) {
	// Registra a tag customizada de RA e faz o validador reportar os nomes
	// vindos das tags json.
	if err := validate.Registrar(); err != nil {
		return nil, err
	}

	// gin.New e não gin.Default: o logger de texto padrão do Gin
	// substituiria o log estruturado em JSON que os Pods emitem.
	r := gin.New()
	r.Use(gin.Recovery(), limitarCorpo(), cors(), registrar())

	api := r.Group("/api")
	{
		obter(api, "/alunos", a.listarAlunos)
		obter(api, "/alunos/:id", a.buscarAluno)
		api.POST("/alunos", a.criarAluno)
		api.PUT("/alunos/:id", a.atualizarAluno)
		api.DELETE("/alunos/:id", a.excluirAluno)

		obter(api, "/alunos/:id/notas", a.buscarNotasDoAluno)
		api.POST("/alunos/:id/notas", a.criarNota)
		api.PUT("/notas/:id", a.atualizarNota)
		api.DELETE("/notas/:id", a.excluirNota)
	}

	obter(r, "/health", a.saude)

	// Rota desconhecida responde JSON, não a página de texto padrão do Gin.
	r.NoRoute(func(c *gin.Context) {
		responderErro(c, http.StatusNotFound, "Rota não encontrada.")
	})

	return r, nil
}

// obter registra a rota em GET e em HEAD. O roteador do Gin, ao contrário
// do ServeMux da biblioteca padrão, não atende HEAD automaticamente nas
// rotas GET — e é com HEAD (curl -I) que se inspeciona o header X-Cache.
// O net/http descarta o corpo da resposta em requisições HEAD, então o
// mesmo handler serve os dois métodos.
func obter(r gin.IRoutes, rota string, handler gin.HandlerFunc) {
	r.GET(rota, handler)
	r.HEAD(rota, handler)
}

// saude alimenta as probes de readiness e liveness do Kubernetes.
func (a *API) saude(c *gin.Context) {
	ctx, cancelar := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancelar()

	estado := map[string]string{"status": "ok", "banco": "ok", "redis": "ok"}
	status := http.StatusOK

	if err := a.store.Ping(ctx); err != nil {
		estado["banco"], estado["status"] = "indisponível", "degradado"
		status = http.StatusServiceUnavailable
	}
	// O Redis é um cache best-effort: sua queda degrada, mas não derruba.
	if err := a.cache.Ping(ctx); err != nil {
		estado["redis"] = "indisponível"
	}

	c.JSON(status, estado)
}

// invalidarAluno apaga as duas chaves afetadas por uma escrita: a lista
// completa e o registro individual do aluno (seção 13).
func (a *API) invalidarAluno(c *gin.Context, alunoID int64) {
	a.cache.Invalidar(c.Request.Context(), cache.ChaveListaAlunos, cache.ChaveAluno(alunoID))
}

// erroInterno registra a causa real no log e devolve 500 sem vazar detalhes.
func (a *API) erroInterno(c *gin.Context, contexto string, err error) {
	slog.Error("erro interno", "contexto", contexto, "erro", err)
	responderErro(c, http.StatusInternalServerError,
		"Erro interno no servidor. Tente novamente em instantes.")
}

// marcarCache expõe o resultado da consulta ao Redis, o que torna a
// demonstração de MISS/HIT visível em um simples curl -I.
func marcarCache(c *gin.Context, resultado string) {
	c.Header("X-Cache", resultado)
}

// limitarCorpo protege contra corpos de requisição desproporcionais.
func limitarCorpo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, tamanhoMaximoCorpo)
		c.Next()
	}
}

// cors libera o consumo direto da API em desenvolvimento (o front servido
// pelo nginx usa caminhos relativos e nem chega a precisar disto).
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Expose-Headers", "X-Cache")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// registrar loga cada requisição com status, duração e resultado do cache.
func registrar() gin.HandlerFunc {
	return func(c *gin.Context) {
		inicio := time.Now()

		c.Next()

		slog.Info("requisição",
			"metodo", c.Request.Method,
			"rota", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"cache", cmpOu(c.Writer.Header().Get("X-Cache"), "-"),
			"duracao", time.Since(inicio).Round(time.Microsecond).String(),
		)
	}
}

func cmpOu(valor, padrao string) string {
	if valor == "" {
		return padrao
	}
	return valor
}
