// Estado compartilhado da aplicação. A lista de alunos vive exclusivamente
// no servidor: aqui ela é apenas um espelho do último GET /api/alunos
// (a seção 6 do enunciado proíbe manter os dados só no JavaScript).

import { api } from "./api.js";

let alunos = [];
const ouvintes = [];

/** Registra um observador e já o executa com o estado atual. */
export function assinar(ouvinte) {
  ouvintes.push(ouvinte);
  ouvinte(alunos);
}

/** Última lista conhecida, sem ir ao servidor. */
export function alunosAtuais() {
  return alunos;
}

export function buscarPorId(id) {
  return alunos.find((aluno) => String(aluno.id) === String(id)) || null;
}

/** Recarrega a lista a partir da API e notifica a interface. */
export async function recarregar() {
  alunos = await api.listarAlunos();
  ouvintes.forEach((ouvinte) => ouvinte(alunos));
  return alunos;
}
