import { FormEvent, useEffect, useRef, useState } from "react";

type Empresa = {
  id: number;
  nombre_empresa: string;
  actividad_economica: string;
  correo_electronico?: string;
  telefono?: string;
  logo?: string;
  estado: string;
};

export default function EmpresaPage() {
  const [lista, setLista] = useState<Empresa[]>([]);
  const [editing, setEditing] = useState<Empresa | null>(null);
  const [logoPreview, setLogoPreview] = useState("");
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);

  const load = async () => {
    const res = await fetch("/api/empresas");
    const data = await res.json();
    if (data.success) setLista(data.data || []);
  };

  useEffect(() => {
    load();
  }, []);

  const abrirNueva = () => {
    setEditing({
      id: 0,
      nombre_empresa: "",
      actividad_economica: "",
      correo_electronico: "",
      telefono: "",
      logo: "",
      estado: "ACTIVO",
    });
    setLogoPreview("");
  };

  const abrirEditar = (e: Empresa) => {
    setEditing(e);
    setLogoPreview(e.logo || "");
  };

  const onLogo = (ev: React.ChangeEvent<HTMLInputElement>) => {
    const file = ev.target.files?.[0];
    if (!file || file.size > 2 * 1024 * 1024) {
      setError("Logo máximo 2MB");
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setLogoPreview(reader.result as string);
    reader.readAsDataURL(file);
  };

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!editing) return;
    setError("");
    setMsg("");
    const payload = {
      id: editing.id || undefined,
      nombre_empresa: editing.nombre_empresa,
      actividad_economica: editing.actividad_economica,
      correo_electronico: editing.correo_electronico,
      telefono: editing.telefono,
      logo: logoPreview,
    };
    const method = editing.id ? "PUT" : "POST";
    const res = await fetch("/api/empresas", {
      method,
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(payload),
    });
    const data = await res.json();
    if (!data.success) {
      setError(data.message || "Error al guardar");
      return;
    }
    setMsg(data.message);
    setEditing(null);
    await load();
  };

  const activas = lista.filter((e) => e.estado === "ACTIVO");

  return (
    <section className="empresa-page">
      <div className="empresa-header">
        <h2>Empresa</h2>
        <button type="button" className="btn-secondary" onClick={abrirNueva}>
          Nueva empresa
        </button>
      </div>
      {msg && <p className="msg-ok">{msg}</p>}
      {error && <p className="login-error">{error}</p>}

      {editing ? (
        <form className="empresa-form" onSubmit={onSubmit}>
          <label>
            Nombre
            <input
              required
              value={editing.nombre_empresa}
              onChange={(e) => setEditing({ ...editing, nombre_empresa: e.target.value })}
            />
          </label>
          <label>
            Actividad económica
            <input
              required
              value={editing.actividad_economica}
              onChange={(e) => setEditing({ ...editing, actividad_economica: e.target.value })}
            />
          </label>
          <label>
            Correo
            <input
              type="email"
              value={editing.correo_electronico || ""}
              onChange={(e) => setEditing({ ...editing, correo_electronico: e.target.value })}
            />
          </label>
          <label>
            Teléfono
            <input
              value={editing.telefono || ""}
              onChange={(e) => setEditing({ ...editing, telefono: e.target.value })}
            />
          </label>
          <div className="logo-row">
            {logoPreview ? <img src={logoPreview} alt="Logo" className="logo-preview" /> : <div className="logo-preview empty">Sin logo</div>}
            <input ref={fileRef} type="file" accept="image/*" hidden onChange={onLogo} />
            <button type="button" className="btn-secondary" onClick={() => fileRef.current?.click()}>
              Subir logo
            </button>
          </div>
          <div className="form-actions">
            <button type="submit">Guardar</button>
            <button type="button" className="btn-secondary" onClick={() => setEditing(null)}>
              Cancelar
            </button>
          </div>
        </form>
      ) : (
        <table className="pesajes-table">
          <thead>
            <tr>
              <th>Nombre</th>
              <th>Actividad</th>
              <th>Estado</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {activas.length === 0 ? (
              <tr>
                <td colSpan={4}>No hay empresa registrada. Creá una para personalizar login y boletas.</td>
              </tr>
            ) : (
              activas.map((e) => (
                <tr key={e.id}>
                  <td>{e.nombre_empresa}</td>
                  <td>{e.actividad_economica}</td>
                  <td>{e.estado}</td>
                  <td>
                    <button type="button" className="btn-link" onClick={() => abrirEditar(e)}>
                      Editar
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      )}
    </section>
  );
}
