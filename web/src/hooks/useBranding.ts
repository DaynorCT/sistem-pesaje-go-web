import { useEffect, useState } from "react";
import logoDefault from "../assets/images/logo-pesaje.jpg";

/** Logo y nombre de la empresa activa (`GET /api/empresas/branding`), con fallback canónico. */
export function useBranding() {
  const [logoSrc, setLogoSrc] = useState(logoDefault);
  const [nombre, setNombre] = useState("Sistema de Pesaje");

  useEffect(() => {
    fetch("/api/empresas/branding")
      .then((r) => r.json())
      .then((d) => {
        if (!d.success || !d.data) return;
        if (d.data.nombre) setNombre(d.data.nombre);
        const logo = (d.data.logo as string)?.trim();
        if (logo) setLogoSrc(logo);
      })
      .catch(() => {});
  }, []);

  return { logoSrc, nombre };
}