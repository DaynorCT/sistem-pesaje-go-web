export type PesajeFila = {
  id: number;
  guia?: string;
  peso_bruto?: number;
  tara?: number;
  peso_neto?: number;
  estado_balanza?: string;
};

/** Igual que pesaje-go motivoNoFinalizar */
export function motivoNoFinalizar(p: PesajeFila): string | null {
  const est = (p.estado_balanza || "").toUpperCase().trim();
  if (est !== "EN_PROCESO") return "El pesaje no está EN PROCESO en balanza.";
  const bruto = Number(p.peso_bruto) || 0;
  const tara = Number(p.tara) || 0;
  const neto = Number(p.peso_neto) || bruto - tara;
  if (bruto <= 0) return "Falta el peso bruto (use Capt. peso).";
  if (tara <= 0) return "Falta la tara (vehículo o Capt. tara / tara manual).";
  if (neto <= 0 || bruto <= tara) {
    return "El peso neto debe ser mayor que cero: bruto tiene que superar a la tara.";
  }
  return null;
}