/** SVG locales — paridad con pesaje-go frontend/src/iconos-modulo.ts */

const attrs =
  'xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"';

export type ModuloIconId =
  | "inicio"
  | "usuarios"
  | "empresas"
  | "choferes"
  | "proveedores"
  | "pesaje"
  | "catalogo"
  | "vehiculos"
  | "productos"
  | "origenes"
  | "destinos"
  | "reportes"
  | "config";

const ICONOS: Record<ModuloIconId, string> = {
  inicio: `<svg ${attrs}><path d="M3 9.5 12 3l9 6.5V21a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1V9.5z"/></svg>`,
  usuarios: `<svg ${attrs}><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/></svg>`,
  empresas: `<svg ${attrs}><rect x="3" y="7" width="18" height="14" rx="1"/><path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>`,
  choferes: `<svg ${attrs}><circle cx="12" cy="6" r="3"/><path d="M6 21v-1a6 6 0 0 1 12 0v1"/></svg>`,
  proveedores: `<svg ${attrs}><path d="M3 21h18M6 21V9l6-4 6 4v12"/><path d="M10 21v-6h4v6"/></svg>`,
  pesaje: `<svg ${attrs}><path d="M4 20h16"/><path d="M6 20V8l6-3 6 3v12"/><path d="M9 12h6"/></svg>`,
  catalogo: `<svg ${attrs}><path d="M4 6h16M4 12h16M4 18h10"/><path d="m19 16 2 2 4-4"/></svg>`,
  vehiculos: `<svg ${attrs}><path d="M5 17h14v-5H5v5z"/><path d="M5 12 7 7h10l2 5"/><circle cx="7.5" cy="17" r="1.5"/><circle cx="16.5" cy="17" r="1.5"/></svg>`,
  productos: `<svg ${attrs}><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>`,
  origenes: `<svg ${attrs}><circle cx="12" cy="10" r="3"/><path d="M12 2a8 8 0 0 0-8 8c0 5.4 8 12 8 12s8-6.6 8-12a8 8 0 0 0-8-8z"/></svg>`,
  destinos: `<svg ${attrs}><path d="m3 11 19-9-9 19-2-8-8-2z"/></svg>`,
  reportes: `<svg ${attrs}><path d="M6 3v18"/><path d="M12 8v13"/><path d="M18 13v8"/></svg>`,
  config: `<svg ${attrs}><circle cx="12" cy="12" r="3"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`,
};

const attrsMenu =
  'xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"';

export const iconoUsuarioSvg = `<svg ${attrs}><circle cx="12" cy="8" r="4.25"/><path d="M5 21v-1a7 7 0 0 1 14 0v1"/></svg>`;
export const iconoSalirSvg = `<svg ${attrsMenu}><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5"/><path d="M21 12H9"/></svg>`;

const RUTA_A_ICONO: Record<string, ModuloIconId> = {
  "/inicio": "inicio",
  "/usuarios": "usuarios",
  "/empresa": "empresas",
  "/choferes": "choferes",
  "/proveedores": "proveedores",
  "/pesajes": "pesaje",
  "/catalogo": "catalogo",
  "/vehiculos": "vehiculos",
  "/productos": "productos",
  "/origenes": "origenes",
  "/destinos": "destinos",
  "/reportes": "reportes",
  "/config": "config",
};

export function iconoPorRuta(path: string): ModuloIconId {
  if (path.startsWith("/config/")) return "config";
  return RUTA_A_ICONO[path] ?? "inicio";
}

type IconoProps = { id: ModuloIconId; className?: string };

export function IconoModulo({ id, className = "shell-item-icon" }: IconoProps) {
  return <span className={className} dangerouslySetInnerHTML={{ __html: ICONOS[id] }} />;
}

export function IconoInline({ html, className }: { html: string; className?: string }) {
  return <span className={className} dangerouslySetInnerHTML={{ __html: html }} />;
}
