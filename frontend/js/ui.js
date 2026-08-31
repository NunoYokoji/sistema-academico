// Utilitários de interface compartilhados pelos dois formulários.

/** Formata uma nota no padrão brasileiro; "—" quando não há valor. */
export function formatarNota(valor) {
  if (valor === null || valor === undefined) return "—";
  return Number(valor).toFixed(1).replace(".", ",");
}

/** Traduz a média na situação acadêmica, com a classe do selo correspondente. */
export function situacaoDaMedia(media) {
  if (media === null || media === undefined) return { texto: "Sem nota", classe: "situacao-pendente" };
  if (media >= 6) return { texto: "Aprovado", classe: "situacao-aprovado" };
  if (media >= 4) return { texto: "Exame", classe: "situacao-exame" };
  return { texto: "Reprovado", classe: "situacao-reprovado" };
}

/** Converte "7,5" ou "7.5" em número; devolve null quando não for numérico. */
export function paraNumero(texto) {
  const bruto = String(texto).trim().replace(",", ".");
  if (bruto === "") return null;
  const valor = Number(bruto);
  return Number.isNaN(valor) ? null : valor;
}

/** Limpa mensagens de erro e destaques de um formulário. */
export function limparErros(formulario) {
  formulario.querySelectorAll(".error-message").forEach((el) => (el.textContent = ""));
  formulario.querySelectorAll(".invalido").forEach((el) => el.classList.remove("invalido"));
}

/**
 * Exibe os erros por campo. Aceita tanto o mapa da validação local quanto o
 * objeto "campos" devolvido pela API — os dois usam os mesmos nomes.
 */
export function mostrarErros(formulario, campos) {
  Object.entries(campos || {}).forEach(([campo, mensagem]) => {
    const alvo = formulario.querySelector(`[data-erro="${campo}"]`);
    if (alvo) alvo.textContent = mensagem;

    const entrada = formulario.querySelector(`[name="${campo}"]`);
    if (entrada) entrada.classList.add("invalido");
  });
}

/** Escreve uma mensagem de sucesso ou de erro no rodapé de uma seção. */
export function feedback(elemento, mensagem, tipo = "sucesso") {
  elemento.textContent = mensagem;
  elemento.className = `form-feedback ${mensagem ? tipo : ""}`.trim();
}

/** Desabilita o botão enquanto a requisição está em andamento. */
export async function comBotaoOcupado(botao, rotulo, acao) {
  const original = botao.textContent;
  botao.disabled = true;
  botao.textContent = rotulo;
  try {
    return await acao();
  } finally {
    botao.disabled = false;
    botao.textContent = original;
  }
}
