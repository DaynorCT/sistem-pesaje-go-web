import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { useConfirmar } from "./components/useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "./tablas-acciones";

type Proveedor = {
  id: number;
  nombre?: string;
  apellido_paterno?: string;
  apellido_materno?: string;
  ci?: string;
  telefono?: string;
  nombre_empresa?: string;
  descripcion?: string;
  estado: string;
};

type ProveedorForm = Proveedor;

function nombreCompleto(p: Proveedor): string {
  return [p.nombre, p.apellido_paterno, p.apellido_materno].filter(Boolean).join(" ").trim() || "—";
}

const vacio = (): ProveedorForm => ({
  id: 0,
  nombre: "",
  apellido_paterno: "",
  apellido_materno: "",
  ci: "",
  telefono: "",
  nombre_empresa: "",
  descripcion: "",
  estado: "ACTIVO",
});

export default function ProveedoresPage() {
  const [lista, setLista] = useState<Proveedor[]>([]);
  const [form, setForm] = useState<ProveedorForm | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [dialogError, setDialogError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  const load = async () => {
    setLoading(true);
    try {
      const { body, sessionExpired } = await fetchApi<Proveedor[]>("/api/proveedores");
      if (sessionExpired) return;
      if (body.success) setLista(body.data || []);
      else setError(body.message || "No se pudo cargar proveedores");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const abrirNuevo = () => {
    setForm(vacio());
    setDialogError("");
    setError("");
    setMsg("");
  };

  const abrirEditar = (p: Proveedor) => {
    setForm({
      ...p,
      nombre_empresa: p.nombre_empresa || "",
      descripcion: p.descripcion || "",
    });
    setDialogError("");
    setError("");
    setMsg("");
  };

  const cerrarDialogo = () => {
    setForm(null);
    setDialogError("");
  };

  const payloadDesdeProveedor = (p: Proveedor, estado: string) => ({
    id: p.id,
    nombre: p.nombre || "",
    apellido_paterno: p.apellido_paterno || "",
    apellido_materno: p.apellido_materno || "",
    ci: p.ci || "",
    telefono: p.telefono || "",
    nombre_empresa: p.nombre_empresa || "",
    descripcion: p.descripcion || "",
    estado,
  });

  const cambiarEstado = async (p: Proveedor, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar("Inactivar proveedor", "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    setError("");
    setMsg("");
    try {
      const { ok, sessionExpired, body: data } = await fetchApi("/api/proveedores", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadDesdeProveedor(p, nuevo)),
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
    setDialogError("");
    setError("");
    setMsg("");
    const estadoGuardar = form.id
      ? lista.find((x) => x.id === form.id)?.estado || form.estado
      : form.estado;
    const method = form.id ? "PUT" : "POST";
    const { ok, sessionExpired, body: data } = await fetchApi("/api/proveedores", {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, estado: estadoGuardar }),
    });
    if (sessionExpired) return;
    if (!ok || !data.success) {
      setDialogError(data.message || "Error al guardar");
      return;
    }
    setMsg(String(data.message || "Proveedor guardado"));
    cerrarDialogo();
    await load();
  };

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
        titulo="Proveedores"
        crearTexto="Crear proveedor"
        onCrear={abrirNuevo}
        mensaje={msg}
        mensajeError={error}
        loading={loading}
        dialogoAbierto={form !== null}
        onCerrarDialogo={cerrarDialogo}
        tituloDialogo={form?.id ? "Editar proveedor" : "Nuevo proveedor"}
        onSubmit={onSubmit}
        cantidadCampos={7}
        scrollDialogo
        cabeceraTabla={
          <>
            <th>Nombre</th>
            <th>CI</th>
            <th>Teléfono</th>
            <th>Empresa</th>
            <th>Descripción</th>
            <th className="tabla-th-estado">Estado</th>
            <th className="tabla-th-acciones">Acciones</th>
          </>
        }
        tabla={
          lista.length === 0 ? (
            <tr>
              <td colSpan={7}>No hay proveedores registrados.</td>
            </tr>
          ) : (
            lista.map((p) => (
              <tr key={p.id} className={form?.id === p.id ? "seleccionado" : undefined}>
                <td>{nombreCompleto(p)}</td>
                <td>{p.ci || "—"}</td>
                <td>{p.telefono || "—"}</td>
                <td>{p.nombre_empresa || "—"}</td>
                <td>{p.descripcion || "—"}</td>
                <EstadoEtiqueta estado={p.estado} />
                <td className="tabla-celda-acciones">
                  <TablaBtnIcono tip="Editar proveedor" act="editar" onClick={() => abrirEditar(p)} />
                  <EstadoSwitch
                    activo={estadoEsActivo(p.estado)}
                    onChange={(activo) => void cambiarEstado(p, activo)}
                  />
                </td>
              </tr>
            ))
          )
        }
        formulario={
          form ? (
            <>
              {dialogError ? (
                <p className="dlg-dialog-msg mensaje-error pesaje-campo" data-ancho-completo>
                  {dialogError}
                </p>
              ) : null}
              <Campo etiqueta="Nombre">
                <input
                  className="text-field"
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
              <Campo etiqueta="Empresa" requerido anchoCompleto>
                <input
                  className="text-field"
                  required
                  value={form.nombre_empresa || ""}
                  onChange={(e) => setForm({ ...form, nombre_empresa: e.target.value })}
                />
              </Campo>
              <Campo etiqueta="Descripción" anchoCompleto>
                <input
                  className="text-field"
                  value={form.descripcion || ""}
                  onChange={(e) => setForm({ ...form, descripcion: e.target.value })}
                />
              </Campo>
            </>
          ) : null
        }
      />
    </>
  );
}
