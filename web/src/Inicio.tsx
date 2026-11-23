import { useEffect, useState } from "react";

export default function Inicio() {
  const [peso, setPeso] = useState<number>(0);
  const [fecha, setFecha] = useState<string | null>(null);
  const [total, setTotal] = useState<number>(0);
  const [error, setError] = useState("");

  useEffect(() => {
    const load = async () => {
      try {
        const [vivo, lista] = await Promise.all([
          fetch("/api/pesajes/ultimo-peso").then((r) => r.json()),
          fetch("/api/pesajes", { credentials: "include" }).then((r) => r.json()),
        ]);
        if (vivo.success) {
          setPeso(Number(vivo.peso) || 0);
          setFecha(vivo.fecha ?? null);
        }
        if (lista.success && Array.isArray(lista.data)) {
          setTotal(lista.data.length);
        }
      } catch {
        setError("No se pudo cargar el resumen");
      }
    };
    load();
    const t = setInterval(async () => {
      const vivo = await fetch("/api/pesajes/ultimo-peso").then((r) => r.json());
      if (vivo.success) {
        setPeso(Number(vivo.peso) || 0);
        setFecha(vivo.fecha ?? null);
      }
    }, 2000);
    return () => clearInterval(t);
  }, []);

  return (
    <section className="dashboard">
      <h2>Inicio</h2>
      {error && <p className="login-error">{error}</p>}
      <div className="dash-grid">
        <article className="dash-card dash-card-live">
          <h3>Peso en vivo (escritorio)</h3>
          <p className="dash-peso">{peso.toLocaleString()} kg</p>
          <p className="muted">{fecha ? `Actualizado: ${fecha}` : "Sin datos de balanza"}</p>
        </article>
        <article className="dash-card">
          <h3>Pesajes en servidor</h3>
          <p className="dash-peso dash-peso-sm">{total}</p>
          <p className="muted">Últimos registros sincronizados</p>
        </article>
      </div>
    </section>
  );
}
