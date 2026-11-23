import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { useConfirmar } from "./components/useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "./tablas-acciones";

type Chofer = {
  id: number;
  nombre?: string;
  apellido_paterno?: string;
  apellido_materno?: string;
  ci?: string;
  telefono?: string;
  licencia?: string;
  estado: string;
};

type ChoferForm = Chofer;

function nombreCompleto(c: Chofer): string {
  return [c.nombre, c.apellido_paterno, c.apellido_materno].filter(Boolean).join(" ").trim() || "—";
}

const vacio = (): ChoferForm => ({
  id: 0,
  nombre: "",
  apellido_paterno: "",
  apellido_materno: "",
  ci: "",
  telefono: "",
  licencia: "",
  estado: "ACTIVO",
});

export default function ChoferesPage() {
  const [lista, setLista] = useState<Chofer[]>([]);
  const [form, setForm] = useState<ChoferForm | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [dialogError, setDialogError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  const load = async () => {
    setLoading(true);
    try {
      const { body, sessionExpired } = await fetchApi<Chofer[]>("/api/choferes");
      if (sessionExpired) return;
      if (body.success) setLista(body.data || []);
      else setError(body.message || "No se pudo cargar choferes");
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

  const abrirEditar = (c: Chofer) => {
    setForm({ ...c, licencia: c.licencia || "" });
    setDialogError("");
    setError("");
    setMsg("");
  };

  const cerrarDialogo = () => {
    setForm(null);
    setDialogError("");
  };

  const payloadDesdeChofer = (c: Chofer, estado: string) => ({
    id: c.id,
    nombre: c.nombre || "",
    apellido_paterno: c.apellido_paterno || "",
    apellido_materno: c.apellido_materno || "",
    ci: c.ci || "",
    telefono: c.telefono || "",
    licencia: c.licencia || "",
    estado,
  });

  const cambiarEstado = async (c: Chofer, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar("Inactivar chofer", "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    setError("");
    setMsg("");
    try {
      const { ok, sessionExpired, body: data } = await fetchApi("/api/choferes", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadDesdeChofer(c, nuevo)),
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
    const { ok, sessionExpired, body: data } = await fetchApi("/api/choferes", {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, estado: estadoGuardar }),
    });
    if (sessionExpired) return;
    if (!ok || !data.success) {
      setDialogError(data.message || "Error al guardar");
      return;
    }
    setMsg(String(data.message || "Chofer guardado"));
    cerrarDialogo();
    await load();
  };

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
        titulo="Choferes"
        crearTexto="Crear chofer"
        onCrear={abrirNuevo}
        mensaje={msg}
        mensajeError={error}
        loading={loading}
        dialogoAbierto={form !== null}
        onCerrarDialogo={cerrarDialogo}
        tituloDialogo={form?.id ? "Editar chofer" : "Nuevo chofer"}
        onSubmit={onSubmit}
        cantidadCampos={6}
        scrollDialogo
        cabeceraTabla={
          <>
            <th>Nombre</th>
            <th>CI</th>
            <th>Teléfono</th>
            <th>Licencia</th>
            <th className="tabla-th-estado">Estado</th>
            <th className="tabla-th-acciones">Acciones</th>
          </>
        }
        tabla={
          lista.length === 0 ? (
            <tr>
              <td colSpan={6}>No hay choferes registrados.</td>
            </tr>
          ) : (
            lista.map((c) => (
              <tr key={c.id} className={form?.id === c.id ? "seleccionado" : undefined}>
                <td>{nombreCompleto(c)}</td>
                <td>{c.ci || "—"}</td>
                <td>{c.telefono || "—"}</td>
                <td>{c.licencia || "—"}</td>
                <EstadoEtiqueta estado={c.estado} />
                <td className="tabla-celda-acciones">
                  <TablaBtnIcono tip="Editar chofer" act="editar" onClick={() => abrirEditar(c)} />
                  <EstadoSwitch
                    activo={estadoEsActivo(c.estado)}
                    onChange={(activo) => void cambiarEstado(c, activo)}
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
              <Campo etiqueta="Licencia" requerido>
                <input
                  className="text-field"
                  required
                  value={form.licencia || ""}
                  onChange={(e) => setForm({ ...form, licencia: e.target.value })}
                />
              </Campo>
            </>
          ) : null
        }
      />
    </>
  );
}
