import { FormEvent, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import { AbmModulo } from "./components/AbmModulo";
import { Campo } from "./components/Campo";
import { EstadoEtiqueta, TablaBtnIcono } from "./tablas-acciones";

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
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    try {
      const { body } = await fetchApi<Chofer[]>("/api/choferes");
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
    setError("");
    setMsg("");
  };

  const abrirEditar = (c: Chofer) => {
    setForm({ ...c, licencia: c.licencia || "" });
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
    const res = await fetch("/api/choferes", {
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

  const campos = form?.id ? 7 : 6;

  return (
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
      cantidadCampos={campos}
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
            <tr key={c.id}>
              <td>{nombreCompleto(c)}</td>
              <td>{c.ci || "—"}</td>
              <td>{c.telefono || "—"}</td>
              <td>{c.licencia || "—"}</td>
              <EstadoEtiqueta estado={c.estado} />
              <td className="tabla-celda-acciones">
                <TablaBtnIcono tip="Editar chofer" act="editar" onClick={() => abrirEditar(c)} />
              </td>
            </tr>
          ))
        )
      }
      formulario={
        form ? (
          <>
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