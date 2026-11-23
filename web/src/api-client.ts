export type ApiBody<T = unknown> = {
  success?: boolean;
  data?: T;
  message?: string;
  [k: string]: unknown;
};

export type FetchApiResult<T = unknown> = {
  ok: boolean;
  status: number;
  body: ApiBody<T>;
  /** Sesión invalidada (token/cookie); conviene ir al login. */
  sessionExpired?: boolean;
};

type FetchApiOpts = RequestInit & { _authRetry?: boolean };

let refreshInFlight: Promise<boolean> | null = null;

async function refreshAccessToken(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    try {
      const res = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "include",
      });
      const ct = res.headers.get("content-type") || "";
      if (!ct.includes("application/json")) return false;
      const body = (await res.json()) as ApiBody;
      return res.ok && body.success === true;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

function clearSessionCliente(): void {
  try {
    localStorage.removeItem("pesaje_user");
  } catch {
    /* ignore */
  }
}

/** Respuesta JSON del backend Go; renueva cookie si el access_token venció (15 min). */
export async function fetchApi<T = unknown>(
  url: string,
  init?: FetchApiOpts,
): Promise<FetchApiResult<T>> {
  const { _authRetry, ...reqInit } = init ?? {};
  const res = await fetch(url, { credentials: "include", ...reqInit });

  if (res.status === 401 && !_authRetry) {
    const renewed = await refreshAccessToken();
    if (renewed) {
      return fetchApi<T>(url, { ...init, _authRetry: true });
    }
    clearSessionCliente();
    return {
      ok: false,
      status: 401,
      sessionExpired: true,
      body: {
        success: false,
        message: "Sesión expirada. Volvé a ingresar.",
      },
    };
  }

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
  let body: ApiBody<T>;
  try {
    body = await res.json();
  } catch {
    body = { success: false, message: "Respuesta inválida del servidor." };
  }

  if (res.status === 401) {
    clearSessionCliente();
    return {
      ok: false,
      status: 401,
      sessionExpired: true,
      body: {
        ...body,
        success: false,
        message: body.message || "Sesión expirada. Volvé a ingresar.",
      },
    };
  }

  return { ok: res.ok, status: res.status, body };
}
