import { ReactNode } from "react";

type Props = {
  titulo?: string;
  /** alias en inglés (compat) */
  title?: string;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
};

/** Cabecera ABM — paridad pesaje-go `abm-panel` / `abm-cabecera`. */
export default function ModuloPanel({ titulo, title, actions, children, className = "" }: Props) {
  const t = titulo || title || "";
  return (
    <div className={`abm-panel ${className}`.trim()}>
      <header className="abm-cabecera">
        <h2 className="titulo-tarjeta abm-titulo-modulo">{t}</h2>
        {actions ? <div className="abm-cabecera-der">{actions}</div> : null}
      </header>
      {children}
    </div>
  );
}