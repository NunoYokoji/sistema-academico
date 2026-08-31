// Package store isola o acesso ao PostgreSQL. As consultas são montadas com
// o goqu; nenhuma outra camada conhece o banco.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres" // dialeto usado pelo goqu
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // driver "pgx" para database/sql

	"sistema-academico/backend/internal/model"
)

// Erros de domínio traduzidos pela camada HTTP em 404 e 409.
var (
	ErrNaoEncontrado = errors.New("registro não encontrado")
	ErrRADuplicado   = errors.New("já existe um aluno com este RA")
	ErrNotaExistente = errors.New("este aluno já possui notas lançadas")
)

// Tabelas com alias, usadas na montagem das consultas.
var (
	tabelaAluno = goqu.T("aluno")
	tabelaNota  = goqu.T("nota")
	alunoA      = tabelaAluno.As("a")
	notaN       = tabelaNota.As("n")
)

type Store struct {
	db *sql.DB
	qb *goqu.Database
}

// Abrir conecta no banco e valida a conexão.
func Abrir(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrindo conexão: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping no banco: %w", err)
	}

	// Por padrão o goqu interpola os valores direto na string SQL. No modo
	// preparado ele emite $1, $2, ..., o que elimina qualquer questão de
	// escaping e deixa o Postgres reaproveitar os planos de execução.
	goqu.SetDefaultPrepared(true)

	return &Store{db: db, qb: goqu.New("postgres", db)}, nil
}

func (s *Store) Fechar() error { return s.db.Close() }

// Ping é usado pelo endpoint /health.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

const esquema = `
CREATE TABLE IF NOT EXISTS aluno (
    id       SERIAL PRIMARY KEY,
    ra       VARCHAR(8)   NOT NULL UNIQUE,
    nome     VARCHAR(120) NOT NULL,
    email    VARCHAR(160) NOT NULL,
    curso    VARCHAR(120) NOT NULL,
    semestre SMALLINT     NOT NULL CHECK (semestre BETWEEN 1 AND 10)
);

CREATE TABLE IF NOT EXISTS nota (
    id       SERIAL PRIMARY KEY,
    aluno_id INTEGER      NOT NULL UNIQUE REFERENCES aluno(id) ON DELETE CASCADE,
    p1       NUMERIC(4,2) NOT NULL CHECK (p1 BETWEEN 0 AND 10),
    p2       NUMERIC(4,2) NOT NULL CHECK (p2 BETWEEN 0 AND 10),
    media    NUMERIC(4,2) NOT NULL,
    situacao VARCHAR(20)  NOT NULL
);`

// Migrar cria o esquema de forma idempotente. O advisory lock evita que as
// duas réplicas do back-end tentem criar as tabelas ao mesmo tempo.
//
// O goqu é um construtor de consultas, não de DDL: a criação das tabelas e
// o lock permanecem em SQL cru, executados na conexão diretamente.
func (s *Store) Migrar(ctx context.Context, semear bool) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(728104)`); err != nil {
		return fmt.Errorf("obtendo lock de migração: %w", err)
	}
	defer conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock(728104)`)

	if _, err := conn.ExecContext(ctx, esquema); err != nil {
		return fmt.Errorf("criando esquema: %w", err)
	}
	if !semear {
		return nil
	}
	return s.semear(ctx)
}

// semear popula o banco com alunos de exemplo, apenas se estiver vazio.
func (s *Store) semear(ctx context.Context) error {
	total, err := s.qb.From(tabelaAluno).CountContext(ctx)
	if err != nil {
		return fmt.Errorf("contando alunos: %w", err)
	}
	if total > 0 {
		return nil
	}

	const curso = "Análise e Desenvolvimento de Sistemas"
	alunos := []model.Aluno{
		{RA: "12345", Nome: "João Silva", Email: "joao.silva@email.com", Curso: curso, Semestre: 4},
		{RA: "12346", Nome: "Maria Souza", Email: "maria.souza@email.com", Curso: curso, Semestre: 4},
		{RA: "12347", Nome: "Pedro Santos", Email: "pedro.santos@email.com", Curso: curso, Semestre: 4},
		{RA: "12348", Nome: "Ana Beatriz Lima", Email: "ana.lima@email.com", Curso: curso, Semestre: 3},
		{RA: "12349", Nome: "Carlos Eduardo Ferreira", Email: "carlos.f@email.com", Curso: curso, Semestre: 4},
	}
	// O último aluno fica sem notas, para exercitar a situação "Sem nota".
	notas := []struct{ p1, p2 float64 }{{8, 7}, {5, 4}, {2, 3}, {9.5, 8.5}}

	for i := range alunos {
		if err := s.CriarAluno(ctx, &alunos[i]); err != nil {
			return fmt.Errorf("semeando alunos: %w", err)
		}
		if i >= len(notas) {
			continue
		}
		// Média e situação saem da mesma regra usada pela API, em vez de um
		// CASE WHEN duplicado em SQL.
		nota := model.Nota{AlunoID: alunos[i].ID, P1: notas[i].p1, P2: notas[i].p2}
		nota.Aplicar()
		if err := s.CriarNota(ctx, &nota); err != nil {
			return fmt.Errorf("semeando notas: %w", err)
		}
	}
	return nil
}

