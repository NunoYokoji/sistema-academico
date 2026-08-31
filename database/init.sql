-- Estrutura do banco (seção 10 do enunciado).
--
-- Este arquivo é a referência do esquema e serve para inspeção/uso manual.
-- Na execução real quem cria as tabelas é o próprio back-end, na
-- inicialização, de forma idempotente e protegida por advisory lock — assim
-- o esquema também é criado no Kubernetes, onde o volume do Postgres já
-- existe e o /docker-entrypoint-initdb.d não é reexecutado.

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
);

-- Relacionamento 1:1 entre aluno e nota: cada aluno possui um lançamento de
-- notas na disciplina (UNIQUE em aluno_id) e excluir o aluno remove a nota.
