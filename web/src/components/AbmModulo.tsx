import { FormEvent, ReactNode } from "react";
import ModuloPanel from "./ModuloPanel";
import { Dialogo, DialogoAcciones } from "./Dialogo";
import { TablaAbm } from "./TablaAbm";
import { anchoDialogoPorCampos, columnasDialogoPorCampos } from "./dialogo-utils";
import type { AnchoDialogo } from "./dialogo-utils";

export type AbmModuloProps = {
  titulo: string;
  crearTexto: string;
  onCrear: () => void;
  mensaje?: string;
  mensajeError?: string;
  loading?: boolean;
  dialogoAbierto: boolean;
  onCerrarDialogo: () => void;
  tituloDialogo: string;
  onSubmit: (e: FormEvent) => void;
  /** Campos del formulario dentro del diálogo */
  formulario: ReactNode;
  /** Filas de la tabla (sin thead; el módulo define columnas) */
  tabla: ReactNode;
  cabeceraTabla: ReactNode;
  vacioFilas?: ReactNode;
  /** Para elegir ancho/columnas del grid automáticamente */
  cantidadCampos?: number;
  /** Fuerza columnas del grid (p. ej. usuarios: 3). */
  columnasDialogo?: 2 | 3;
  anchoDialogo?: AnchoDialogo;
  scrollDialogo?: boolean;
  accionesExtra?: ReactNode;
};

/**
 * Shell ABM reutilizable: grilla fija + Crear + formulario en `<dialog>` (como pesaje-go).
 */
export function AbmModulo({
  titulo,
  crearTexto,
  onCrear,
  mensaje,
  mensajeError,
  loading,
  dialogoAbierto,
  onCerrarDialogo,
  tituloDialogo,
  onSubmit,
  formulario,
  tabla,
  cabeceraTabla,
  vacioFilas,
  cantidadCampos = 6,
  columnasDialogo,
  anchoDialogo,
  scrollDialogo,
  accionesExtra,
}: AbmModuloProps) {
  const cols = columnasDialogo ?? columnasDialogoPorCampos(cantidadCampos);
  const ancho = anchoDialogo ?? anchoDialogoPorCampos(cantidadCampos);

  return (
    <ModuloPanel
      titulo={titulo}
      actions={
        <div className="abm-cabecera-der">
          <p className="subtitulo abm-mensaje">
            {mensaje ? <span className="mensaje-exito">{mensaje}</span> : null}
            {mensajeError ? <span className="mensaje-error"> {mensajeError}</span> : null}
          </p>
          <button type="button" className="boton-primario" onClick={onCrear}>
            {crearTexto}
          </button>
        </div>
      }
    >
      <TablaAbm loading={loading}>
        <table className="tabla-modulo">
          <thead>
            <tr>{cabeceraTabla}</tr>
          </thead>
          <tbody>{tabla}</tbody>
        </table>
        {vacioFilas}
      </TablaAbm>

      <Dialogo
        abierto={dialogoAbierto}
        onCerrar={onCerrarDialogo}
        titulo={tituloDialogo}
        ancho={ancho}
        scroll={scrollDialogo ?? cols === 3}
        onSubmit={onSubmit}
        acciones={
          <DialogoAcciones onCancelar={onCerrarDialogo} extra={accionesExtra} />
        }
      >
        <div className={`abm-dialog-grid abm-dialog-grid--${cols}`}>{formulario}</div>
      </Dialogo>
    </ModuloPanel>
  );
}