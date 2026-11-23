import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { useConfirmar } from "./components/useConfirmar";
import { estadoEsActivo, EstadoEtiqueta, EstadoSwitch, TablaBtnIcono } from "./tablas-acciones";

type Vehiculo = {
  id: number;
  placa: string;
  tipo_vehiculo?: string;
  tara: number;
  estado: string;
};

export default function VehiculosPage() {
  const [lista, setLista] = useState<Vehiculo[]>([]);
  const [form, setForm] = useState<Vehiculo | null>(null);
  const [msg, setMsg] = useState("");
  const [error, setError] = useState("");
  const [dialogError, setDialogError] = useState("");
  const [loading, setLoading] = useState(true);
  const { preguntar, dialogo: dialogoConfirmar } = useConfirmar();

  const load = async () => {
    setLoading(true);
    try {
      const { body, sessionExpired } = await fetchApi<Vehiculo[]>("/api/vehiculos");
      if (sessionExpired) return;
      if (body.success) setLista(body.data || []);
      else setError(body.message || "No se pudo cargar vehículos");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const vacio = (): Vehiculo => ({ id: 0, placa: "", tipo_vehiculo: "", tara: 0, estado: "ACTIVO" });

  const payload = (v: Vehiculo, estado: string) => ({
    id: v.id,
    placa: v.placa,
    tipo_vehiculo: v.tipo_vehiculo || "",
    tara: v.tara ?? 0,
    estado,
  });

  const cambiarEstado = async (v: Vehiculo, activo: boolean) => {
    const nuevo = activo ? "ACTIVO" : "INACTIVO";
    if (
      !activo &&
      !(await preguntar("Inactivar vehículo", "Se cambiará el estado a INACTIVO. ¿Continuar?", {
        confirmar: "INACTIVAR",
      }))
    ) {
      setLista((prev) => [...prev]);
      return;
    }
    const { ok, sessionExpired, body: data } = await fetchApi("/api/vehiculos", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload(v, nuevo)),
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
    const estadoGuardar = form.id ? lista.find((x) => x.id === form.id)?.estado || form.estado : form.estado;
    const method = form.id ? "PUT" : "POST";
    const { ok, sessionExpired, body: data } = await fetchApi("/api/vehiculos", {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload(form, estadoGuardar)),
    });
    if (sessionExpired) return;
    if (!ok || !data.success) {
      setDialogError(data.message || "Error al guardar");
      return;
    }
    setMsg(String(data.message || "Vehículo guardado"));
    setForm(null);
    await load();
  };

  return (
    <>
      {dialogoConfirmar}
      <AbmModulo
        titulo="Vehículos"
        crearTexto="Crear vehículo"
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
        tituloDialogo={form?.id ? "Editar vehículo" : "Nuevo vehículo"}
        onSubmit={onSubmit}
        cantidadCampos={3}
        columnasDialogo={2}
        anchoDialogo="medio"
        cabeceraTabla={
          <>
            <th>Placa</th>
            <th>Tipo</th>
            <th>Tara</th>
            <th className="tabla-th-estado">Estado</th>
            <th className="tabla-th-acciones">Acciones</th>
          </>
        }
        tabla={
          lista.length === 0 ? (
            <tr>
              <td colSpan={5}>No hay vehículos registrados.</td>
            </tr>
          ) : (
            lista.map((v) => (
              <tr key={v.id} className={form?.id === v.id ? "seleccionado" : undefined}>
                <td>{v.placa}</td>
                <td>{v.tipo_vehiculo || "—"}</td>
                <td>{v.tara ?? 0}</td>
                <EstadoEtiqueta estado={v.estado} />
                <td className="tabla-celda-acciones">
                  <TablaBtnIcono
                    tip="Editar vehículo"
                    act="editar"
                    onClick={() => {
                      setForm({ ...v });
                      setDialogError("");
                      setMsg("");
                      setError("");
                    }}
                  />
                  <EstadoSwitch
                    activo={estadoEsActivo(v.estado)}
                    onChange={(activo) => void cambiarEstado(v, activo)}
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
              <Campo etiqueta="Placa" requerido>
                <input
                  className="text-field"
                  required
                  value={form.placa}
                  onChange={(e) => setForm({ ...form, placa: e.target.value })}
                />
              </Campo>
              <Campo etiqueta="Tipo">
                <input
                  className="text-field"
                  value={form.tipo_vehiculo || ""}
                  onChange={(e) => setForm({ ...form, tipo_vehiculo: e.target.value })}
                />
              </Campo>
              <Campo etiqueta="Tara (kg)">
                <input
                  className="text-field"
                  inputMode="numeric"
                  value={String(form.tara ?? 0)}
                  onChange={(e) => setForm({ ...form, tara: parseInt(e.target.value, 10) || 0 })}
                />
              </Campo>
            </>
          ) : null
        }
      />
    </>
  );
}
