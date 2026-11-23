import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useBranding } from "./hooks/useBranding";

export default function Login() {
  const navigate = useNavigate();
  const { logoSrc, nombre } = useBranding();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [form, setForm] = useState({ usuario: "", contrasena: "" });

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
        setError(data.message || "No se pudo iniciar sesión.");
        return;
      }
      localStorage.setItem("pesaje_user", JSON.stringify(data.user));
      navigate("/inicio");
    } catch {
      setError("No se pudo iniciar sesión.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-shell root">
      <form className="tarjeta login" onSubmit={onSubmit} autoComplete="on">
        <div className="login-logo">
          <img src={logoSrc} alt={nombre} width={320} />
        </div>
        <h1 className="titulo-modulo">Ingresar al sistema</h1>
        <p className="subtitulo">Ingrese con su usuario y contraseña para continuar.</p>
        <input
          className="text-field"
          type="email"
          name="email"
          autoComplete="username"
          placeholder="Correo electrónico"
          required
          value={form.usuario}
          onChange={(e) => setForm({ ...form, usuario: e.target.value })}
        />
        <input
          className="password-field"
          type="password"
          name="password"
          autoComplete="current-password"
          placeholder="Contraseña"
          required
          value={form.contrasena}
          onChange={(e) => setForm({ ...form, contrasena: e.target.value })}
        />
        <p
          className={error ? "login-mensaje mensaje-error" : "login-mensaje"}
          role="alert"
          aria-live="polite"
        >
          {error}
        </p>
        <button type="submit" className="boton-primario login-submit" disabled={loading}>
          {loading ? "Ingresando…" : "Ingresar"}
        </button>
      </form>
    </div>
  );
}