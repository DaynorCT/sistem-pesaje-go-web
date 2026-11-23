import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import ModuloPanel from "./components/ModuloPanel";
import { fetchApi, fetchApiBlob } from "./api-client";
import { EstadoEtiqueta } from "./tablas-acciones";
import "./css/reportes.css";

const ICONO_CALENDARIO = (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2.2"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden
  >
    <rect x="3" y="4" width="18" height="18" rx="2" />
    <path d="M16 2v4M8 2v4M3 10h18" />
  </svg>
);

type FilaReporte = {
  id: number;
  guia: string;
  nombreChofer: string;
  placa: string;
  nombreProveedor: string;
  pesoBruto: number;
  tara: number;
  pesoNeto: number;
  estadoBalanza: string;
  fecha: string;
  hora: string;
};

type Filtros = { desde: string; hasta: string; placa: string };

function aIsoFecha(raw: string): string {
  const t = raw.trim();
  if (!t) return "";
  if (/^\d{4}-\d{2}-\d{2}$/.test(t)) return t;
  const m = t.match(/^(\d{1,2})[/.-](\d{1,2})[/.-](\d{4})$/);
  if (m) {
    const dd = m[1].padStart(2, "0");
    const mm = m[2].padStart(2, "0");
    return `${m[3]}-${mm}-${dd}`;
  }
  return "";
}

function esFechaIsoValida(iso: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(iso)) return false;
  const [y, m, d] = iso.split("-").map((x) => Number(x));
  const dt = new Date(y, m - 1, d);
  return dt.getFullYear() === y && dt.getMonth() === m - 1 && dt.getDate() === d;
}

function isoAVisual(iso: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(iso)) return "";
  const [y, m, d] = iso.split("-");
  return `${d}/${m}/${y}`;
}

function queryFiltros(f: Filtros): string {
  const p = new URLSearchParams();
  if (f.desde) p.set("desde", f.desde);
  if (f.hasta) p.set("hasta", f.hasta);
  if (f.placa) p.set("placa", f.placa);
  const s = p.toString();
  return s ? `?${s}` : "";
}

function CampoFecha({
  id,
  etiqueta,
  value,
  onChange,
}: {
  id: string;
  etiqueta: string;
  value: string;
  onChange: (v: string) => void;
}) {
  const pickerRef = useRef<HTMLInputElement>(null);

  const blurTexto = () => {
    const iso = aIsoFecha(value);
    if (esFechaIsoValida(iso)) onChange(isoAVisual(iso));
  };

  return (
    <label className="rep-filtro rep-filtro--fecha" htmlFor={id}>
      <span className="etiqueta-mini">{etiqueta}</span>
      <span className="rep-fecha-wrap">
        <input
          type="text"
          className="text-field rep-fecha"
          id={id}
          inputMode="numeric"
          placeholder="dd/mm/aaaa"
          autoComplete="off"
          spellCheck={false}
          maxLength={10}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onBlur={blurTexto}
        />
        <span className="rep-fecha-cal" title="Calendario">
          <span className="rep-fecha-icono-visual">{ICONO_CALENDARIO}</span>
          <input
            ref={pickerRef}
            type="date"
            className="rep-fecha-picker-native"
            title="Elegir fecha"
            onChange={(e) => {
              if (e.target.value) onChange(isoAVisual(e.target.value));
            }}
            onClick={() => {
              const iso = aIsoFecha(value);
              if (pickerRef.current && esFechaIsoValida(iso)) pickerRef.current.value = iso;
            }}
          />
        </span>
      </span>
    </label>
  );
}

