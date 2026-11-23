/** Paridad pesaje-go frontend/src/dialogo.ts */
export type AnchoDialogo = "compacto" | "medio" | "amplio";

export function columnasDialogoPorCampos(visibles: number): 2 | 3 {
  return visibles > 7 ? 3 : 2;
}

export function anchoDialogoPorCampos(visibles: number): AnchoDialogo {
  const cols = columnasDialogoPorCampos(visibles);
  if (cols === 3) return "amplio";
  if (visibles <= 3) return "compacto";
  return "medio";
}