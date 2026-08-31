// Lançamento e alteração das notas de um aluno.

import { api, ErroApi } from "./api.js";
import * as estado from "./estado.js";
import {
  comBotaoOcupado, feedback, limparErros, mostrarErros, paraNumero, situacaoDaMedia,
} from "./ui.js";

const formulario = document.getElementById("notasForm");
const seletor = document.getElementById("alunoSelect");
const campoRa = document.getElementById("raNota");
const campoP1 = document.getElementById("notaP1");
const campoP2 = document.getElementById("notaP2");
const btnSalvar = formulario.querySelector('button[type="submit"]');
const btnLimpar = document.getElementById("btnLimparNotas");
const btnExcluir = document.getElementById("btnExcluirNotas");
const feedbackNota = document.getElementById("feedbackNota");

const previa = document.getElementById("previaNota");
const previaMedia = document.getElementById("previaMedia");
const previaSituacao = document.getElementById("previaSituacao");

/** Mantém as opções do seletor em sincronia com a lista de alunos. */
function preencherSeletor(alunos) {
  const selecionado = seletor.value;

  seletor.replaceChildren(new Option("Selecione um aluno", "", true, true));
  seletor.firstChild.disabled = true;

  alunos.forEach((aluno) => {
    seletor.appendChild(new Option(`${aluno.ra} — ${aluno.nome}`, aluno.id));
  });

  // Preserva a seleção quando o aluno continua existindo após um recarregamento.
  if (selecionado && estado.buscarPorId(selecionado)) {
    seletor.value = selecionado;
    aplicarSelecao();
  } else if (selecionado) {
    limparFormulario();
  }
}

/** Preenche RA e notas atuais do aluno escolhido. */
function aplicarSelecao() {
  const aluno = estado.buscarPorId(seletor.value);
  if (!aluno) return;

  campoRa.value = aluno.ra;
  campoP1.value = aluno.nota ? aluno.nota.p1 : "";
  campoP2.value = aluno.nota ? aluno.nota.p2 : "";
  btnExcluir.hidden = !aluno.nota;
  btnSalvar.textContent = aluno.nota ? "Alterar notas" : "Salvar notas";
  atualizarPrevia();
}

function limparFormulario() {
  formulario.reset();
  seletor.value = "";
  campoRa.value = "";
  btnExcluir.hidden = true;
  btnSalvar.textContent = "Salvar notas";
  previa.hidden = true;
  limparErros(formulario);
}

/**
 * Mostra a média antes de salvar. É apenas um auxílio visual: o valor
 * gravado é sempre o calculado pelo back-end.
 */
function atualizarPrevia() {
  const p1 = paraNumero(campoP1.value);
  const p2 = paraNumero(campoP2.value);

  if (p1 === null || p2 === null || p1 < 0 || p1 > 10 || p2 < 0 || p2 > 10) {
    previa.hidden = true;
    return;
  }

  const media = Math.round(((p1 + p2) / 2) * 100) / 100;
  const situacao = situacaoDaMedia(media);

  previaMedia.textContent = media.toFixed(1).replace(".", ",");
  previaSituacao.textContent = situacao.texto;
  previaSituacao.className = `situacao ${situacao.classe}`;
  previa.hidden = false;
}

/** Validação no cliente, espelhando as regras do back-end. */
function validar(p1, p2) {
  const erros = {};

  if (!seletor.value) erros.aluno = "Selecione um aluno.";

  [["p1", "P1", p1], ["p2", "P2", p2]].forEach(([campo, rotulo, valor]) => {
    if (valor === null) erros[campo] = `A nota da ${rotulo} é obrigatória.`;
    else if (valor < 0 || valor > 10) erros[campo] = `A nota da ${rotulo} deve estar entre 0 e 10.`;
  });

  return erros;
}

formulario.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  limparErros(formulario);
  feedback(feedbackNota, "");

  const p1 = paraNumero(campoP1.value);
  const p2 = paraNumero(campoP2.value);
  const erros = validar(p1, p2);

  if (Object.keys(erros).length > 0) {
    mostrarErros(formulario, erros);
    feedback(feedbackNota, "Corrija os campos destacados antes de salvar.", "erro");
    return;
  }

  const aluno = estado.buscarPorId(seletor.value);

  try {
    // Sem notas ainda: cria o lançamento. Já existentes: altera.
    const nota = await comBotaoOcupado(btnSalvar, "Salvando…", () =>
      aluno.nota
        ? api.atualizarNota(aluno.nota.id, { p1, p2 })
        : api.criarNota(aluno.id, { p1, p2 }));

    await estado.recarregar();
    feedback(feedbackNota,
      `Notas de ${aluno.nome} salvas. Média ${nota.media.toFixed(1).replace(".", ",")} — ${nota.situacao}.`);
  } catch (erro) {
    if (erro instanceof ErroApi) mostrarErros(formulario, erro.campos);
    feedback(feedbackNota, erro.message, "erro");
  }
});

btnExcluir.addEventListener("click", async () => {
  const aluno = estado.buscarPorId(seletor.value);
  if (!aluno || !aluno.nota) return;

  if (!window.confirm(`Excluir as notas de ${aluno.nome}?`)) return;

  try {
    await api.excluirNota(aluno.nota.id);
    await estado.recarregar();
    feedback(feedbackNota, `Notas de ${aluno.nome} excluídas.`);
  } catch (erro) {
    feedback(feedbackNota, erro.message, "erro");
  }
});

btnLimpar.addEventListener("click", () => {
  limparFormulario();
  feedback(feedbackNota, "");
});

seletor.addEventListener("change", () => {
  limparErros(formulario);
  feedback(feedbackNota, "");
  aplicarSelecao();
});

[campoP1, campoP2].forEach((campo) => campo.addEventListener("input", atualizarPrevia));

estado.assinar(preencherSeletor);
