import { ReactNode } from "react";

type Props = {
  etiqueta: string;
  requerido?: boolean;
  anchoCompleto?: boolean;
  children: ReactNode;
};

/** Campo de formulario ABM (paridad `pesaje-campo`). */
export function Campo({ etiqueta, requerido, anchoCompleto, children }: Props) {
  return (
    <label className="pesaje-campo" data-ancho-completo={anchoCompleto ? "" : undefined}>
      {etiqueta}
      {requerido ? " *" : ""}
      {children}
    </label>
  );
}