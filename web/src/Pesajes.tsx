import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { fetchApi } from "./api-client";
import ModuloPanel from "./components/ModuloPanel";
import { IconoModulo } from "./iconos-modulo";
import { motivoNoFinalizar, type PesajeFila } from "./pesajes-logica";
import { EstadoEtiqueta, TablaBtnIcono } from "./tablas-acciones";
import "./css/pesaje.css";

type Opcion = { id: number; label: string };

type OpcionesPesaje = {
  vehiculos: Opcion[];
  choferes: Opcion[];
  proveedores: Opcion[];
  productos: Opcion[];
  origenes: Opcion[];
  destinos: Opcion[];
};

type FilaGrilla = PesajeFila & {
  guia?: string;
  nombre_chofer?: string;
  placa?: string;
  licencia?: string;
  tipo_vehiculo?: string;
  nombre_proveedor?: string;
  producto?: string;
  origen?: string;
  destino?: string;
  costo_bs?: number;
  flete_bs_pendiente?: number;
  fecha_creacion?: string;
  estado?: string;
};

const opcionesVacias: OpcionesPesaje = {
  vehiculos: [],
  choferes: [],
  proveedores: [],
  productos: [],
  origenes: [],
  destinos: [],
};

function fechaHora(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso.length > 16 ? iso.slice(0, 16) : iso;
  const fecha = d.toLocaleDateString("es-BO");
  const hora = d.toLocaleTimeString("es-BO", { hour: "2-digit", minute: "2-digit" });
  return `${fecha} ${hora}`.trim();
}

function mayus(s?: string): string {
  return (s || "—").toUpperCase();
}

function etiquetaBalanza(estadoBalanza?: string): string {
  return (estadoBalanza || "—").toUpperCase().replace(/_/g, " ");
}

function ComboCampo({
  etiqueta,
  requerido,
  value,
  onChange,
  items,
  plusTo,
}: {
  etiqueta: string;
  requerido?: boolean;
  value: string;
  onChange: (v: string) => void;
  items: Opcion[];
  plusTo?: string;
}) {
  return (
    <label className="pesaje-campo">
      <span className="etiqueta-mini">
        {etiqueta}
        {requerido ? " *" : ""}
      </span>
      <span className="pesaje-fila-combo">
        <select className="combo-web" value={value} onChange={(e) => onChange(e.target.value)}>
          <option value="">—</option>
          {items.map((o) => (
            <option key={o.id} value={String(o.id)}>
              {o.label || `#${o.id}`}
            </option>
          ))}
        </select>
        {plusTo ? (
          <Link className="boton-mas" to={plusTo} title="Crear nuevo">
            +
          </Link>
        ) : null}
      </span>
    </label>
  );
}

