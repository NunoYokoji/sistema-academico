const RA_REGEX = /^\d{5}$/;

const alunos = [
  { ra: "12345", nome: "João Silva", p1: 8.0, p2: 7.0 },
  { ra: "12346", nome: "Maria Souza", p1: 5.0, p2: 4.0 },
  { ra: "12347", nome: "Pedro Santos", p1: 2.0, p2: 3.0 },
  { ra: "12348", nome: "Ana Beatriz Lima", p1: 9.5, p2: 8.5 },
  { ra: "12349", nome: "Carlos Eduardo Ferreira", p1: null, p2: null },
  { ra: "12350", nome: "Fernanda Oliveira", p1: 6.0, p2: 6.5 },
];

const form = document.getElementById("notasForm");
const alunoSelect = document.getElementById("alunoSelect");
const raInput = document.getElementById("raInput");
const notaP1 = document.getElementById("notaP1");
const notaP2 = document.getElementById("notaP2");
const erroAluno = document.getElementById("erroAluno");
const erroRa = document.getElementById("erroRa");
const erroP1 = document.getElementById("erroP1");
const erroP2 = document.getElementById("erroP2");
const formFeedback = document.getElementById("formFeedback");
const tabelaBody = document.getElementById("alunosTableBody");
const menuToggle = document.getElementById("menuToggle");
const mainNav = document.getElementById("mainNav");

function calcularMedia(aluno) {
  if (aluno.p1 === null || aluno.p2 === null) return null;
  // Arredonda para uma casa decimal para que a média exibida na tabela
  // seja exatamente a mesma usada na definição da situação do aluno.
  return Math.round(((aluno.p1 + aluno.p2) / 2) * 10) / 10;
}

function definirSituacao(media) {
  if (media === null) return { texto: "Sem nota", classe: "situacao-pendente" };
  if (media >= 6) return { texto: "Aprovado", classe: "situacao-aprovado" };
  if (media >= 4) return { texto: "Exame", classe: "situacao-exame" };
  return { texto: "Reprovado", classe: "situacao-reprovado" };
}

function formatarNota(nota) {
  return nota === null ? "—" : nota.toFixed(1).replace(".", ",");
}

function preencherSelect() {
  alunos.forEach((aluno) => {
    const option = document.createElement("option");
    option.value = aluno.ra;
    option.textContent = `${aluno.ra} — ${aluno.nome}`;
    alunoSelect.appendChild(option);
  });
}

function renderizarTabela(raDestaque) {
  tabelaBody.innerHTML = "";

  alunos.forEach((aluno) => {
    const media = calcularMedia(aluno);
    const situacao = definirSituacao(media);

    const tr = document.createElement("tr");
    if (aluno.ra === raDestaque) tr.classList.add("linha-destaque");

    const celulas = [
      { label: "RA", valor: aluno.ra },
      { label: "Aluno", valor: aluno.nome },
      { label: "P1", valor: formatarNota(aluno.p1) },
      { label: "P2", valor: formatarNota(aluno.p2) },
      { label: "Média", valor: formatarNota(media) },
    ];

    celulas.forEach(({ label, valor }) => {
      const td = document.createElement("td");
      td.dataset.label = label;
      td.textContent = valor;
      tr.appendChild(td);
    });

    const tdSituacao = document.createElement("td");
    tdSituacao.dataset.label = "Situação";
    const badge = document.createElement("span");
    badge.className = `situacao ${situacao.classe}`;
    badge.textContent = situacao.texto;
    tdSituacao.appendChild(badge);
    tr.appendChild(tdSituacao);

    tabelaBody.appendChild(tr);
  });
}

function limparErros() {
  [erroAluno, erroRa, erroP1, erroP2].forEach((el) => (el.textContent = ""));
  [alunoSelect, raInput, notaP1, notaP2].forEach((el) => el.classList.remove("invalido"));
  formFeedback.textContent = "";
  formFeedback.className = "form-feedback";
}

