import { IconoInline } from "./iconos-modulo";

const SVG_TABLA =
  'xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"';

export const ICONO_EDITAR_TABLA = `<svg ${SVG_TABLA}><path d="M17 3a2.83 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>`;

export const ICONO_FINALIZAR_TABLA = `<svg ${SVG_TABLA}><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m22 4-10 10-3.5-3.5"/></svg>`;

export const ICONO_BOLETA_TABLA = `<svg ${SVG_TABLA}><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M8 13h8M8 17h5"/></svg>`;

type BtnProps = {
  tip: string;
  onClick?: () => void;
  act?: "editar" | "finalizar" | "boleta";
  disabled?: boolean;
};

export function TablaBtnIcono({ tip, onClick, act = "editar", disabled }: BtnProps) {
  const html =
    act === "finalizar" ? ICONO_FINALIZAR_TABLA : act === "boleta" ? ICONO_BOLETA_TABLA : ICONO_EDITAR_TABLA;
  return (
    <button
      type="button"
      className="abm-btn-editar tabla-btn-icono"
      data-tip={tip}
      data-act={act}
      aria-label={tip}
      disabled={disabled}
      onClick={onClick}
    >
      <IconoInline html={html} />
    </button>
  );
}

export function estadoEsActivo(estado: string): boolean {
  return (estado || "").toUpperCase() === "ACTIVO";
}

export function etiquetaSwitchEstado(activo: boolean): string {
  return activo ? "ACTIVO" : "INACTIVO";
}

type EstadoSwitchProps = {
  activo: boolean;
  onChange: (activo: boolean) => void;
  disabled?: boolean;
};

/** Interruptor de fila (paridad pesaje-go): no va en el diálogo. */
export function EstadoSwitch({ activo, onChange, disabled }: EstadoSwitchProps) {
  const txt = etiquetaSwitchEstado(activo);
  return (
    <label className="abm-switch" data-tip={txt}>
      <input
        type="checkbox"
        className="abm-switch-input"
        role="switch"
        checked={activo}
        disabled={disabled}
        aria-label={txt}
        aria-checked={activo}
        onClick={(e) => e.stopPropagation()}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span className="abm-switch-pista" />
    </label>
  );
}

export function EstadoEtiqueta({ estado }: { estado: string }) {
  const u = estado.toUpperCase();
  let clase = "estado-etiqueta--inactivo";
  if (u === "ACTIVO") clase = "estado-etiqueta--activo";
  else if (u.includes("PROCESO") || u.includes("PROCES")) clase = "estado-etiqueta--proceso";
  else if (u.includes("FINAL")) clase = "estado-etiqueta--finalizado";
  return (
    <td className="tabla-celda-estado">
      <span className={`estado-etiqueta ${clase}`}>{estado || "—"}</span>
    </td>
  );
}