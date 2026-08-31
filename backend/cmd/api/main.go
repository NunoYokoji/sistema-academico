// Comando api: back-end do Sistema Acadêmico (API REST + PostgreSQL + Redis).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sistema-academico/backend/internal/cache"
	"sistema-academico/backend/internal/config"
	"sistema-academico/backend/internal/httpapi"
	"sistema-academico/backend/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := executar(); err != nil {
		slog.Error("aplicação encerrada com erro", "erro", err)
		os.Exit(1)
	}
}

func executar() error {
	cfg := config.Load()

	// O Pod do back-end pode subir antes do banco: tenta por até 60s.
	ctx, cancelarInicio := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelarInicio()

	banco, err := conectarComRetentativa(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer banco.Fechar()

	if err := banco.Migrar(ctx, cfg.Seed); err != nil {
		return err
	}
	slog.Info("banco pronto")

	redis := cache.Novo(cfg.RedisAddr, cfg.CacheTTL)
	defer redis.Fechar()

	if err := redis.Ping(ctx); err != nil {
		// Cache indisponível não impede a API de servir: apenas avisa.
		slog.Warn("redis indisponível no start; seguindo sem cache", "erro", err)
	} else {
		slog.Info("redis conectado", "endereco", cfg.RedisAddr, "ttl", cfg.CacheTTL.String())
	}

	rotas, err := httpapi.Nova(banco, redis).Rotas()
	if err != nil {
		return err
	}

	servidor := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           rotas,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	encerrar := make(chan os.Signal, 1)
	signal.Notify(encerrar, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("servidor ouvindo", "porta", cfg.Port)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("falha no servidor", "erro", err)
			encerrar <- syscall.SIGTERM
		}
	}()

	<-encerrar
	slog.Info("encerrando o servidor")

	// Desligamento gracioso: o Kubernetes envia SIGTERM ao reescalar Pods.
	ctxParada, cancelarParada := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelarParada()
	return servidor.Shutdown(ctxParada)
}

// conectarComRetentativa insiste na conexão enquanto o banco inicializa.
func conectarComRetentativa(ctx context.Context, dsn string) (*store.Store, error) {
	var ultimoErro error

	for tentativa := 1; ; tentativa++ {
		banco, err := store.Abrir(ctx, dsn)
		if err == nil {
			return banco, nil
		}
		ultimoErro = err
		slog.Warn("banco indisponível, tentando de novo", "tentativa", tentativa, "erro", err)

		select {
		case <-ctx.Done():
			return nil, errors.Join(errors.New("não foi possível conectar ao banco"), ultimoErro)
		case <-time.After(2 * time.Second):
		}
	}
}