function marcarErro(campo, elementoErro, mensagem) {
  campo.classList.add("invalido");
  elementoErro.textContent = mensagem;
}

function validarNota(campo, elementoErro, rotulo) {
  const bruto = campo.value.trim();

  if (bruto === "") {
    marcarErro(campo, elementoErro, `A nota da ${rotulo} é obrigatória.`);
    return null;
  }

  const valor = Number(bruto.replace(",", "."));

  if (Number.isNaN(valor)) {
    marcarErro(campo, elementoErro, `A nota da ${rotulo} deve ser um número válido.`);
    return null;
  }

  if (valor < 0 || valor > 10) {
    marcarErro(campo, elementoErro, `A nota da ${rotulo} deve estar entre 0 e 10.`);
    return null;
  }

  return valor;
}

function preencherNotasDoAluno(aluno) {
  if (!aluno) return;

  notaP1.value = aluno.p1 === null ? "" : aluno.p1;
  notaP2.value = aluno.p2 === null ? "" : aluno.p2;
}

function validarRa() {
  const bruto = raInput.value.trim();

  if (bruto === "") {
    marcarErro(raInput, erroRa, "O RA é obrigatório.");
    return null;
  }

  if (!RA_REGEX.test(bruto)) {
    marcarErro(raInput, erroRa, "RA inválido: informe exatamente 5 dígitos (ex.: 12345).");
    return null;
  }

  const aluno = alunos.find((a) => a.ra === bruto);

  if (!aluno) {
    marcarErro(raInput, erroRa, `Nenhum aluno da disciplina possui o RA ${bruto}.`);
    return null;
  }

  return aluno;
}

form.addEventListener("submit", (evento) => {
  evento.preventDefault();
  limparErros();

  let valido = true;

  if (!alunoSelect.value) {
    marcarErro(alunoSelect, erroAluno, "Selecione um aluno.");
    valido = false;
  }

  const alunoDoRa = validarRa();
  if (alunoDoRa === null) valido = false;

  const p1 = validarNota(notaP1, erroP1, "P1");
  if (p1 === null) valido = false;

  const p2 = validarNota(notaP2, erroP2, "P2");
  if (p2 === null) valido = false;

  if (!valido) {
    formFeedback.textContent = "Corrija os campos destacados antes de salvar.";
    formFeedback.classList.add("erro");
    return;
  }

  const aluno = alunoDoRa;
  aluno.p1 = p1;
  aluno.p2 = p2;

  const media = calcularMedia(aluno);
  const situacao = definirSituacao(media);

  renderizarTabela(aluno.ra);

  formFeedback.textContent = `Notas de ${aluno.nome} salvas. Média ${formatarNota(media)} — ${situacao.texto}.`;
  formFeedback.classList.add("sucesso");
});

form.addEventListener("reset", () => {
  window.setTimeout(limparErros, 0);
});

alunoSelect.addEventListener("change", () => {
  limparErros();
  const aluno = alunos.find((a) => a.ra === alunoSelect.value);
  raInput.value = aluno ? aluno.ra : "";
  preencherNotasDoAluno(aluno);
});

raInput.addEventListener("input", () => {
  const aluno = alunos.find((a) => a.ra === raInput.value);
  if (aluno) {
    limparErros();
    alunoSelect.value = aluno.ra;
    preencherNotasDoAluno(aluno);
  }
});

menuToggle.addEventListener("click", () => {
  const aberto = mainNav.classList.toggle("aberto");
  menuToggle.classList.toggle("aberto", aberto);
  menuToggle.setAttribute("aria-expanded", String(aberto));
});

mainNav.addEventListener("click", (evento) => {
  if (evento.target.classList.contains("nav-link")) {
    mainNav.classList.remove("aberto");
    menuToggle.classList.remove("aberto");
    menuToggle.setAttribute("aria-expanded", "false");
  }
});

preencherSelect();
renderizarTabela();
