import { useEffect, useRef, useState } from "react";
import { Navigate, Outlet, useNavigate } from "react-router-dom";
import { fetchApi } from "./api-client";
import ShellNav from "./components/ShellNav";
import { useBranding } from "./hooks/useBranding";
import { IconoInline, iconoSalirSvg, iconoUsuarioSvg } from "./iconos-modulo";

type User = { email?: string; nombre?: string };

export default function Layout() {
  const navigate = useNavigate();
  const { logoSrc, nombre } = useBranding();
  const user = JSON.parse(localStorage.getItem("pesaje_user") || "null") as User | null;
  const [menuAbierto, setMenuAbierto] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!user) return;
    void fetchApi("/api/choferes").then(({ sessionExpired }) => {
      if (sessionExpired) navigate("/login", { replace: true });
    });
  }, [user, navigate]);

  useEffect(() => {
    const cerrar = (ev: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(ev.target as Node)) {
        setMenuAbierto(false);
      }
    };
    document.addEventListener("click", cerrar);
    return () => document.removeEventListener("click", cerrar);
  }, []);

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  function cerrarSesion() {
    localStorage.removeItem("pesaje_user");
    navigate("/login", { replace: true });
  }

  return (
    <div className="shell root">
      <header className="barra-superior shell-cabecera">
        <div className="shell-fila-principal">
          <img className="shell-logo" src={logoSrc} alt={nombre} width={132} />
          <ShellNav />
          <div className="shell-usuario" ref={menuRef}>
            <button
              type="button"
              className="boton-usuario-cabecera"
              aria-label="Menú de usuario"
              title="Usuario"
              onClick={(ev) => {
                ev.stopPropagation();
                setMenuAbierto((v) => !v);
              }}
            >
              <IconoInline html={iconoUsuarioSvg} className="shell-usuario-icon" />
            </button>
            {menuAbierto ? (
              <div className="menu-flotante shell-menu-usuario" role="menu">
                <div className="shell-menu-usuario-cabecera">
                  <p className="shell-menu-usuario-nombre">{user.nombre || "Usuario"}</p>
                  <p className="shell-menu-usuario-correo">{user.email}</p>
                </div>
                <button
                  type="button"
                  className="shell-menu-usuario-item shell-menu-usuario-item--salir"
                  onClick={cerrarSesion}
                >
                  <IconoInline html={iconoSalirSvg} className="shell-item-icon shell-menu-item-icono" />
                  Cerrar sesión
                </button>
              </div>
            ) : null}
          </div>
        </div>
      </header>
      <main className="shell-contenido shell-contenido-abm">
        <Outlet />
      </main>
    </div>
  );
}