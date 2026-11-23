import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { useConfirmar } from "./components/useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "./tablas-acciones";

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
  const [dialogError, setDialogError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  function correoValido(v: string): boolean {
    if (!v.trim()) return true;
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.trim());
  }

  const load = async () => {
    setLoading(true);
    try {
      const { body, sessionExpired } = await fetchApi<Empresa[]>("/api/empresas");
      if (sessionExpired) return;
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
    setDialogError("");
    setError("");
    setMsg("");
  };

  const payloadDesdeEmpresa = (e: Empresa, estado: string) => ({
    id: e.id,
    nombre_empresa: e.nombre_empresa,
    actividad_economica: e.actividad_economica || "",
    correo_electronico: e.correo_electronico || "",
    telefono: e.telefono || "",
    estado,
    ...(e.logo ? { logo: e.logo } : {}),
  });

  const cambiarEstado = async (e: Empresa, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar("Inactivar empresa", "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    setError("");
    setMsg("");
    try {
      const { ok, sessionExpired, body: data } = await fetchApi("/api/empresas", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadDesdeEmpresa(e, nuevo)),
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

  const abrirEditar = (e: Empresa) => {
    setForm({
      id: Number(e.id),
      nombre_empresa: e.nombre_empresa ?? "",
      actividad_economica: e.actividad_economica ?? "",
      correo_electronico: e.correo_electronico ?? "",
      telefono: e.telefono ?? "",
      logo: e.logo ?? "",
      estado: e.estado || "ACTIVO",
    });
    setDialogError("");
    setError("");
    setMsg("");
  };

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!form) return;
    setDialogError("");
    setError("");
    setMsg("");

    const nombre_empresa = (form.nombre_empresa || "").trim();
    const actividad_economica = (form.actividad_economica || "").trim();
    const correo_electronico = (form.correo_electronico || "").trim();
    const telefono = (form.telefono || "").trim();
    const id = Number(form.id);
    const esEdicion = id > 0;

    if (!nombre_empresa) {
      setDialogError("El nombre es obligatorio.");
      return;
    }
    if (!correoValido(correo_electronico)) {
      setDialogError("El correo no es válido. Dejalo vacío o usá un formato tipo nombre@empresa.com");
      return;
    }

    const payload: Record<string, unknown> = {
      nombre_empresa,
      actividad_economica,
      correo_electronico,
      telefono,
    };
    if (esEdicion) {
      payload.id = id;
      payload.estado = lista.find((x) => x.id === id)?.estado || form.estado || "ACTIVO";
      if (form.logo) payload.logo = form.logo;
    }
    const method = esEdicion ? "PUT" : "POST";
    try {
      const { ok, status, sessionExpired, body: data } = await fetchApi("/api/empresas", {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (sessionExpired) return;
      if (!ok || !data.success) {
        const msgErr =
          status === 401
            ? "Sesión expirada o no autorizado. Cerrá sesión e ingresá de nuevo."
            : data.message || "Error al guardar";
        setDialogError(msgErr);
        return;
      }
      setMsg(String(data.message || "Empresa guardada"));
      setForm(null);
      await load();
    } catch {
      setDialogError("Error al guardar");
    }
  };

  const tieneActiva = lista.some((e) => e.estado === "ACTIVO");

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
      titulo="Empresa"
      crearTexto="Nueva empresa"
      onCrear={abrirNueva}
      mensaje={msg}
      mensajeError={error}
      loading={loading}
      dialogoAbierto={form !== null}
      onCerrarDialogo={() => {
        setForm(null);
        setDialogError("");
      }}
      tituloDialogo={form?.id ? "Editar empresa" : "Nueva empresa"}
      onSubmit={onSubmit}
      cantidadCampos={4}
      columnasDialogo={2}
      anchoDialogo="medio"
      cabeceraTabla={
        <>
          <th>Nombre</th>
          <th>Actividad</th>
          <th>Correo</th>
          <th>Teléfono</th>
          <th className="tabla-th-estado">Estado</th>
          <th className="tabla-th-acciones">Acciones</th>
        </>
      }
      tabla={
        lista.length === 0 ? (
          <tr>
            <td colSpan={6}>No hay empresa registrada. Creá una para personalizar login y boletas.</td>
          </tr>
        ) : (
          lista.map((e) => (
            <tr key={e.id} className={form?.id === e.id ? "seleccionado" : undefined}>
              <td title={e.nombre_empresa}>{e.nombre_empresa}</td>
              <td title={e.actividad_economica}>{e.actividad_economica || "—"}</td>
              <td title={e.correo_electronico}>{e.correo_electronico || "—"}</td>
              <td title={e.telefono}>{e.telefono || "—"}</td>
              <EstadoEtiqueta estado={e.estado} />
              <td className="tabla-celda-acciones">
                <TablaBtnIcono tip="Editar empresa" act="editar" onClick={() => abrirEditar(e)} />
                <EstadoSwitch
                  activo={estadoEsActivo(e.estado)}
                  onChange={(activo) => void cambiarEstado(e, activo)}
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
            {!form.id && tieneActiva ? (
              <p className="dlg-dialog-msg subtitulo pesaje-campo" data-ancho-completo>
                Ya hay una empresa activa. Podés crear otra o editar la existente desde la grilla.
              </p>
            ) : null}
            <Campo etiqueta="Nombre" requerido>
              <input
                className="text-field"
                value={form.nombre_empresa}
                onChange={(e) => setForm({ ...form, nombre_empresa: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Actividad">
              <input
                className="text-field"
                value={form.actividad_economica}
                onChange={(e) => setForm({ ...form, actividad_economica: e.target.value })}
              />
            </Campo>
            <Campo etiqueta="Correo">
              <input
                className="text-field"
                type="text"
                inputMode="email"
                autoComplete="email"
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
    </>
  );
}