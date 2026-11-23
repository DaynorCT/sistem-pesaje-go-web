import { FormEvent, ReactNode, useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import type { AnchoDialogo } from "./dialogo-utils";

type DialogoProps = {
  abierto: boolean;
  onCerrar: () => void;
  titulo?: string;
  ancho?: AnchoDialogo;
  scroll?: boolean;
  children: ReactNode;
  /** Si se pasa, el diálogo envuelve un `<form>` con submit. */
  onSubmit?: (e: FormEvent) => void;
  acciones?: ReactNode;
};

export function Dialogo({
  abierto,
  onCerrar,
  titulo,
  ancho = "medio",
  scroll,
  children,
  onSubmit,
  acciones,
}: DialogoProps) {
  const ref = useRef<HTMLDialogElement>(null);
  const ignorarCierreRef = useRef(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    if (abierto && !el.open) {
      ignorarCierreRef.current = true;
      const id = window.setTimeout(() => {
        ignorarCierreRef.current = false;
      }, 200);
      // Un frame después del clic en “Crear…” evita que el mismo click cierre el modal.
      requestAnimationFrame(() => {
        if (!el.isConnected) return;
        if (!el.open) {
          try {
            el.showModal();
          } catch {
            /* ya abierto */
          }
        }
      });
      return () => window.clearTimeout(id);
    }

    if (!abierto && el.open) {
      el.close();
    }
  }, [abierto]);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const onCancel = (ev: Event) => {
      ev.preventDefault();
      onCerrar();
    };
    el.addEventListener("cancel", onCancel);
    return () => el.removeEventListener("cancel", onCancel);
  }, [onCerrar]);

  const cerrarSiCorresponde = () => {
    if (ignorarCierreRef.current) return;
    onCerrar();
  };

  const cuerpo = (
    <>
      {titulo ? <h3 className="titulo-tarjeta">{titulo}</h3> : null}
      {children}
      {acciones}
    </>
  );

  const dialog = (
    <dialog
      ref={ref}
      className={`dlg-aviso dlg-ancho-${ancho}`}
      onClose={() => {
        if (ignorarCierreRef.current) return;
        onCerrar();
      }}
      onClick={(ev) => {
        if (ev.target === ref.current) cerrarSiCorresponde();
      }}
    >
      {onSubmit ? (
        <form
          className={`tarjeta dlg-contenido${scroll ? " dlg-contenido--scroll" : ""}`}
          onSubmit={onSubmit}
          onClick={(ev) => ev.stopPropagation()}
        >
          {cuerpo}
        </form>
      ) : (
        <div
          className={`tarjeta dlg-contenido${scroll ? " dlg-contenido--scroll" : ""}`}
          onClick={(ev) => ev.stopPropagation()}
        >
          {cuerpo}
        </div>
      )}
    </dialog>
  );

  return createPortal(dialog, document.body);
}

export function DialogoAcciones({
  onCancelar,
  guardar = true,
  guardarTexto = "GUARDAR",
  cancelarTexto = "CANCELAR",
  extra,
}: {
  onCancelar: () => void;
  guardar?: boolean;
  guardarTexto?: string;
  cancelarTexto?: string;
  extra?: ReactNode;
}) {
  return (
    <div className="dlg-acciones">
      {extra}
      <button type="button" className="boton-secundario" onClick={onCancelar}>
        {cancelarTexto}
      </button>
      {guardar ? (
        <button type="submit" className="boton-primario">
          {guardarTexto}
        </button>
      ) : null}
    </div>
  );
}