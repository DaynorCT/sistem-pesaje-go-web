import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
} from "chart.js";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import ModuloPanel from "./components/ModuloPanel";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

type Pesaje = { fecha_creacion?: string; estado_balanza?: string };

function diaLocal(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso.slice(0, 10);
  return d.toISOString().slice(0, 10);
}

function hoyLocal(): string {
  return new Date().toISOString().slice(0, 10);
}

function ultimos7Dias(): string[] {
  const out: string[] = [];
  const base = new Date();
  base.setHours(12, 0, 0, 0);
  for (let i = 6; i >= 0; i--) {
    const d = new Date(base);
    d.setDate(d.getDate() - i);
    out.push(d.toISOString().slice(0, 10));
  }
  return out;
}

export default function Inicio() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const chartRef = useRef<Chart | null>(null);
  const [error, setError] = useState("");
  const [pesoVivo, setPesoVivo] = useState(0);
  const [kpis, setKpis] = useState({
    hoy: 0,
    proceso: 0,
    finalizadosHoy: 0,
    total: 0,
  });
  const [puntos7, setPuntos7] = useState<{ fecha: string; cantidad: number }[]>([]);

  useEffect(() => {
    const load = async () => {
      try {
        const [vivo, lista] = await Promise.all([
          fetch("/api/pesajes/ultimo-peso").then((r) => r.json()),
          fetch("/api/pesajes", { credentials: "include" }).then((r) => r.json()),
        ]);
        if (vivo.success) setPesoVivo(Number(vivo.peso) || 0);
        const rows: Pesaje[] = lista.success && Array.isArray(lista.data) ? lista.data : [];
        const hoy = hoyLocal();
        let proceso = 0;
        let finalizadosHoy = 0;
        let hoyCount = 0;
        const map = new Map<string, number>();
        for (const f of ultimos7Dias()) map.set(f, 0);
        for (const p of rows) {
          const dia = diaLocal(p.fecha_creacion);
          if (dia === hoy) hoyCount++;
          const est = (p.estado_balanza || "").toLowerCase();
          if (est.includes("proceso")) proceso++;
          if (dia === hoy && (est.includes("final") || est.includes("cerrado"))) finalizadosHoy++;
          if (map.has(dia)) map.set(dia, (map.get(dia) || 0) + 1);
        }
        setKpis({ hoy: hoyCount, proceso, finalizadosHoy, total: rows.length });
        setPuntos7(
          ultimos7Dias().map((fecha) => ({
            fecha: fecha.slice(5),
            cantidad: map.get(fecha) || 0,
          })),
        );
      } catch {
        setError("No se pudo cargar el resumen");
      }
    };
    load();
    const t = setInterval(async () => {
      const vivo = await fetch("/api/pesajes/ultimo-peso").then((r) => r.json());
      if (vivo.success) setPesoVivo(Number(vivo.peso) || 0);
    }, 2000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const primario =
      getComputedStyle(document.documentElement).getPropertyValue("--pesaje-primario").trim() || "#9a3412";
    const suave =
      getComputedStyle(document.documentElement).getPropertyValue("--pesaje-primario-suave").trim() || "#f6ece6";
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
          y: { beginAtZero: true, ticks: { stepSize: 1 } },
        },
      },
    });
    return () => {
      chartRef.current?.destroy();
      chartRef.current = null;
    };
  }, [puntos7]);

  const kpiItems = [
    { label: "Pesajes hoy", valor: kpis.hoy, destacado: true },
    { label: "En proceso", valor: kpis.proceso },
    { label: "Finalizados hoy", valor: kpis.finalizadosHoy },
    { label: "Total histórico", valor: kpis.total },
  ];

  return (
    <ModuloPanel title="Inicio" className="inicio-abm">
      {error ? <p className="abm-mensaje mensaje-error login-error">{error}</p> : null}
      <div className="inicio-dashboard">
        <section className="tarjeta inicio-panel">
          <h3 className="titulo-tarjeta">Peso en vivo (escritorio)</h3>
          <div className="display-peso">
            <span className="valor-peso">{pesoVivo.toLocaleString()} kg</span>
          </div>
        </section>
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
            <Link to="/pesajes" className="boton-primario" style={{ textDecoration: "none", display: "inline-block" }}>
              Ir a Pesaje
            </Link>
            <Link to="/reportes" className="boton-secundario" style={{ textDecoration: "none", display: "inline-block" }}>
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
            {[
              ["Vehículos", "—"],
              ["Productos", "—"],
              ["Choferes", "—"],
              ["Proveedores", "—"],
            ].map(([label, n]) => (
              <div key={label}>
                <dt className="subtitulo">{label}</dt>
                <dd className="inicio-valor">{n}</dd>
              </div>
            ))}
          </dl>
          <p className="modulo-vacio">Los totales de catálogo se sincronizan desde el escritorio.</p>
        </section>
      </div>
    </ModuloPanel>
  );
}