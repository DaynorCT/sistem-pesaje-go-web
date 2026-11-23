import { FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

export default function Login() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [branding, setBranding] = useState({ nombre: "Sistema de Pesaje", logo: "" });
  const [form, setForm] = useState({ usuario: "", contrasena: "" });

  useEffect(() => {
    fetch("/api/empresas/branding")
      .then((r) => r.json())
      .then((d) => {
        if (d.success && d.data) setBranding(d.data);
      })
      .catch(() => {});
  }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(form),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.message || "Error al iniciar sesión");
        return;
      }
      localStorage.setItem("pesaje_user", JSON.stringify(data.user));
      navigate("/inicio");
    } catch {
      setError("Error de conexión");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-header">
          {branding.logo ? (
            <img src={branding.logo} alt="Logo" className="login-logo" />
          ) : (
            <div className="login-logo-placeholder">🚛</div>
          )}
          <h1 className="login-title">{branding.nombre}</h1>
          <p className="login-sub">Medición y registro de peso</p>
        </div>
        <hr className="login-divider" />
        <form onSubmit={onSubmit} className="login-form">
          <label>
            Usuario
            <input
              type="email"
              autoComplete="username"
              value={form.usuario}
              onChange={(e) => setForm({ ...form, usuario: e.target.value })}
            />
          </label>
          <label>
            Contraseña
            <input
              type="password"
              autoComplete="current-password"
              value={form.contrasena}
              onChange={(e) => setForm({ ...form, contrasena: e.target.value })}
            />
          </label>
          {error && <p className="login-error">{error}</p>}
          <button type="submit" disabled={loading}>
            {loading ? "Ingresando…" : "Iniciar sesión"}
          </button>
        </form>
      </div>
    </div>
  );
}