// linhaAluno espelha o resultado do LEFT JOIN entre aluno e nota. Os campos
// da nota são ponteiros porque vêm nulos quando o aluno ainda não tem notas.
type linhaAluno struct {
	ID       int64    `db:"id"`
	RA       string   `db:"ra"`
	Nome     string   `db:"nome"`
	Email    string   `db:"email"`
	Curso    string   `db:"curso"`
	Semestre int      `db:"semestre"`
	NotaID   *int64   `db:"nota_id"`
	P1       *float64 `db:"p1"`
	P2       *float64 `db:"p2"`
	Media    *float64 `db:"media"`
	Situacao *string  `db:"situacao"`
}

func (l linhaAluno) paraModelo() model.Aluno {
	aluno := model.Aluno{
		ID: l.ID, RA: l.RA, Nome: l.Nome,
		Email: l.Email, Curso: l.Curso, Semestre: l.Semestre,
	}
	if l.NotaID != nil {
		aluno.Nota = &model.Nota{
			ID: *l.NotaID, AlunoID: l.ID,
			P1: *l.P1, P2: *l.P2,
			Media: *l.Media, Situacao: *l.Situacao,
		}
	}
	return aluno
}

// selecaoAluno monta o SELECT com o LEFT JOIN. O id da nota é aliasado para
// "nota_id" porque colidiria com o id do aluno no resultado.
func (s *Store) selecaoAluno() *goqu.SelectDataset {
	return s.qb.
		From(alunoA).
		LeftJoin(notaN, goqu.On(goqu.I("n.aluno_id").Eq(goqu.I("a.id")))).
		Select(
			goqu.I("a.id"), goqu.I("a.ra"), goqu.I("a.nome"),
			goqu.I("a.email"), goqu.I("a.curso"), goqu.I("a.semestre"),
			goqu.I("n.id").As("nota_id"),
			goqu.I("n.p1"), goqu.I("n.p2"), goqu.I("n.media"), goqu.I("n.situacao"),
		)
}

// ListarAlunos devolve todos os alunos, ordenados por RA, com suas notas.
func (s *Store) ListarAlunos(ctx context.Context) ([]model.Aluno, error) {
	var linhas []linhaAluno
	if err := s.selecaoAluno().Order(goqu.I("a.ra").Asc()).
		Executor().ScanStructsContext(ctx, &linhas); err != nil {
		return nil, err
	}

	alunos := make([]model.Aluno, 0, len(linhas))
	for _, linha := range linhas {
		alunos = append(alunos, linha.paraModelo())
	}
	return alunos, nil
}

// BuscarAluno devolve um aluno pelo id, ou ErrNaoEncontrado.
func (s *Store) BuscarAluno(ctx context.Context, id int64) (model.Aluno, error) {
	var linha linhaAluno

	encontrado, err := s.selecaoAluno().Where(goqu.I("a.id").Eq(id)).
		Executor().ScanStructContext(ctx, &linha)
	if err != nil {
		return model.Aluno{}, err
	}
	if !encontrado {
		return model.Aluno{}, ErrNaoEncontrado
	}
	return linha.paraModelo(), nil
}

