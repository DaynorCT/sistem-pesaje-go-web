import { FormEvent, useEffect, useMemo, useState } from "react";
import { fetchApi } from "./api-client";
import { Campo } from "./components/Campo";
import { Dialogo, DialogoAcciones } from "./components/Dialogo";
import { anchoDialogoPorCampos, columnasDialogoPorCampos } from "./components/dialogo-utils";

export type RapidoTipo = "vehiculo" | "chofer" | "proveedor";

type VehiculoForm = { placa: string; tipo_vehiculo: string; tara: string };
type ChoferForm = {
  nombre: string;
  apellido_paterno: string;
  apellido_materno: string;
  ci: string;
  telefono: string;
  licencia: string;
};
type ProveedorForm = {
  nombre: string;
  apellido_paterno: string;
  apellido_materno: string;
  ci: string;
  telefono: string;
  nombre_empresa: string;
  descripcion: string;
};

const titulos: Record<RapidoTipo, string> = {
  vehiculo: "Nuevo vehículo",
  chofer: "Nuevo chofer",
  proveedor: "Nuevo proveedor",
};

const vacioVehiculo = (): VehiculoForm => ({ placa: "", tipo_vehiculo: "", tara: "0" });
const vacioChofer = (): ChoferForm => ({
  nombre: "",
  apellido_paterno: "",
  apellido_materno: "",
  ci: "",
  telefono: "",
  licencia: "",
});
const vacioProveedor = (): ProveedorForm => ({
  nombre: "",
  apellido_paterno: "",
  apellido_materno: "",
  ci: "",
  telefono: "",
  nombre_empresa: "",
  descripcion: "",
});

