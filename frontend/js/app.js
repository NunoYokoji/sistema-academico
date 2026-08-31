// Ponto de entrada: menu responsivo e carga inicial dos dados da API.

import * as estado from "./estado.js";
import { feedback } from "./ui.js";
import "./alunos.js";
import "./notas.js";

const menuToggle = document.getElementById("menuToggle");
const mainNav = document.getElementById("mainNav");
const estadoLista = document.getElementById("estadoLista");
const feedbackLista = document.getElementById("feedbackLista");

menuToggle.addEventListener("click", () => {
  const aberto = mainNav.classList.toggle("aberto");
  menuToggle.classList.toggle("aberto", aberto);
  menuToggle.setAttribute("aria-expanded", String(aberto));
});

// Fecha o menu ao navegar, para não cobrir o conteúdo no celular.
mainNav.addEventListener("click", (evento) => {
  if (evento.target.classList.contains("nav-link")) {
    mainNav.classList.remove("aberto");
    menuToggle.classList.remove("aberto");
    menuToggle.setAttribute("aria-expanded", "false");
  }
});

async function iniciar() {
  try {
    await estado.recarregar();
  } catch (erro) {
    estadoLista.hidden = false;
    estadoLista.textContent = "Não foi possível carregar os alunos.";
    feedback(feedbackLista, `${erro.message} Verifique se o back-end está no ar e recarregue a página.`, "erro");
  }
}

iniciar();
