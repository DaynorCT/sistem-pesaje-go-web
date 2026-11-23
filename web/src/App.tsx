import { Navigate, Route, Routes } from "react-router-dom";
import Login from "./Login";
import Layout from "./Layout";
import Placeholder from "./Placeholder";
import Inicio from "./Inicio";
import Pesajes from "./Pesajes";
import Empresa from "./Empresa";
import Usuarios from "./Usuarios";
import Choferes from "./Choferes";
import Proveedores from "./Proveedores";
import Catalogo from "./Catalogo";
import Vehiculos from "./Vehiculos";
import Productos from "./Productos";
import Origenes from "./Origenes";
import Destinos from "./Destinos";
import ConfigBoleta from "./ConfigBoleta";
import "./app.css";
import "./css/shell.css";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/inicio" replace />} />
        <Route path="inicio" element={<Inicio />} />
        <Route path="usuarios" element={<Usuarios />} />
        <Route path="empresa" element={<Empresa />} />
        <Route path="choferes" element={<Choferes />} />
        <Route path="proveedores" element={<Proveedores />} />
        <Route path="pesajes" element={<Pesajes />} />
        <Route path="catalogo" element={<Catalogo />} />
        <Route path="vehiculos" element={<Vehiculos />} />
        <Route path="productos" element={<Productos />} />
        <Route path="origenes" element={<Origenes />} />
        <Route path="destinos" element={<Destinos />} />
        <Route path="reportes" element={<Placeholder title="Reportes" />} />
        <Route path="config" element={<Placeholder title="Configuración" note="Elegí Balanza o Boleta en el submenú Config." />} />
        <Route path="config/balanza" element={<Placeholder title="Balanza" />} />
        <Route path="config/boleta" element={<ConfigBoleta />} />
      </Route>
    </Routes>
  );
}
