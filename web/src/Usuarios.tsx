import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { useConfirmar } from "./components/useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "./tablas-acciones";

type Usuario = {
  id: number;
  ci?: string;
  nombre_completo?: string;
  correo?: string;
  telefono?: string;
  rol?: string;
  estado: string;
  nombre?: string;
  apellido_paterno?: string;
  apellido_materno?: string;
};

type UsuarioForm = Usuario & { clave?: string };

const ROLES = ["usuario", "administrador"];

function usuarioSesionId(): number {
  try {
    const raw = localStorage.getItem("pesaje_user");
    if (!raw) return 0;
    const u = JSON.parse(raw) as { id?: number };
    return Number(u.id) || 0;
  } catch {
    return 0;
  }
}

const vacio = (): UsuarioForm => ({
  id: 0,
  nombre: "",
  apellido_paterno: "",
  apellido_materno: "",
  ci: "",
  telefono: "",
  correo: "",
  clave: "",
  rol: "usuario",
  estado: "ACTIVO",
});

export default function UsuariosPage() {
  const [lista, setLista] = useState<Usuario[]>([]);
  const [form, setForm] = useState<UsuarioForm | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  const load = async () => {
    setLoading(true);
    try {
      const { body } = await fetchApi<Usuario[]>("/api/usuarios");
      if (body.success) setLista(body.data || []);
      else setError(body.message || "No se pudo cargar usuarios");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const abrirNuevo = () => {
    setForm(vacio());
    setError("");
    setMsg("");
  };

  const abrirEditar = (u: Usuario) => {
    setForm({
      ...u,
      nombre: u.nombre || "",
      apellido_paterno: u.apellido_paterno || "",
      apellido_materno: u.apellido_materno || "",
      ci: u.ci || "",
      telefono: u.telefono || "",
      correo: u.correo || "",
      rol: u.rol || "usuario",
      clave: "",
    });
    setError("");
    setMsg("");
  };

  const cerrarDialogo = () => setForm(null);

  const payloadDesdeUsuario = (u: Usuario, estado: string) => ({
    id: u.id,
    nombre: u.nombre || "",
    apellido_paterno: u.apellido_paterno || "",
    apellido_materno: u.apellido_materno || "",
    ci: u.ci || "",
    telefono: u.telefono || "",
    correo: u.correo || "",
    rol: u.rol || "usuario",
    estado,
  });

  const cambiarEstado = async (u: Usuario, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar("Inactivar usuario", "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    setError("");
    setMsg("");
    try {
      const { ok, sessionExpired, body: data } = await fetchApi("/api/usuarios", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadDesdeUsuario(u, nuevo)),
      });
      if (sessionExpired) return;
      if (!ok || !data.success) {
        setError(data.message || "No se pudo cambiar el estado");
        setLista((prev) => [...prev]);
        return;
      }
      setMsg(String(data.message || "Estado actualizado"));
      await load();
    } catch {
      setError("No se pudo cambiar el estado");
      setLista((prev) => [...prev]);
    }
  };

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!form) return;
    setError("");
    setMsg("");
    const estadoGuardar = form.id
      ? lista.find((x) => x.id === form.id)?.estado || form.estado
      : form.estado;
    const payload = {
      id: form.id || undefined,
      nombre: form.nombre,
      apellido_paterno: form.apellido_paterno,
      apellido_materno: form.apellido_materno,
      ci: form.ci,
      telefono: form.telefono,
      correo: form.correo,
      rol: form.rol,
      estado: estadoGuardar,
      clave: form.clave || undefined,
    };
    const method = form.id ? "PUT" : "POST";
    try {
      const { ok, body: data } = await fetchApi("/api/usuarios", {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!ok || !data.success) {
        setError(data.message || "Error al guardar");
        return;
      }
      setMsg(String(data.message || "Usuario guardado"));
      cerrarDialogo();
      await load();
    } catch {
      setError("Error al guardar");
    }
  };

  const esEdicion = Boolean(form?.id);
  const camposVisibles = 8;

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
      titulo="Usuarios"
      crearTexto="Crear usuario"
      onCrear={abrirNuevo}
      mensaje={msg}
      mensajeError={error}
      loading={loading}
      dialogoAbierto={form !== null}
      onCerrarDialogo={cerrarDialogo}
      tituloDialogo={esEdicion ? "Editar usuario" : "Nuevo usuario"}
      onSubmit={onSubmit}
      cantidadCampos={camposVisibles}
      columnasDialogo={3}
      anchoDialogo="amplio"
      scrollDialogo
      cabeceraTabla={
        <>
          <th>Nro. documento</th>
          <th>Nombres</th>
          <th>Correo electrónico</th>
          <th>Teléfono</th>
          <th>Rol</th>
          <th className="tabla-th-estado">Estado</th>
          <th className="tabla-th-acciones">Acciones</th>
        </>
      }
      tabla={
        lista.length === 0 ? (
          <tr>
            <td colSpan={7}>No hay usuarios registrados.</td>
          </tr>
        ) : (
          lista.map((u) => (
            <tr key={u.id}>
              <td>{u.ci || "—"}</td>
              <td>{u.nombre_completo || u.nombre || "—"}</td>
              <td>{u.correo || "—"}</td>
              <td>{u.telefono || "—"}</td>
              <td>{u.rol || "—"}</td>
              <EstadoEtiqueta estado={u.estado} />
              <td className="tabla-celda-acciones">
                <TablaBtnIcono tip="Editar usuario" act="editar" onClick={() => abrirEditar(u)} />
                <EstadoSwitch
                  activo={estadoEsActivo(u.estado)}
                  disabled={u.id === usuarioSesionId()}
                  onChange={(activo) => void cambiarEstado(u, activo)}
                />
              </td>
            </tr>
          ))
        )
      }
      formulario={
        form ? (
          <>
            {error ? (
              <p className="dlg-dialog-msg mensaje-error pesaje-campo" data-ancho-completo>
                {error}
              </p>
            ) : null}
            <Campo etiqueta="Nombre" requerido>
              <input
                className="text-field"
                required
                value={form.nombre || ""}
                onChange={(e) => setForm({ ...form, nombre: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Ap. paterno">
              <input
                className="text-field"
                value={form.apellido_paterno || ""}
                onChange={(e) => setForm({ ...form, apellido_paterno: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Ap. materno">
              <input
                className="text-field"
                value={form.apellido_materno || ""}
                onChange={(e) => setForm({ ...form, apellido_materno: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="CI">
              <input
                className="text-field"
                value={form.ci || ""}
                onChange={(e) => setForm({ ...form, ci: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Teléfono">
              <input
                className="text-field"
                value={form.telefono || ""}
                onChange={(e) => setForm({ ...form, telefono: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Correo" requerido>
              <input
                className="text-field"
                type="email"
                required
                value={form.correo || ""}
                onChange={(e) => setForm({ ...form, correo: e.target.value })}
              />
            </Campo>
            <Campo
              etiqueta="Contraseña"
              requerido={!esEdicion}
            >
              <input
                className="password-field"
                type="password"
                autoComplete="new-password"
                required={!esEdicion}
                value={form.clave || ""}
                onChange={(e) => setForm({ ...form, clave: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Rol">
              <select
                className="combo-web"
                value={form.rol || "usuario"}
                onChange={(e) => setForm({ ...form, rol: e.target.value })}
              >
                {ROLES.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </Campo>
          </>
        ) : null
      }
    />
    </>
  );
}