export default function Reportes() {
  const nav = useNavigate();
  const [desdeTexto, setDesdeTexto] = useState("");
  const [hastaTexto, setHastaTexto] = useState("");
  const [placaTexto, setPlacaTexto] = useState("");
  const [filas, setFilas] = useState<FilaReporte[]>([]);
  const [seleccion, setSeleccion] = useState(0);
  const [mensaje, setMensaje] = useState("");
  const [esError, setEsError] = useState(false);

  const filtrosIso = useCallback((): Filtros | null => {
    for (const [raw, etiqueta] of [
      [desdeTexto, "Desde"],
      [hastaTexto, "Hasta"],
    ] as const) {
      if (!raw.trim()) continue;
      const iso = aIsoFecha(raw);
      if (!esFechaIsoValida(iso)) {
        setMensaje(`Fecha inválida en «${etiqueta}». Use dd/mm/aaaa.`);
        setEsError(true);
        return null;
      }
    }
    const desde = aIsoFecha(desdeTexto);
    const hasta = aIsoFecha(hastaTexto);
    if (desde && hasta && desde > hasta) {
      setMensaje("«Desde» no puede ser posterior a «Hasta».");
      setEsError(true);
      return null;
    }
    const placa = placaTexto.trim().toUpperCase();
    return { desde, hasta, placa };
  }, [desdeTexto, hastaTexto, placaTexto]);

  const cargar = useCallback(async () => {
    const f = filtrosIso();
    if (!f) return;
    setEsError(false);
    const { ok, body, sessionExpired } = await fetchApi<FilaReporte[]>(`/api/reportes${queryFiltros(f)}`);
    if (sessionExpired) {
      nav("/login");
      return;
    }
    if (!ok || !Array.isArray(body.data)) {
      setMensaje(body.message || "No se pudo cargar el listado.");
      setEsError(true);
      setFilas([]);
      setSeleccion(0);
      return;
    }
    setFilas(body.data);
    setSeleccion(0);
    setMensaje(body.data.length === 0 ? "No hay pesajes con esos filtros." : `${body.data.length} registro(s).`);
    setEsError(false);
  }, [filtrosIso, nav]);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  async function abrirBlob(url: string) {
    setEsError(false);
    const { ok, blob, message, sessionExpired } = await fetchApiBlob(url);
    if (sessionExpired) {
      nav("/login");
      return;
    }
    if (!ok || !blob) {
      setMensaje(message || "No se pudo generar el PDF.");
      setEsError(true);
      return;
    }
    const obj = URL.createObjectURL(blob);
    window.open(obj, "_blank", "noopener,noreferrer");
    window.setTimeout(() => URL.revokeObjectURL(obj), 60_000);
    setMensaje("PDF generado.");
    setEsError(false);
  }

  async function boletaSeleccionada() {
    if (!seleccion) {
      setMensaje("Seleccione un pesaje de la lista.");
      setEsError(true);
      return;
    }
    const fila = filas.find((x) => x.id === seleccion);
    if (fila && fila.estadoBalanza.toUpperCase() !== "FINALIZADO") {
      setMensaje("La boleta solo está disponible para pesajes FINALIZADOS.");
      setEsError(true);
      return;
    }
    await abrirBlob(`/api/pesajes/${seleccion}/boleta`);
  }

  async function abrirReportePdf() {
    const f = filtrosIso();
    if (!f) return;
    await abrirBlob(`/api/reportes/pdf${queryFiltros(f)}`);
  }

  return (
    <ModuloPanel titulo="Reportes" className="reportes-panel">
      <div className="reportes-barra">
        <CampoFecha id="rep-desde" etiqueta="Desde" value={desdeTexto} onChange={setDesdeTexto} />
        <CampoFecha id="rep-hasta" etiqueta="Hasta" value={hastaTexto} onChange={setHastaTexto} />
        <label className="rep-filtro rep-filtro--placa">
          <span className="etiqueta-mini">Placa</span>
          <input
            className="text-field"
            id="rep-placa"
            placeholder="Todas"
            value={placaTexto}
            onChange={(e) => setPlacaTexto(e.target.value)}
          />
        </label>
        <div className="rep-filtro-acciones">
          <button type="button" className="boton-secundario" onClick={() => void cargar()}>
            Actualizar
          </button>
          <button type="button" className="boton-primario" onClick={() => void abrirReportePdf()}>
            Reporte PDF
          </button>
          <button type="button" className="boton-primario" onClick={() => void boletaSeleccionada()}>
            Boleta del pesaje
          </button>
        </div>
      </div>
      {mensaje ? (
        <p className={esError ? "mensaje-error" : "subtitulo mensaje-exito"}>{mensaje}</p>
      ) : (
        <p className="subtitulo">Filtre por fecha y placa; seleccione una fila para la boleta.</p>
      )}
      <div className="tarjeta tarjeta-sin-padding tabla-modulo-wrap">
        <table className="tabla-modulo" id="tabla-reportes">
          <thead>
            <tr>
              <th>Guía</th>
              <th>Vehículo</th>
              <th>Chofer</th>
              <th>Proveedor</th>
              <th>Bruto</th>
              <th>Tara</th>
              <th>Neto</th>
              <th className="tabla-th-estado">Estado</th>
              <th>Fecha</th>
            </tr>
          </thead>
          <tbody>
            {filas.map((f) => {
              const estadoTxt = (f.estadoBalanza || "—").replace(/_/g, " ");
              const sel = f.id === seleccion;
              return (
                <tr
                  key={f.id}
                  className={sel ? "seleccionado" : undefined}
                  onClick={() => setSeleccion(f.id)}
                >
                  <td>{f.guia || f.id}</td>
                  <td>{f.placa}</td>
                  <td>{f.nombreChofer}</td>
                  <td>{f.nombreProveedor}</td>
                  <td>{f.pesoBruto}</td>
                  <td>{f.tara}</td>
                  <td>{f.pesoNeto}</td>
                  <EstadoEtiqueta estado={estadoTxt} />
                  <td>
                    {f.fecha} {f.hora}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </ModuloPanel>
  );
}
