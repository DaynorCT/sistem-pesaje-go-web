import { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { IconoModulo, type ModuloIconId } from "../iconos-modulo";
import { hijosDeGrupo, NAV_ITEMS, RUTAS_CATALOGO, RUTAS_CONFIG, type NavItem } from "../navegacion";

function itemActivo(pathname: string, item: NavItem): boolean {
  if (item.to && pathname === item.to) return true;
  if (item.grupo === "Catálogo" && (pathname === "/catalogo" || RUTAS_CATALOGO.includes(pathname))) return true;
  if (item.grupo === "Configuración" && (pathname === "/config" || RUTAS_CONFIG.some((p) => pathname.startsWith(p)))) return true;
  return false;
}

export default function ShellNav() {
  const location = useLocation();
  const navigate = useNavigate();
  const navRef = useRef<HTMLElement>(null);
  const [submenuAbierto, setSubmenuAbierto] = useState<string | null>(null);

  useEffect(() => {
    const cerrar = (ev: MouseEvent) => {
      if (navRef.current && !navRef.current.contains(ev.target as Node)) {
        setSubmenuAbierto(null);
      }
    };
    document.addEventListener("click", cerrar);
    return () => document.removeEventListener("click", cerrar);
  }, []);

  const toggleSubmenu = (item: NavItem, ev: { stopPropagation: () => void }) => {
    ev.stopPropagation();
    if (!item.grupo) return;
    setSubmenuAbierto((prev) => (prev === item.grupo ? null : item.grupo ?? null));
  };

  const ir = (to: string) => {
    setSubmenuAbierto(null);
    navigate(to);
  };

  return (
    <nav className="shell-nav" aria-label="Módulos" ref={navRef}>
      {NAV_ITEMS.map((item) => {
        const activo = itemActivo(location.pathname, item);
        const hijos = item.grupo ? hijosDeGrupo(item.grupo) : [];
        const submenuVisible = Boolean(item.grupo && submenuAbierto === item.grupo);

        if (hijos.length > 0) {
          return (
            <div key={item.id} className="shell-menu-ancla">
              <button
                type="button"
                className={`shell-item-menu${activo ? " activo" : ""}`}
                onClick={(ev) => toggleSubmenu(item, ev)}
                aria-expanded={submenuVisible ? true : false}
                aria-haspopup="menu"
              >
                <IconoModulo id={item.icon} />
                <span className="shell-item-label">{item.label}</span>
                <span className="shell-item-chevron" aria-hidden>
                  ▾
                </span>
              </button>
              {submenuVisible ? (
                <div className="menu-flotante shell-submenu" role="menu">
                  <p className="subtitulo menu-encabezado">{item.grupo}</p>
                  <hr className="shell-submenu-sep" />
                  {hijos.map((h) => (
                    <button
                      key={h.id}
                      type="button"
                      className="item-menu shell-submenu-item shell-item-menu"
                      role="menuitem"
                      onClick={(ev) => {
                        ev.stopPropagation();
                        if (h.to) ir(h.to);
                      }}
                    >
                      <IconoModulo id={h.icon as ModuloIconId} />
                      <span>{h.label}</span>
                    </button>
                  ))}
                </div>
              ) : null}
            </div>
          );
        }

        return (
          <button
            key={item.id}
            type="button"
            className={`shell-item-menu${activo ? " activo" : ""}`}
            onClick={() => item.to && ir(item.to)}
          >
            <IconoModulo id={item.icon} />
            <span className="shell-item-label">{item.label}</span>
          </button>
        );
      })}
    </nav>
  );
}
