import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { EstadoEtiqueta, TablaBtnIcono } from "./tablas-acciones";

type Empresa = {
  id: number;
  nombre_empresa: string;
  actividad_economica: string;
  correo_electronico?: string;
  telefono?: string;
  logo?: string;
  estado: string;
};

const vacia = (): Empresa => ({
  id: 0,
  nombre_empresa: "",
  actividad_economica: "",
  correo_electronico: "",
  telefono: "",
  logo: "",
  estado: "ACTIVO",
});

export default function EmpresaPage() {
  const [lista, setLista] = useState<Empresa[]>([]);
  const [form, setForm] = useState<Empresa | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const { body } = await fetchApi<Empresa[]>("/api/empresas");
      if (body.success) setLista(body.data || []);
      else setError(body.message || "No se pudo cargar empresas");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const abrirNueva = () => {
    setForm(vacia());
    setError("");
    setMsg("");
  };

  const abrirEditar = (e: Empresa) => {
    setForm(e);
    setError("");
    setMsg("");
  };

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!form) return;
    setError("");
    setMsg("");
    const payload = {
      id: form.id || undefined,
      nombre_empresa: form.nombre_empresa,
      actividad_economica: form.actividad_economica,
      correo_electronico: form.correo_electronico,
      telefono: form.telefono,
      logo: form.logo || "",
    };
    const method = form.id ? "PUT" : "POST";
    try {
      const { ok, body: data } = await fetchApi("/api/empresas", {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!ok || !data.success) {
        setError(data.message || "Error al guardar");
        return;
      }
      setMsg(String(data.message || "Empresa guardada"));
      setForm(null);
      await load();
    } catch {
      setError("Error al guardar");
    }
  };

  const activas = lista.filter((e) => e.estado === "ACTIVO");

  return (
    <AbmModulo
      titulo="Empresa"
      crearTexto="Nueva empresa"
      onCrear={abrirNueva}
      mensaje={msg}
      mensajeError={error}
      loading={loading}
      dialogoAbierto={form !== null}
      onCerrarDialogo={() => setForm(null)}
      tituloDialogo={form?.id ? "Editar empresa" : "Nueva empresa"}
      onSubmit={onSubmit}
      cantidadCampos={4}
      columnasDialogo={2}
      anchoDialogo="medio"
      cabeceraTabla={
        <>
          <th>Nombre</th>
          <th>Actividad</th>
          <th className="tabla-th-estado">Estado</th>
          <th className="tabla-th-acciones">Acciones</th>
        </>
      }
      tabla={
        activas.length === 0 ? (
          <tr>
            <td colSpan={4}>No hay empresa registrada. Creá una para personalizar login y boletas.</td>
          </tr>
        ) : (
          activas.map((e) => (
            <tr key={e.id}>
              <td>{e.nombre_empresa}</td>
              <td>{e.actividad_economica}</td>
              <EstadoEtiqueta estado={e.estado} />
              <td className="tabla-celda-acciones">
                <TablaBtnIcono tip="Editar empresa" act="editar" onClick={() => abrirEditar(e)} />
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
                value={form.nombre_empresa}
                onChange={(e) => setForm({ ...form, nombre_empresa: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Actividad" requerido>
              <input
                className="text-field"
                required
                value={form.actividad_economica}
                onChange={(e) => setForm({ ...form, actividad_economica: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Correo">
              <input
                className="text-field"
                type="email"
                value={form.correo_electronico || ""}
                onChange={(e) => setForm({ ...form, correo_electronico: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Teléfono">
              <input
                className="text-field"
                value={form.telefono || ""}
                onChange={(e) => setForm({ ...form, telefono: e.target.value })}
              />
            </Campo>
          </>
        ) : null
      }
    />
  );
}