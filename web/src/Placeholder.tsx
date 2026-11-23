import ModuloPanel from "./components/ModuloPanel";

export default function Placeholder({ title, note }: { title: string; note?: string }) {
  return (
    <ModuloPanel title={title}>
      <section className="tarjeta">
        <p className="modulo-vacio">Módulo en construcción — la API Go ya está lista para sync y pesajes.</p>
        {note ? <p className="subtitulo">{note}</p> : null}
      </section>
    </ModuloPanel>
  );
}