import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
} from "chart.js";
import { useCallback, useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { fetchApi } from "./api-client";
import ModuloPanel from "./components/ModuloPanel";

type PuntoDia = { fecha: string; cantidad: number };

type Resumen = {
  pesajesTotal: number;
  pesajesHoy: number;
  pesajesEnProceso: number;
  pesajesFinalizadosHoy: number;
  vehiculosActivos: number;
  productosActivos: number;
  choferesActivos: number;
  proveedoresActivos: number;
  ultimos7Dias: PuntoDia[];
};

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

export default function Inicio() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const [error, setError] = useState("");
  const [resumen, setResumen] = useState<Resumen | null>(null);

  const cargar = useCallback(async () => {
    try {
      const { body, sessionExpired } = await fetchApi<Resumen>("/api/inicio/resumen");
      if (sessionExpired) return;
      if (body.success && body.data) {
        setResumen(body.data);
        setError("");
      } else {
        setError(body.message || "No se pudo cargar el resumen");
      }
    } catch {
      setError("No se pudo cargar el resumen");
    }
  }, []);

  useEffect(() => {
    void cargar();
    const timer = window.setInterval(() => void cargar(), 60_000);
    return () => window.clearInterval(timer);
  }, [cargar]);

  const puntos7 = resumen?.ultimos7Dias ?? [];

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || puntos7.length === 0) return;
    const primario =
      getComputedStyle(document.documentElement).getPropertyValue("--pesaje-primario").trim() ||
      "#255ea6";
    const suave =
      getComputedStyle(document.documentElement).getPropertyValue("--pesaje-primario-suave").trim() ||
      "#e8f0fa";
    chartRef.current?.destroy();
    chartRef.current = new Chart(canvas, {
      type: "bar",
      data: {
        labels: puntos7.map((p) => p.fecha),
        datasets: [
          {
            label: "Pesajes",
            data: puntos7.map((p) => p.cantidad),
            backgroundColor: suave,
            borderColor: primario,
            borderWidth: 2,
            borderRadius: 4,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          y: { beginAtZero: true, ticks: { precision: 0 } },
        },
      },
    });
    requestAnimationFrame(() => chartRef.current?.resize());
    return () => {
      chartRef.current?.destroy();
      chartRef.current = null;
    };
  }, [puntos7]);

  const kpiItems = [
    { label: "Pesajes hoy", valor: resumen?.pesajesHoy ?? "—", destacado: true },
    { label: "En proceso", valor: resumen?.pesajesEnProceso ?? "—" },
    { label: "Finalizados hoy", valor: resumen?.pesajesFinalizadosHoy ?? "—" },
    { label: "Total histórico", valor: resumen?.pesajesTotal ?? "—" },
  ];

  const catalogos: ReadonlyArray<readonly [string, number | string]> = [
    ["Vehículos", resumen?.vehiculosActivos ?? "—"],
    ["Productos", resumen?.productosActivos ?? "—"],
    ["Choferes", resumen?.choferesActivos ?? "—"],
    ["Proveedores", resumen?.proveedoresActivos ?? "—"],
  ];

  return (
    <ModuloPanel titulo="Inicio" className="inicio-abm">
      {error ? <p className="abm-mensaje mensaje-error login-error">{error}</p> : null}
      <div className="inicio-dashboard">
        <section className="tarjeta inicio-panel">
          <h3 className="titulo-tarjeta">Resumen del día</h3>
          <div className="inicio-kpis">
            {kpiItems.map((k) => (
              <p key={k.label} className={`inicio-kpi${k.destacado ? " inicio-kpi--destacado" : ""}`}>
                <span className="etiqueta-mini">{k.label}</span>
                <strong className="inicio-kpi-valor">{k.valor}</strong>
              </p>
            ))}
          </div>
          <div className="inicio-atalhos">
            <Link to="/pesajes" className="boton-primario inicio-link-btn">
              Ir a Pesaje
            </Link>
            <Link to="/reportes" className="boton-secundario inicio-link-btn">
              Reportes
            </Link>
          </div>
        </section>
        <section className="tarjeta inicio-panel inicio-panel--grafico">
          <h3 className="titulo-tarjeta">Pesajes — últimos 7 días</h3>
          <div className="inicio-grafico-wrap">
            <canvas ref={canvasRef} aria-label="Gráfico de pesajes por día" />
          </div>
        </section>
        <section className="tarjeta inicio-panel">
          <h3 className="titulo-tarjeta">Catálogos activos</h3>
          <dl className="inicio-grid inicio-grid--catalogos">
            {catalogos.map(([label, n]) => (
              <div key={label}>
                <dt className="subtitulo">{label}</dt>
                <dd className="inicio-valor">{n}</dd>
              </div>
            ))}
          </dl>
        </section>
      </div>
    </ModuloPanel>
  );
}
