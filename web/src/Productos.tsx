import { CatalogoSimpleAbm } from "./components/CatalogoSimpleAbm";

export default function ProductosPage() {
  return (
    <CatalogoSimpleAbm
      apiPath="/api/productos"
      titulo="Productos"
      crearTexto="Crear producto"
      tituloEntidad="producto"
    />
  );
}
