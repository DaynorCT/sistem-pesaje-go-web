import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { EstadoEtiqueta, TablaBtnIcono } from "./tablas-acciones";

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
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const { body } = await fetchApi<Proveedor[]>("/api/proveedores");
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
    setError("");
    setMsg("");
  };

  const abrirEditar = (p: Proveedor) => {
    setForm({
      ...p,
      nombre_empresa: p.nombre_empresa || "",
      descripcion: p.descripcion || "",
    });
    setError("");
    setMsg("");
  };

  const cerrarDialogo = () => setForm(null);

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!form) return;
    setError("");
    setMsg("");
    const method = form.id ? "PUT" : "POST";
    const res = await fetch("/api/proveedores", {
      method,
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(form),
    });
    const data = await res.json();
    if (!res.ok || !data.success) {
      setError(data.message || "Error al guardar");
      return;
    }
    setMsg(data.message);
    cerrarDialogo();
    await load();
  };

  const campos = form?.id ? 8 : 7;

  return (
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
      cantidadCampos={campos}
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
            <tr key={p.id}>
              <td>{nombreCompleto(p)}</td>
              <td>{p.ci || "—"}</td>
              <td>{p.telefono || "—"}</td>
              <td>{p.nombre_empresa || "—"}</td>
              <td>{p.descripcion || "—"}</td>
              <EstadoEtiqueta estado={p.estado} />
              <td className="tabla-celda-acciones">
                <TablaBtnIcono tip="Editar proveedor" act="editar" onClick={() => abrirEditar(p)} />
              </td>
            </tr>
          ))
        )
      }
      formulario={
        form ? (
          <>
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
            {form.id ? (
              <Campo etiqueta="Estado">
                <select
                  className="combo-web"
                  value={form.estado}
                  onChange={(e) => setForm({ ...form, estado: e.target.value })}
                >
                  <option value="ACTIVO">ACTIVO</option>
                  <option value="INACTIVO">INACTIVO</option>
                </select>
              </Campo>
            ) : null}
          </>
        ) : null
      }
    />
  );
}