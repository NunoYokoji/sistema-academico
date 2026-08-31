// Cadastro, alteração e exclusão de alunos, mais a renderização da lista.

import { api, ErroApi } from "./api.js";
import * as estado from "./estado.js";
import {
  comBotaoOcupado, feedback, formatarNota, limparErros, mostrarErros, situacaoDaMedia,
} from "./ui.js";

const RA_REGEX = /^\d{5,8}$/;
const EMAIL_REGEX = /^[^@\s]+@[^@\s.]+(\.[^@\s.]+)+$/;

const formulario = document.getElementById("alunoForm");
const tituloForm = document.getElementById("tituloFormAluno");
const campoId = document.getElementById("alunoId");
const btnSalvar = document.getElementById("btnSalvarAluno");
const btnCancelar = document.getElementById("btnCancelarAluno");
const feedbackAluno = document.getElementById("feedbackAluno");

const corpoTabela = document.getElementById("alunosTableBody");
const estadoLista = document.getElementById("estadoLista");
const feedbackLista = document.getElementById("feedbackLista");
const totalAlunos = document.getElementById("totalAlunos");

// Colunas da tabela. O data-label alimenta o layout de cartões do mobile.
const colunas = [
  { label: "RA", valor: (a) => a.ra },
  { label: "Aluno", valor: (a) => a.nome },
  { label: "E-mail", valor: (a) => a.email },
  { label: "Curso", valor: (a) => a.curso },
  { label: "Semestre", valor: (a) => String(a.semestre) },
  { label: "P1", valor: (a) => formatarNota(a.nota && a.nota.p1) },
  { label: "P2", valor: (a) => formatarNota(a.nota && a.nota.p2) },
  { label: "Média", valor: (a) => formatarNota(a.nota && a.nota.media) },
];

/** Validação no cliente, espelhando as regras do back-end. */
function validar(dados) {
  const erros = {};

  if (!dados.ra) erros.ra = "O RA é obrigatório.";
  else if (!RA_REGEX.test(dados.ra)) erros.ra = "RA inválido: informe de 5 a 8 dígitos.";

  if (!dados.nome) erros.nome = "O nome é obrigatório.";
  else if (dados.nome.length < 3) erros.nome = "O nome deve ter ao menos 3 caracteres.";

  if (!dados.email) erros.email = "O e-mail é obrigatório.";
  else if (!EMAIL_REGEX.test(dados.email)) erros.email = "Formato de e-mail inválido (ex.: joao@email.com).";

  if (!dados.curso) erros.curso = "O curso é obrigatório.";

  if (dados.semestre === null) erros.semestre = "O semestre é obrigatório.";
  else if (!Number.isInteger(dados.semestre) || dados.semestre < 1 || dados.semestre > 10) {
    erros.semestre = "O semestre deve ser um número inteiro entre 1 e 10.";
  }

  return erros;
}

function lerFormulario() {
  const semestreBruto = document.getElementById("alunoSemestre").value.trim();
  return {
    ra: document.getElementById("alunoRa").value.trim(),
    nome: document.getElementById("alunoNome").value.trim(),
    email: document.getElementById("alunoEmail").value.trim(),
    curso: document.getElementById("alunoCurso").value.trim(),
    semestre: semestreBruto === "" ? null : Number(semestreBruto),
  };
}

/** Coloca o formulário em modo de edição do aluno informado. */
function editar(aluno) {
  campoId.value = aluno.id;
  document.getElementById("alunoRa").value = aluno.ra;
  document.getElementById("alunoNome").value = aluno.nome;
  document.getElementById("alunoEmail").value = aluno.email;
  document.getElementById("alunoCurso").value = aluno.curso;
  document.getElementById("alunoSemestre").value = aluno.semestre;

  tituloForm.textContent = `Alterando: ${aluno.nome}`;
  btnSalvar.textContent = "Salvar alterações";
  btnCancelar.hidden = false;

  limparErros(formulario);
  feedback(feedbackAluno, "");
  document.getElementById("cadastro-alunos").scrollIntoView({ behavior: "smooth", block: "start" });
}

