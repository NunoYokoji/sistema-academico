// Cliente da API REST. Todas as chamadas usam caminhos relativos (/api/...):
// o nginx que serve este front faz o proxy para o back-end, então não há
// URL de servidor fixada aqui nem necessidade de CORS.

const BASE = "/api";

// ErroApi carrega o status HTTP e o mapa de erros por campo devolvido pela
// API, para que o formulário possa destacar cada input individualmente.
export class ErroApi extends Error {
  constructor(mensagem, status, campos) {
    super(mensagem);
    this.name = "ErroApi";
    this.status = status;
    this.campos = campos || {};
  }
}

async function requisitar(caminho, opcoes = {}) {
  let resposta;

  try {
    resposta = await fetch(BASE + caminho, {
      headers: opcoes.corpo ? { "Content-Type": "application/json" } : undefined,
      method: opcoes.metodo || "GET",
      body: opcoes.corpo ? JSON.stringify(opcoes.corpo) : undefined,
    });
  } catch (erro) {
    throw new ErroApi("Não foi possível falar com o servidor. Verifique sua conexão.", 0);
  }

  if (resposta.status === 204) return null;

  const conteudo = resposta.headers.get("Content-Type") || "";
  const corpo = conteudo.includes("application/json") ? await resposta.json() : null;

  if (!resposta.ok) {
    const mensagem = (corpo && corpo.erro) || `Erro inesperado (HTTP ${resposta.status}).`;
    throw new ErroApi(mensagem, resposta.status, corpo && corpo.campos);
  }

  return corpo;
}

export const api = {
  listarAlunos: () => requisitar("/alunos"),
  buscarAluno: (id) => requisitar(`/alunos/${id}`),
  criarAluno: (aluno) => requisitar("/alunos", { metodo: "POST", corpo: aluno }),
  atualizarAluno: (id, aluno) => requisitar(`/alunos/${id}`, { metodo: "PUT", corpo: aluno }),
  excluirAluno: (id) => requisitar(`/alunos/${id}`, { metodo: "DELETE" }),

  criarNota: (alunoId, notas) => requisitar(`/alunos/${alunoId}/notas`, { metodo: "POST", corpo: notas }),
  atualizarNota: (notaId, notas) => requisitar(`/notas/${notaId}`, { metodo: "PUT", corpo: notas }),
  excluirNota: (notaId) => requisitar(`/notas/${notaId}`, { metodo: "DELETE" }),
};
