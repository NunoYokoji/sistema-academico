// Package cache implementa a camada de cache em Redis descrita nas seções
// 11 a 13 do enunciado: leitura com TTL e invalidação em toda escrita.
//
// O cache é sempre best-effort: se o Redis estiver indisponível, a
// requisição segue para o banco e a API continua respondendo normalmente.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ChaveListaAlunos guarda a resposta de GET /api/alunos.
const ChaveListaAlunos = "alunos:all"

// ChaveAluno guarda a resposta de GET /api/alunos/{id}.
func ChaveAluno(id int64) string { return "aluno:" + strconv.FormatInt(id, 10) }

type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

func Novo(endereco string, ttl time.Duration) *Cache {
	return &Cache{
		rdb: redis.NewClient(&redis.Options{Addr: endereco}),
		ttl: ttl,
	}
}

func (c *Cache) Fechar() error { return c.rdb.Close() }

// Ping é usado pelo endpoint /health.
func (c *Cache) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

// TTL expõe o tempo de expiração configurado, para logs e para o /health.
func (c *Cache) TTL() time.Duration { return c.ttl }

// Buscar tenta desserializar o valor da chave em destino.
// Devolve true no cache HIT; false no MISS ou em qualquer falha do Redis.
func (c *Cache) Buscar(ctx context.Context, chave string, destino any) bool {
	bruto, err := c.rdb.Get(ctx, chave).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			slog.Warn("falha ao ler do cache", "chave", chave, "erro", err)
		}
		return false
	}
	if err := json.Unmarshal(bruto, destino); err != nil {
		slog.Warn("valor inválido no cache", "chave", chave, "erro", err)
		return false
	}
	return true
}

// Guardar grava o valor serializado com o TTL configurado (SET ... EX).
func (c *Cache) Guardar(ctx context.Context, chave string, valor any) {
	bruto, err := json.Marshal(valor)
	if err != nil {
		slog.Warn("falha ao serializar para o cache", "chave", chave, "erro", err)
		return
	}
	if err := c.rdb.Set(ctx, chave, bruto, c.ttl).Err(); err != nil {
		slog.Warn("falha ao gravar no cache", "chave", chave, "erro", err)
	}
}

// Invalidar apaga as chaves afetadas por uma escrita. É chamado depois de
// todo POST, PUT e DELETE, para que a próxima leitura resulte em MISS e
// releia do banco.
func (c *Cache) Invalidar(ctx context.Context, chaves ...string) {
	if len(chaves) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, chaves...).Err(); err != nil {
		slog.Warn("falha ao invalidar o cache", "chaves", chaves, "erro", err)
		return
	}
	slog.Info("cache invalidado", "chaves", chaves)
}