export default function Pesajes() {
  const [opciones, setOpciones] = useState<OpcionesPesaje>(opcionesVacias);
  const [filas, setFilas] = useState<FilaGrilla[]>([]);
  const [loading, setLoading] = useState(true);

  const [pesoVivo, setPesoVivo] = useState(0);
  const [conectado, setConectado] = useState(false);
  const [pesoBruto, setPesoBruto] = useState(0);
  const [pesoTara, setPesoTara] = useState(0);
  const [taraManual, setTaraManual] = useState(false);
  const [taraKg, setTaraKg] = useState("0");

  const [choferId, setChoferId] = useState("");
  const [vehiculoId, setVehiculoId] = useState("");
  const [proveedorId, setProveedorId] = useState("");
  const [productoId, setProductoId] = useState("");
  const [origenId, setOrigenId] = useState("");
  const [destinoId, setDestinoId] = useState("");
  const [licencia, setLicencia] = useState("");
  const [tipoVehiculo, setTipoVehiculo] = useState("");
  const [costo, setCosto] = useState("0");
  const [flete, setFlete] = useState("0");

  const [mensaje, setMensaje] = useState("");
  const [error, setError] = useState("");

  const pesoNeto = useMemo(() => Math.max(0, pesoBruto - pesoTara), [pesoBruto, pesoTara]);
  const puedeCapturar = pesoVivo > 0;

  const refreshGrilla = useCallback(async () => {
    setLoading(true);
    try {
      const { body } = await fetchApi<FilaGrilla[]>("/api/pesajes/grilla");
      if (body.success && Array.isArray(body.data)) setFilas(body.data);
    } finally {
      setLoading(false);
    }
  }, []);

  const cargarOpciones = useCallback(async () => {
    const { body } = await fetchApi<OpcionesPesaje>("/api/pesajes/opciones");
    if (body.success && body.data) setOpciones(body.data as OpcionesPesaje);
  }, []);

  useEffect(() => {
    cargarOpciones();
    refreshGrilla();
    const poll = setInterval(() => {
      fetch("/api/pesajes/ultimo-peso")
        .then((r) => r.json())
        .then((d) => {
          if (d.success) setPesoVivo(Number(d.peso) || 0);
        })
        .catch(() => {});
    }, 1500);
    return () => clearInterval(poll);
  }, [cargarOpciones, refreshGrilla]);

  useEffect(() => {
    if (!choferId) {
      setLicencia("");
      return;
    }
    fetchApi<{ licencia: string }>(`/api/pesajes/lookup/chofer/${choferId}`).then(({ body }) => {
      if (body.success && body.data) setLicencia(String((body.data as { licencia?: string }).licencia || ""));
    });
  }, [choferId]);

  useEffect(() => {
    if (!vehiculoId) {
      setTipoVehiculo("");
      return;
    }
    fetchApi<{ tipo: string; tara: number }>(`/api/pesajes/lookup/vehiculo/${vehiculoId}`).then(({ body }) => {
      if (!body.success || !body.data) return;
      const d = body.data as { tipo?: string; tara?: number };
      setTipoVehiculo(d.tipo || "");
      if (!taraManual && d.tara != null) {
        const t = String(d.tara);
        setTaraKg(t);
        setPesoTara(Number(d.tara) || 0);
      }
    });
  }, [vehiculoId, taraManual]);

  useEffect(() => {
    if (taraManual) {
      setPesoTara(Number(taraKg) || 0);
    }
  }, [taraManual, taraKg]);

  function capturarBruto() {
    setPesoBruto(pesoVivo);
    setMensaje("Peso bruto capturado desde planta.");
    setError("");
  }

  function capturarTara() {
    setPesoTara(pesoVivo);
    setTaraKg(String(pesoVivo));
    setTaraManual(true);
    setMensaje("Tara capturada desde planta.");
    setError("");
  }

  async function registrarPesaje() {
    setError("");
    setMensaje("");
    if (!vehiculoId) {
      setError("Seleccioná la placa (vehículo).");
      return;
    }
    try {
      const { ok, body } = await fetchApi("/api/pesajes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          vehiculo_id: Number(vehiculoId),
          chofer_id: choferId ? Number(choferId) : undefined,
          proveedor_id: proveedorId ? Number(proveedorId) : undefined,
          producto_id: productoId ? Number(productoId) : undefined,
          id_origen: origenId ? Number(origenId) : undefined,
          id_destino: destinoId ? Number(destinoId) : undefined,
          peso_bruto: pesoBruto,
          tara: pesoTara,
          costo_bs: Number(costo) || 0,
          flete_bs_pendiente: Number(flete) || 0,
          estado_balanza: "EN_PROCESO",
        }),
      });
      if (!ok || !body.success) {
        setError(body.message || "No se pudo registrar");
        return;
      }
      setMensaje(String(body.message || "Pesaje registrado"));
      setPesoBruto(0);
      setPesoTara(0);
      setTaraKg("0");
      setTaraManual(false);
      await refreshGrilla();
    } catch {
      setError("No se pudo registrar");
    }
  }

  async function finalizarPesaje(p: FilaGrilla) {
    const bloqueo = motivoNoFinalizar(p);
    if (bloqueo) {
      setError(bloqueo);
      return;
    }
    const ok = window.confirm(`¿Finalizar pesaje #${p.id}?`);
    if (!ok) return;
    const res = await fetch(`/api/pesajes/${p.id}/finalizar`, { method: "POST", credentials: "include" });
    const data = await res.json();
    if (!res.ok || !data.success) {
      setError(data.message || "No se pudo finalizar");
      return;
    }
    setMensaje(data.message || "Pesaje finalizado");
    await refreshGrilla();
  }

  return (
    <ModuloPanel titulo="Pesaje">
      <div className="panel-pesaje">
        <div className="panel-pesaje-medio">
          <section className="tarjeta pesaje-registro">
            <h2 className="titulo-tarjeta">Registro de peso</h2>
            <div className="separador-horizontal" />
            <div className="display-peso">
              <span className="valor-peso">{pesoVivo.toLocaleString("es-BO")} kg</span>
            </div>
            <div className="separador-horizontal" />
            <div className="pesaje-acciones-balanza">
              <div className={`pesaje-col-conexion${conectado ? " pesaje-col-conexion--conectada" : ""}`}>
                <div className="pesaje-conexion-fila">
                  <span className="pesaje-icono-balanza">
                    <IconoModulo id="pesaje" />
                  </span>
                  <label className="abm-switch pesaje-switch-balanza" data-tip="Simular conexión (web)">
                    <input
                      type="checkbox"
                      className="abm-switch-input"
                      checked={conectado}
                      onChange={(e) => setConectado(e.target.checked)}
                      aria-label="Conexión balanza"
                    />
                    <span className="abm-switch-pista" />
                  </label>
                </div>
                <span className="etiqueta-mini pesaje-leyenda-conexion">Conexión</span>
                <span
                  className={`estado-etiqueta ${conectado ? "estado-etiqueta--activo" : "estado-etiqueta--inactivo"}`}
                >
                  {conectado ? "CONECTADO" : "DESCONECTADO"}
                </span>
              </div>
              <button type="button" className="boton-primario" disabled={!puedeCapturar} onClick={capturarBruto}>
                Cap. peso
              </button>
              <button type="button" className="boton-info" disabled={!puedeCapturar} onClick={capturarTara}>
                Cap. tara
              </button>
            </div>
            <div className="pesaje-resumen">
              <div className="caja-peso-resumen">
                <span className="etiqueta-mini">Peso Bruto</span>
                <strong className="valor-caja-peso">{pesoBruto}</strong>
              </div>
              <div className="caja-peso-resumen">
                <span className="etiqueta-mini">Peso tara</span>
                <strong className="valor-caja-peso">{pesoTara}</strong>
              </div>
              <div className="caja-peso-resumen">
                <span className="etiqueta-mini">Peso Neto</span>
                <strong className="valor-caja-peso">{pesoNeto}</strong>
              </div>
            </div>
            <p
              className={
                error
                  ? "mensaje-error pesaje-mensaje"
                  : mensaje
                    ? "mensaje-exito pesaje-mensaje"
                    : "subtitulo pesaje-mensaje"
              }
            >
              {error || mensaje || "El peso en vivo llega desde el escritorio/planta."}
            </p>
          </section>

          <section className="tarjeta pesaje-datos">
            <h2 className="titulo-tarjeta">Datos del Pesaje</h2>
            <div className="separador-horizontal" />
            <div className="pesaje-grid-datos">
              <ComboCampo
                etiqueta="Chofer"
                value={choferId}
                onChange={setChoferId}
                items={opciones.choferes}
                plusTo="/choferes"
              />
              <ComboCampo
                etiqueta="Placa"
                requerido
                value={vehiculoId}
                onChange={setVehiculoId}
                items={opciones.vehiculos}
                plusTo="/catalogo"
              />
              <label className="pesaje-campo">
                <span className="etiqueta-mini">Licencia</span>
                <input className="text-field" readOnly value={licencia} placeholder="—" />
              </label>
              <label className="pesaje-campo">
                <span className="etiqueta-mini">Tipo vehículo</span>
                <input className="text-field" readOnly value={tipoVehiculo} placeholder="—" />
              </label>
              <ComboCampo
                etiqueta="Proveedor"
                value={proveedorId}
                onChange={setProveedorId}
                items={opciones.proveedores}
                plusTo="/proveedores"
              />
              <ComboCampo etiqueta="Producto" value={productoId} onChange={setProductoId} items={opciones.productos} />
              <ComboCampo etiqueta="Origen" value={origenId} onChange={setOrigenId} items={opciones.origenes} />
              <ComboCampo etiqueta="Destino" value={destinoId} onChange={setDestinoId} items={opciones.destinos} />
              <label className="pesaje-campo">
                <span className="etiqueta-mini">Costo (Bs)</span>
                <input
                  className="text-field"
                  inputMode="numeric"
                  value={costo}
                  onChange={(e) => setCosto(e.target.value)}
                />
              </label>
              <label className="pesaje-campo">
                <span className="etiqueta-mini">Flete pend. (Bs)</span>
                <input
                  className="text-field"
                  inputMode="numeric"
                  value={flete}
                  onChange={(e) => setFlete(e.target.value)}
                />
              </label>
              <label className="pesaje-campo pesaje-tara-campo">
                <label className="pesaje-check">
                  <input
                    type="checkbox"
                    checked={taraManual}
                    onChange={(e) => setTaraManual(e.target.checked)}
                  />{" "}
                  Tara manual
                </label>
                <span className="etiqueta-mini">Tara (kg)</span>
                <input
                  className="text-field"
                  inputMode="numeric"
                  readOnly={!taraManual}
                  value={taraKg}
                  onChange={(e) => setTaraKg(e.target.value)}
                />
              </label>
            </div>
            <div className="separador-horizontal" />
            <div className="pesaje-guardar">
              <button type="button" className="boton-primario" onClick={registrarPesaje}>
                REGISTRAR PESAJE
              </button>
            </div>
          </section>
        </div>

        <section className="tarjeta tarjeta-sin-padding pesaje-grilla-wrap pesaje-tabla-ajuste tabla-modulo-wrap">
          {loading ? (
            <p className="modulo-vacio" style={{ padding: 12 }}>
              Cargando…
            </p>
          ) : (
            <table id="tabla-pesajes" className="tabla-modulo tabla-modulo--ajuste-ventana">
              <colgroup>
                <col className="pesaje-col-id" />
                <col className="pesaje-col-texto" />
                <col className="pesaje-col-placa" />
                <col className="pesaje-col-corto" />
                <col className="pesaje-col-corto" />
                <col className="pesaje-col-texto" />
                <col className="pesaje-col-texto" />
                <col className="pesaje-col-texto" />
                <col className="pesaje-col-texto" />
                <col className="pesaje-col-num" />
                <col className="pesaje-col-num" />
                <col className="pesaje-col-fecha" />
                <col className="pesaje-col-num" />
                <col className="pesaje-col-num" />
                <col className="pesaje-col-num" />
                <col className="pesaje-col-estado" />
                <col className="pesaje-col-estado" />
                <col className="pesaje-col-acciones" />
              </colgroup>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Chofer</th>
                  <th>Placa</th>
                  <th>Lic.</th>
                  <th>Tipo</th>
                  <th>Prov.</th>
                  <th>Prod.</th>
                  <th>Orig.</th>
                  <th>Dest.</th>
                  <th>Costo</th>
                  <th>Flete</th>
                  <th>Fecha</th>
                  <th>Bruto</th>
                  <th>Tara</th>
                  <th>Neto</th>
                  <th className="tabla-th-estado">Estado</th>
                  <th className="tabla-th-estado">Balanza</th>
                  <th className="tabla-th-acciones">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {filas.length === 0 ? (
                  <tr>
                    <td colSpan={18}>No hay pesajes. Registrá uno o sincronizá desde el escritorio.</td>
                  </tr>
                ) : (
                  filas.map((p) => {
                    const balanza = (p.estado_balanza || "").toUpperCase().trim();
                    const enProceso = balanza === "EN_PROCESO";
                    const finalizado = balanza === "FINALIZADO";
                    const activo = (p.estado || "ACTIVO").toUpperCase() === "ACTIVO";
                    return (
                      <tr key={p.id}>
                        <td title={String(p.id)}>{p.id}</td>
                        <td title={p.nombre_chofer}>{mayus(p.nombre_chofer)}</td>
                        <td title={p.placa}>{mayus(p.placa)}</td>
                        <td title={p.licencia}>{mayus(p.licencia)}</td>
                        <td title={p.tipo_vehiculo}>{mayus(p.tipo_vehiculo)}</td>
                        <td title={p.nombre_proveedor}>{mayus(p.nombre_proveedor)}</td>
                        <td title={p.producto}>{mayus(p.producto)}</td>
                        <td title={p.origen}>{mayus(p.origen)}</td>
                        <td title={p.destino}>{mayus(p.destino)}</td>
                        <td>{p.costo_bs ?? 0}</td>
                        <td>{p.flete_bs_pendiente ?? 0}</td>
                        <td title={fechaHora(p.fecha_creacion)}>{fechaHora(p.fecha_creacion)}</td>
                        <td>{p.peso_bruto ?? 0}</td>
                        <td>{p.tara ?? 0}</td>
                        <td>{p.peso_neto ?? 0}</td>
                        <EstadoEtiqueta estado={activo ? "ACTIVO" : "INACTIVO"} />
                        <EstadoEtiqueta estado={etiquetaBalanza(p.estado_balanza)} />
                        <td className="tabla-celda-acciones">
                          {finalizado ? (
                            <TablaBtnIcono tip="Boleta PDF" act="boleta" disabled />
                          ) : null}
                          {activo && enProceso ? (
                            <>
                              <TablaBtnIcono tip="Editar pesaje" act="editar" disabled />
                              <TablaBtnIcono tip="Finalizar pesaje" act="finalizar" onClick={() => finalizarPesaje(p)} />
                            </>
                          ) : null}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          )}
        </section>
      </div>
    </ModuloPanel>
  );
}