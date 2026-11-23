import { useCallback, useEffect, useState } from "react";
import { fetchApi } from "./api-client";
import ModuloPanel from "./components/ModuloPanel";

type BoletaConfig = {
  id?: number;
  salida?: string;
  direccion?: string;
  celular1?: string;
  celular2?: string;
  ciudad?: string;
  tituloSuperior?: string;
  marca?: string;
  tituloDocumento?: string;
  acreditacion?: string;
};

const CAMPOS: { clave: keyof BoletaConfig; etiqueta: string }[] = [
  { clave: "salida", etiqueta: "Salida" },
  { clave: "direccion", etiqueta: "Dirección" },
  { clave: "celular1", etiqueta: "Celular 1" },
  { clave: "celular2", etiqueta: "Celular 2" },
  { clave: "ciudad", etiqueta: "Ciudad" },
  { clave: "tituloSuperior", etiqueta: "Título superior" },
  { clave: "marca", etiqueta: "Marca" },
  { clave: "tituloDocumento", etiqueta: "Título documento" },
  { clave: "acreditacion", etiqueta: "Acreditación" },
];

function normalizar(clave: string, valor: string): string {
  const v = valor.trim();
  if (clave === "celular1" || clave === "celular2") return v;
  return v.toUpperCase();
}

export default function ConfigBoleta() {
  const [cfg, setCfg] = useState<BoletaConfig>({});
  const [mensaje, setMensaje] = useState("");
  const [error, setError] = useState("");

  const cargar = useCallback(async () => {
    const { body, sessionExpired } = await fetchApi<BoletaConfig>("/api/config/boleta");
    if (sessionExpired) return;
    if (body.success && body.data) setCfg(body.data);
  }, []);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  function setCampo(clave: keyof BoletaConfig, valor: string) {
    setCfg((prev) => ({ ...prev, [clave]: valor }));
  }

  async function guardar() {
    setError("");
    setMensaje("");
    const dto: Record<string, string | number | undefined> = { id: cfg.id };
    for (const c of CAMPOS) {
      dto[c.clave] = normalizar(c.clave, String(cfg[c.clave] ?? ""));
    }
    const { ok, body } = await fetchApi("/api/config/boleta", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(dto),
    });
    if (!ok || !body.success) {
      setError(body.message || "No se pudo guardar");
      return;
    }
    if (body.data) setCfg(body.data as BoletaConfig);
    setMensaje(String(body.message || "Configuración de boleta guardada."));
  }

  return (
    <ModuloPanel titulo="Boleta">
      <section className="tarjeta boleta-config">
        <h2 className="titulo-tarjeta">Configuración de boleta</h2>
        <div className="pesaje-grid-datos">
          {CAMPOS.map((c) => (
            <label key={c.clave} className="pesaje-campo">
              <span className="etiqueta-mini">{c.etiqueta}</span>
              <input
                className="text-field"
                value={String(cfg[c.clave] ?? "")}
                onChange={(e) => setCampo(c.clave, e.target.value)}
              />
            </label>
          ))}
        </div>
        <button type="button" className="boton-primario" onClick={() => void guardar()} style={{ marginTop: 12 }}>
          Guardar
        </button>
        {error ? <p className="mensaje-error">{error}</p> : null}
        {mensaje ? <p className="subtitulo">{mensaje}</p> : null}
      </section>
    </ModuloPanel>
  );
}
