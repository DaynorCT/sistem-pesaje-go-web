import { useEffect, useState } from "react";

type Pesaje = {
  id: number;
  guia?: string;
  peso_bruto?: number;
  tara?: number;
  peso_neto?: number;
  estado_balanza?: string;
  fecha_creacion?: string;
};

export default function Pesajes() {
  const [pesoVivo, setPesoVivo] = useState(0);
  const [fechaVivo, setFechaVivo] = useState<string | null>(null);
  const [rows, setRows] = useState<Pesaje[]>([]);
  const [loading, setLoading] = useState(true);

  const refreshList = () =>
    fetch("/api/pesajes", { credentials: "include" })
      .then((r) => r.json())
      .then((d) => {
        if (d.success && Array.isArray(d.data)) setRows(d.data);
      })
      .finally(() => setLoading(false));

  useEffect(() => {
    refreshList();
    const poll = setInterval(() => {
      fetch("/api/pesajes/ultimo-peso")
        .then((r) => r.json())
        .then((d) => {
          if (d.success) {
            setPesoVivo(Number(d.peso) || 0);
            setFechaVivo(d.fecha ?? null);
          }
        });
    }, 1500);
    return () => clearInterval(poll);
  }, []);

  return (
    <section className="pesajes-page">
      <h2>Pesaje</h2>
      <div className="peso-vivo-banner">
        <span className="label">Peso en vivo desde planta</span>
        <span className="valor">{pesoVivo.toLocaleString()} kg</span>
        <span className="muted">{fechaVivo ?? "—"}</span>
      </div>
      {loading ? (
        <p>Cargando…</p>
      ) : (
        <table className="pesajes-table">
          <thead>
            <tr>
              <th>Guía</th>
              <th>Bruto</th>
              <th>Tara</th>
              <th>Neto</th>
              <th>Estado</th>
              <th>Fecha</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td colSpan={6}>No hay pesajes aún. El escritorio los sincroniza acá.</td>
              </tr>
            ) : (
              rows.map((p) => (
                <tr key={p.id}>
                  <td>{p.guia || `#${p.id}`}</td>
                  <td>{p.peso_bruto ?? 0}</td>
                  <td>{p.tara ?? 0}</td>
                  <td>{p.peso_neto ?? 0}</td>
                  <td>{p.estado_balanza ?? "—"}</td>
                  <td>{p.fecha_creacion ?? "—"}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      )}
    </section>
  );
}
