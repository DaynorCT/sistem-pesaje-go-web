import type { ModuloIconId } from "./iconos-modulo";

export type NavItem = {
  id: string;
  label: string;
  icon: ModuloIconId;
  to?: string;
  /** Si está set, el ítem abre submenú con hijos del mismo grupo */
  grupo?: string;
};

export const NAV_ITEMS: NavItem[] = [
  { id: "inicio", label: "Inicio", icon: "inicio", to: "/inicio" },
  { id: "usuarios", label: "Usuarios", icon: "usuarios", to: "/usuarios" },
  { id: "empresa", label: "Empresa", icon: "empresas", to: "/empresa" },
  { id: "choferes", label: "Choferes", icon: "choferes", to: "/choferes" },
  { id: "proveedores", label: "Proveedores", icon: "proveedores", to: "/proveedores" },
  { id: "pesajes", label: "Pesaje", icon: "pesaje", to: "/pesajes" },
  { id: "catalogo", label: "Catálogo", icon: "catalogo", grupo: "Catálogo" },
  { id: "reportes", label: "Reportes", icon: "reportes", to: "/reportes" },
  { id: "config", label: "Config.", icon: "config", grupo: "Configuración" },
];

export const NAV_GRUPO_CATALOGO: NavItem[] = [
  { id: "vehiculos", label: "Vehículos", icon: "vehiculos", to: "/vehiculos" },
  { id: "productos", label: "Productos", icon: "productos", to: "/productos" },
  { id: "origenes", label: "Orígenes", icon: "origenes", to: "/origenes" },
  { id: "destinos", label: "Destinos", icon: "destinos", to: "/destinos" },
];

export const NAV_GRUPO_CONFIG: NavItem[] = [
  { id: "balanza", label: "Balanza", icon: "config", to: "/config/balanza" },
  { id: "boleta", label: "Boleta", icon: "config", to: "/config/boleta" },
];

export const RUTAS_CATALOGO = NAV_GRUPO_CATALOGO.map((i) => i.to!).filter(Boolean);
export const RUTAS_CONFIG = NAV_GRUPO_CONFIG.map((i) => i.to!).filter(Boolean);

export function hijosDeGrupo(nombre: string): NavItem[] {
  if (nombre === "Catálogo") return NAV_GRUPO_CATALOGO;
  if (nombre === "Configuración") return NAV_GRUPO_CONFIG;
  return [];
}