export function PesajeDialogoRapido({
  tipo,
  onCerrar,
  onExito,
}: {
  tipo: RapidoTipo | null;
  onCerrar: () => void;
  onExito: (tipo: RapidoTipo, id: number) => void;
}) {
  const [vehiculo, setVehiculo] = useState(vacioVehiculo);
  const [chofer, setChofer] = useState(vacioChofer);
  const [proveedor, setProveedor] = useState(vacioProveedor);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!tipo) return;
    setError("");
    setVehiculo(vacioVehiculo());
    setChofer(vacioChofer());
    setProveedor(vacioProveedor());
  }, [tipo]);

  const campos = tipo === "vehiculo" ? 3 : tipo === "chofer" ? 6 : 7;
  const cols = useMemo(() => columnasDialogoPorCampos(campos), [campos]);
  const ancho = useMemo(() => anchoDialogoPorCampos(campos), [campos]);

  async function guardar(e: FormEvent) {
    e.preventDefault();
    if (!tipo) return;
    setError("");

    let endpoint = "";
    let body: Record<string, unknown> = {};

    if (tipo === "vehiculo") {
      if (!vehiculo.placa.trim()) {
        setError("La placa es obligatoria.");
        return;
      }
      endpoint = "/api/vehiculos";
      body = {
        placa: vehiculo.placa.trim().toUpperCase(),
        tipo_vehiculo: vehiculo.tipo_vehiculo.trim().toUpperCase(),
        tara: Number(vehiculo.tara) || 0,
      };
    } else if (tipo === "chofer") {
      if (!chofer.nombre.trim()) {
        setError("El nombre es obligatorio.");
        return;
      }
      if (!chofer.licencia.trim()) {
        setError("La licencia es obligatoria.");
        return;
      }
      endpoint = "/api/choferes";
      body = {
        nombre: chofer.nombre.trim().toUpperCase(),
        apellido_paterno: chofer.apellido_paterno.trim().toUpperCase(),
        apellido_materno: chofer.apellido_materno.trim().toUpperCase(),
        ci: chofer.ci.trim(),
        telefono: chofer.telefono.trim(),
        licencia: chofer.licencia.trim().toUpperCase(),
      };
    } else {
      if (!proveedor.nombre.trim() && !proveedor.nombre_empresa.trim()) {
        setError("Indique el nombre o la empresa (al menos uno).");
        return;
      }
      endpoint = "/api/proveedores";
      body = {
        nombre: proveedor.nombre.trim().toUpperCase(),
        apellido_paterno: proveedor.apellido_paterno.trim().toUpperCase(),
        apellido_materno: proveedor.apellido_materno.trim().toUpperCase(),
        ci: proveedor.ci.trim(),
        telefono: proveedor.telefono.trim(),
        nombre_empresa: proveedor.nombre_empresa.trim().toUpperCase(),
        descripcion: proveedor.descripcion.trim().toUpperCase(),
      };
    }

    const { ok, body: res } = await fetchApi<{ id?: number }>(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!ok || !res.success) {
      setError(res.message || "No se pudo guardar");
      return;
    }
    const id = Number(res.id);
    if (!id) {
      setError("Creado pero sin ID en la respuesta");
      return;
    }
    onExito(tipo, id);
    onCerrar();
  }

  const formulario =
    tipo === "vehiculo" ? (
      <>
        <Campo etiqueta="Placa" requerido>
          <input
            className="text-field"
            value={vehiculo.placa}
            onChange={(ev) => setVehiculo({ ...vehiculo, placa: ev.target.value })}
            autoFocus
          />
        </Campo>
        <Campo etiqueta="Tipo">
          <input
            className="text-field"
            value={vehiculo.tipo_vehiculo}
            onChange={(ev) => setVehiculo({ ...vehiculo, tipo_vehiculo: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Tara (kg)">
          <input
            className="text-field"
            inputMode="numeric"
            value={vehiculo.tara}
            onChange={(ev) => setVehiculo({ ...vehiculo, tara: ev.target.value })}
          />
        </Campo>
      </>
    ) : tipo === "chofer" ? (
      <>
        <Campo etiqueta="Nombre" requerido>
          <input
            className="text-field"
            value={chofer.nombre}
            onChange={(ev) => setChofer({ ...chofer, nombre: ev.target.value })}
            autoFocus
          />
        </Campo>
        <Campo etiqueta="Apellido paterno">
          <input
            className="text-field"
            value={chofer.apellido_paterno}
            onChange={(ev) => setChofer({ ...chofer, apellido_paterno: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Apellido materno">
          <input
            className="text-field"
            value={chofer.apellido_materno}
            onChange={(ev) => setChofer({ ...chofer, apellido_materno: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Nro. documento">
          <input
            className="text-field"
            value={chofer.ci}
            onChange={(ev) => setChofer({ ...chofer, ci: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Teléfono">
          <input
            className="text-field"
            value={chofer.telefono}
            onChange={(ev) => setChofer({ ...chofer, telefono: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Licencia" requerido>
          <input
            className="text-field"
            value={chofer.licencia}
            onChange={(ev) => setChofer({ ...chofer, licencia: ev.target.value })}
          />
        </Campo>
      </>
    ) : tipo === "proveedor" ? (
      <>
        <Campo etiqueta="Nombre" requerido>
          <input
            className="text-field"
            value={proveedor.nombre}
            onChange={(ev) => setProveedor({ ...proveedor, nombre: ev.target.value })}
            autoFocus
          />
        </Campo>
        <Campo etiqueta="Apellido paterno">
          <input
            className="text-field"
            value={proveedor.apellido_paterno}
            onChange={(ev) => setProveedor({ ...proveedor, apellido_paterno: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Apellido materno">
          <input
            className="text-field"
            value={proveedor.apellido_materno}
            onChange={(ev) => setProveedor({ ...proveedor, apellido_materno: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="CI">
          <input
            className="text-field"
            value={proveedor.ci}
            onChange={(ev) => setProveedor({ ...proveedor, ci: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Teléfono">
          <input
            className="text-field"
            value={proveedor.telefono}
            onChange={(ev) => setProveedor({ ...proveedor, telefono: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Empresa" requerido>
          <input
            className="text-field"
            value={proveedor.nombre_empresa}
            onChange={(ev) => setProveedor({ ...proveedor, nombre_empresa: ev.target.value })}
          />
        </Campo>
        <Campo etiqueta="Descripción" anchoCompleto>
          <input
            className="text-field"
            value={proveedor.descripcion}
            onChange={(ev) => setProveedor({ ...proveedor, descripcion: ev.target.value })}
          />
        </Campo>
      </>
    ) : null;

  return (
    <Dialogo
      abierto={tipo !== null}
      onCerrar={onCerrar}
      titulo={tipo ? titulos[tipo] : undefined}
      ancho={ancho}
      scroll={cols === 3}
      onSubmit={guardar}
      acciones={<DialogoAcciones onCancelar={onCerrar} />}
    >
      <div className={`abm-dialog-grid abm-dialog-grid--${cols}`}>{formulario}</div>
      {error ? (
        <p className="mensaje-error" style={{ marginTop: 8 }}>
          {error}
        </p>
      ) : null}
    </Dialogo>
  );
}