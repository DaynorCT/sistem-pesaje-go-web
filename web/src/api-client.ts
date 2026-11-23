/** Respuesta JSON del backend Go; evita quedar en "Cargando…" si el servidor devuelve HTML (SPA vieja). */
export async function fetchApi<T = unknown>(
  url: string,
  init?: RequestInit,
): Promise<{ ok: boolean; status: number; body: { success?: boolean; data?: T; message?: string; [k: string]: unknown } }> {
  const res = await fetch(url, { credentials: "include", ...init });
  const ct = res.headers.get("content-type") || "";
  if (!ct.includes("application/json")) {
    return {
      ok: false,
      status: res.status,
      body: {
        success: false,
        message:
          res.status === 200
            ? "La API no respondió JSON (reiniciá el servidor Go con el código actual)."
            : `Error del servidor (${res.status}).`,
      },
    };
  }
  let body: { success?: boolean; data?: T; message?: string; [k: string]: unknown };
  try {
    body = await res.json();
  } catch {
    body = { success: false, message: "Respuesta inválida del servidor." };
  }
  return { ok: res.ok, status: res.status, body };
}