// CriarAluno insere e preenche o id gerado.
func (s *Store) CriarAluno(ctx context.Context, a *model.Aluno) error {
	_, err := s.qb.Insert(tabelaAluno).
		Rows(goqu.Record{
			"ra": a.RA, "nome": a.Nome, "email": a.Email,
			"curso": a.Curso, "semestre": a.Semestre,
		}).
		Returning("id").
		Executor().ScanValContext(ctx, &a.ID)
	if violaUnicidade(err) {
		return ErrRADuplicado
	}
	return err
}

// AtualizarAluno altera o cadastro; não mexe nas notas.
func (s *Store) AtualizarAluno(ctx context.Context, a *model.Aluno) error {
	resultado, err := s.qb.Update(tabelaAluno).
		Set(goqu.Record{
			"ra": a.RA, "nome": a.Nome, "email": a.Email,
			"curso": a.Curso, "semestre": a.Semestre,
		}).
		Where(goqu.C("id").Eq(a.ID)).
		Executor().ExecContext(ctx)
	if violaUnicidade(err) {
		return ErrRADuplicado
	}
	return conferirAfetadas(resultado, err)
}

// ExcluirAluno remove o aluno; a nota vai junto por ON DELETE CASCADE.
func (s *Store) ExcluirAluno(ctx context.Context, id int64) error {
	resultado, err := s.qb.Delete(tabelaAluno).
		Where(goqu.C("id").Eq(id)).
		Executor().ExecContext(ctx)
	return conferirAfetadas(resultado, err)
}

// AlunoExiste evita 500 em rotas de nota quando o aluno não existe.
func (s *Store) AlunoExiste(ctx context.Context, id int64) (bool, error) {
	total, err := s.qb.From(tabelaAluno).Where(goqu.C("id").Eq(id)).CountContext(ctx)
	return total > 0, err
}

func (s *Store) selecaoNota() *goqu.SelectDataset {
	return s.qb.From(tabelaNota).Select("id", "aluno_id", "p1", "p2", "media", "situacao")
}

// BuscarNotaDoAluno devolve a nota de um aluno.
func (s *Store) BuscarNotaDoAluno(ctx context.Context, alunoID int64) (model.Nota, error) {
	return lerNota(ctx, s.selecaoNota().Where(goqu.C("aluno_id").Eq(alunoID)))
}

// BuscarNota devolve a nota pelo id dela.
func (s *Store) BuscarNota(ctx context.Context, id int64) (model.Nota, error) {
	return lerNota(ctx, s.selecaoNota().Where(goqu.C("id").Eq(id)))
}

func lerNota(ctx context.Context, consulta *goqu.SelectDataset) (model.Nota, error) {
	var nota model.Nota

	encontrada, err := consulta.Executor().ScanStructContext(ctx, &nota)
	if err != nil {
		return model.Nota{}, err
	}
	if !encontrada {
		return model.Nota{}, ErrNaoEncontrado
	}
	return nota, nil
}

// CriarNota lança as notas de um aluno que ainda não as possui.
func (s *Store) CriarNota(ctx context.Context, n *model.Nota) error {
	_, err := s.qb.Insert(tabelaNota).
		Rows(goqu.Record{
			"aluno_id": n.AlunoID, "p1": n.P1, "p2": n.P2,
			"media": n.Media, "situacao": n.Situacao,
		}).
		Returning("id").
		Executor().ScanValContext(ctx, &n.ID)
	if violaUnicidade(err) {
		return ErrNotaExistente
	}
	return err
}

// AtualizarNota grava as novas notas já com média e situação recalculadas.
func (s *Store) AtualizarNota(ctx context.Context, n *model.Nota) error {
	resultado, err := s.qb.Update(tabelaNota).
		Set(goqu.Record{"p1": n.P1, "p2": n.P2, "media": n.Media, "situacao": n.Situacao}).
		Where(goqu.C("id").Eq(n.ID)).
		Executor().ExecContext(ctx)
	return conferirAfetadas(resultado, err)
}

// ExcluirNota remove o lançamento de notas.
func (s *Store) ExcluirNota(ctx context.Context, id int64) error {
	resultado, err := s.qb.Delete(tabelaNota).
		Where(goqu.C("id").Eq(id)).
		Executor().ExecContext(ctx)
	return conferirAfetadas(resultado, err)
}

func conferirAfetadas(resultado sql.Result, err error) error {
	if err != nil {
		return err
	}
	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afetadas == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// violaUnicidade identifica o SQLSTATE 23505 (unique_violation) do Postgres.
func violaUnicidade(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
