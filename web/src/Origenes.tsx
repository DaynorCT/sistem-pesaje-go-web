import { CatalogoSimpleAbm } from "./components/CatalogoSimpleAbm";

export default function OrigenesPage() {
  return (
    <CatalogoSimpleAbm
      apiPath="/api/origenes"
      titulo="Orígenes"
      crearTexto="Crear origen"
      tituloEntidad="origen"
    />
  );
}