/** Volta o formulário ao modo de cadastro. */
function encerrarEdicao() {
  formulario.reset();
  campoId.value = "";
  tituloForm.textContent = "Cadastro de Alunos";
  btnSalvar.textContent = "Cadastrar aluno";
  btnCancelar.hidden = true;
  limparErros(formulario);
}

async function excluir(aluno) {
  if (!window.confirm(`Excluir o aluno ${aluno.nome} (RA ${aluno.ra})? As notas dele também serão removidas.`)) {
    return;
  }

  try {
    await api.excluirAluno(aluno.id);
    await estado.recarregar();
    feedback(feedbackLista, `Aluno ${aluno.nome} excluído com sucesso.`);
    if (campoId.value === String(aluno.id)) encerrarEdicao();
  } catch (erro) {
    feedback(feedbackLista, erro.message, "erro");
  }
}

/** Redesenha a tabela a partir da lista vinda da API. */
function renderizar(alunos, idDestaque) {
  corpoTabela.replaceChildren();
  totalAlunos.textContent = String(alunos.length);
  estadoLista.hidden = alunos.length > 0;
  estadoLista.textContent = "Nenhum aluno cadastrado ainda. Use o formulário acima para começar.";

  alunos.forEach((aluno) => {
    const linha = document.createElement("tr");
    if (String(aluno.id) === String(idDestaque)) linha.classList.add("linha-destaque");

    colunas.forEach(({ label, valor }) => {
      const celula = document.createElement("td");
      celula.dataset.label = label;
      celula.textContent = valor(aluno);
      linha.appendChild(celula);
    });

    const celulaSituacao = document.createElement("td");
    celulaSituacao.dataset.label = "Situação";
    const selo = document.createElement("span");
    const situacao = situacaoDaMedia(aluno.nota ? aluno.nota.media : null);
    selo.className = `situacao ${situacao.classe}`;
    selo.textContent = aluno.nota ? aluno.nota.situacao : situacao.texto;
    celulaSituacao.appendChild(selo);
    linha.appendChild(celulaSituacao);

    const celulaAcoes = document.createElement("td");
    celulaAcoes.dataset.label = "Ações";
    celulaAcoes.className = "coluna-acoes";
    celulaAcoes.append(
      botaoAcao("Editar", "btn-acao", () => editar(aluno)),
      botaoAcao("Excluir", "btn-acao btn-acao-perigo", () => excluir(aluno)),
    );
    linha.appendChild(celulaAcoes);

    corpoTabela.appendChild(linha);
  });
}

function botaoAcao(rotulo, classe, aoClicar) {
  const botao = document.createElement("button");
  botao.type = "button";
  botao.className = classe;
  botao.textContent = rotulo;
  botao.addEventListener("click", aoClicar);
  return botao;
}

formulario.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  limparErros(formulario);
  feedback(feedbackAluno, "");
  feedback(feedbackLista, "");

  const dados = lerFormulario();
  const erros = validar(dados);

  if (Object.keys(erros).length > 0) {
    mostrarErros(formulario, erros);
    feedback(feedbackAluno, "Corrija os campos destacados antes de salvar.", "erro");
    return;
  }

  const id = campoId.value;

  try {
    const salvo = await comBotaoOcupado(btnSalvar, "Salvando…", () =>
      id ? api.atualizarAluno(id, dados) : api.criarAluno(dados));

    await estado.recarregar();
    renderizar(estado.alunosAtuais(), salvo.id);
    encerrarEdicao();
    feedback(feedbackAluno, id
      ? `Cadastro de ${salvo.nome} atualizado com sucesso.`
      : `Aluno ${salvo.nome} cadastrado com sucesso (RA ${salvo.ra}).`);
  } catch (erro) {
    // A API repete a validação; seus erros por campo são exibidos aqui.
    if (erro instanceof ErroApi) mostrarErros(formulario, erro.campos);
    feedback(feedbackAluno, erro.message, "erro");
  }
});

btnCancelar.addEventListener("click", () => {
  encerrarEdicao();
  feedback(feedbackAluno, "Edição cancelada.");
});

estado.assinar((alunos) => renderizar(alunos));

