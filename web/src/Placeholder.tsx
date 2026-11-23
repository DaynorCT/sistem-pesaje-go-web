export default function Placeholder({ title, note }: { title: string; note?: string }) {
  return (
    <section className="placeholder">
      <h2>{title}</h2>
      <p>Módulo en construcción — la API Go ya está lista para sync y pesajes.</p>
      {note && <p className="muted">{note}</p>}
    </section>
  );
}
