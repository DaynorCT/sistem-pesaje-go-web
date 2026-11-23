import { CatalogoSimpleAbm } from "./components/CatalogoSimpleAbm";

export default function DestinosPage() {
  return (
    <CatalogoSimpleAbm
      apiPath="/api/destinos"
      titulo="Destinos"
      crearTexto="Crear destino"
      tituloEntidad="destino"
    />
  );
}
