import { ReactNode } from "react";

export function TablaAbm({
  loading,
  children,
  vacio,
}: {
  loading?: boolean;
  children: ReactNode;
  vacio?: ReactNode;
}) {
  return (
    <div className="tabla-modulo-wrap tarjeta tarjeta-sin-padding">
      {loading ? (
        <p className="modulo-vacio" style={{ padding: 12 }}>
          Cargando…
        </p>
      ) : (
        children
      )}
      {!loading && vacio}
    </div>
  );
}