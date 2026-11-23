import { Navigate, Route, Routes } from "react-router-dom";
import Login from "./Login";
import Layout from "./Layout";
import Placeholder from "./Placeholder";
import Inicio from "./Inicio";
import Pesajes from "./Pesajes";
import Empresa from "./Empresa";
import "./app.css";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/inicio" replace />} />
        <Route path="inicio" element={<Inicio />} />
        <Route path="usuarios" element={<Placeholder title="Usuarios" />} />
        <Route path="empresa" element={<Empresa />} />
        <Route path="choferes" element={<Placeholder title="Choferes" />} />
        <Route path="proveedores" element={<Placeholder title="Proveedores" />} />
        <Route path="pesajes" element={<Pesajes />} />
        <Route path="catalogo" element={<Placeholder title="Catálogo" />} />
        <Route path="reportes" element={<Placeholder title="Reportes" />} />
        <Route path="config" element={<Placeholder title="Configuración" />} />
      </Route>
    </Routes>
  );
}
