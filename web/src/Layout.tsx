import { NavLink, Outlet, useNavigate } from "react-router-dom";

const items = [
  { to: "/inicio", label: "Inicio" },
  { to: "/usuarios", label: "Usuarios" },
  { to: "/empresa", label: "Empresa" },
  { to: "/choferes", label: "Choferes" },
  { to: "/proveedores", label: "Proveedores" },
  { to: "/pesajes", label: "Pesaje" },
  { to: "/catalogo", label: "Catálogo" },
  { to: "/reportes", label: "Reportes" },
  { to: "/config", label: "Configuración" },
];

export default function Layout() {
  const navigate = useNavigate();
  const user = JSON.parse(localStorage.getItem("pesaje_user") || "null");

  if (!user) {
    navigate("/login");
    return null;
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">Sistema de Pesaje</div>
        <nav>
          {items.map((it) => (
            <NavLink key={it.to} to={it.to} className={({ isActive }) => (isActive ? "nav active" : "nav")}>
              {it.label}
            </NavLink>
          ))}
        </nav>
        <p className="sidebar-user">{user.email}</p>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  );
}
