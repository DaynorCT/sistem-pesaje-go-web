import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "../api-client";
import { AbmModulo } from "./AbmModulo";
import { Campo } from "./Campo";
import { useConfirmar } from "./useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "../tablas-acciones";

type Fila = {
  id: number;
  nombre: string;
  descripcion?: string;
  estado: string;
};

type Props = {
  apiPath: string;
  titulo: string;
  crearTexto: string;
  tituloEntidad: string;
  etiquetaNombre?: string;
};

export function CatalogoSimpleAbm({
  apiPath,
  titulo,
  crearTexto,
  tituloEntidad,
  etiquetaNombre = "Nombre",
}: Props) {
  const [lista, setLista] = useState<Fila[]>([]);
  const [form, setForm] = useState<Fila | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [dialogError, setDialogError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  const load = async () => {
    setLoading(true);
    try {
      const { body, sessionExpired } = await fetchApi<Fila[]>(apiPath);
      if (sessionExpired) return;
      if (body.success) setLista(body.data || []);
      else setError(body.message || `No se pudo cargar ${titulo.toLowerCase()}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [apiPath]);

  const vacio = (): Fila => ({ id: 0, nombre: "", descripcion: "", estado: "ACTIVO" });

  const payload = (f: Fila, estado: string) => ({
    id: f.id,
    nombre: f.nombre,
    descripcion: f.descripcion || "",
    estado,
  });

  const cambiarEstado = async (f: Fila, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar(`Inactivar ${tituloEntidad}`, "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    setError("");
    setMsg("");
    const { ok, sessionExpired, body: data } = await fetchApi(apiPath, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload(f, nuevo)),
    });
    if (sessionExpired) return;
    if (!ok || !data.success) {
      setError(data.message || "No se pudo cambiar el estado");
      setLista((prev) => [...prev]);
      return;
    }
    setMsg(String(data.message || "Estado actualizado"));
    await load();
  };

  const onSubmit = async (ev: FormEvent) => {
    ev.preventDefault();
    if (!form) return;
    setDialogError("");
    setError("");
    setMsg("");
    const estadoGuardar = form.id ? lista.find((x) => x.id === form.id)?.estado || form.estado : form.estado;
    const method = form.id ? "PUT" : "POST";
    const { ok, sessionExpired, body: data } = await fetchApi(apiPath, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload(form, estadoGuardar)),
    });
    if (sessionExpired) return;
    if (!ok || !data.success) {
      setDialogError(data.message || "Error al guardar");
      return;
    }
    setMsg(String(data.message || "Guardado"));
    setForm(null);
    await load();
  };

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
        titulo={titulo}
        crearTexto={crearTexto}
        onCrear={() => {
          setForm(vacio());
          setDialogError("");
          setError("");
          setMsg("");
        }}
        mensaje={msg}
        mensajeError={error}
        loading={loading}
        dialogoAbierto={form !== null}
        onCerrarDialogo={() => {
          setForm(null);
          setDialogError("");
        }}
        tituloDialogo={form?.id ? `Editar ${tituloEntidad}` : `Nuevo ${tituloEntidad}`}
        onSubmit={onSubmit}
        cantidadCampos={2}
        columnasDialogo={2}
        anchoDialogo="medio"
        cabeceraTabla={
          <>
            <th>{etiquetaNombre}</th>
            <th>Descripción</th>
            <th className="tabla-th-estado">Estado</th>
            <th className="tabla-th-acciones">Acciones</th>
          </>
        }
        tabla={
          lista.length === 0 ? (
            <tr>
              <td colSpan={4}>No hay registros.</td>
            </tr>
          ) : (
            lista.map((f) => (
              <tr key={f.id} className={form?.id === f.id ? "seleccionado" : undefined}>
                <td title={f.nombre}>{f.nombre}</td>
                <td title={f.descripcion}>{f.descripcion || "—"}</td>
                <EstadoEtiqueta estado={f.estado} />
                <td className="tabla-celda-acciones">
                  <TablaBtnIcono
                    tip={`Editar ${tituloEntidad}`}
                    act="editar"
                    onClick={() => {
                      setForm({ ...f });
                      setDialogError("");
                      setError("");
                      setMsg("");
                    }}
                  />
                  <EstadoSwitch
                    activo={estadoEsActivo(f.estado)}
                    onChange={(activo) => void cambiarEstado(f, activo)}
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
              <Campo etiqueta={etiquetaNombre} requerido anchoCompleto>
                <input
                  className="text-field"
                  required
                  value={form.nombre}
                  onChange={(e) => setForm({ ...form, nombre: e.target.value })}
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
