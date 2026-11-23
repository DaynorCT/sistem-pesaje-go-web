import { useCallback, useState } from "react";
import { Dialogo, DialogoConfirmarAcciones } from "./Dialogo";

export type ConfirmarOpts = {
  confirmar?: string;
  cancelar?: string;
};

type Pendiente = {
  titulo: string;
  mensaje: string;
  opts?: ConfirmarOpts;
  resolve: (valor: boolean) => void;
};

/** Sustituto de `window.confirm` con el `<dialog>` de la app (paridad `crearDialogoConfirmar`). */
export function useConfirmar() {
  const [pendiente, setPendiente] = useState<Pendiente | null>(null);

  const preguntar = useCallback(
    (titulo: string, mensaje: string, opts?: ConfirmarOpts) =>
      new Promise<boolean>((resolve) => {
        setPendiente({ titulo, mensaje, opts, resolve });
      }),
    [],
  );

  const cerrar = (valor: boolean) => {
    if (!pendiente) return;
    pendiente.resolve(valor);
    setPendiente(null);
  };

  const dialogo = pendiente ? (
    <Dialogo abierto onCerrar={() => cerrar(false)} ancho="compacto" titulo={pendiente.titulo}>
      <p className="subtitulo dlg-intro">{pendiente.mensaje}</p>
      <DialogoConfirmarAcciones
        cancelarTexto={pendiente.opts?.cancelar ?? "CANCELAR"}
        confirmarTexto={pendiente.opts?.confirmar ?? "CONFIRMAR"}
        onCancelar={() => cerrar(false)}
        onConfirmar={() => cerrar(true)}
      />
    </Dialogo>
  ) : null;

  return { preguntar, dialogo };
}
