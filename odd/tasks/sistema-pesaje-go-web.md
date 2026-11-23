# Sistema pesaje Go + Vite (nube)

Servidor web compatible con `pesaje-go` (auth token, `/api/pesajes`, `/api/sync/*`, `ultimo-peso`) y UI Vite con login estilo sistema completo.

## Specs

S1. Stack: Go (`cmd/web`) + SQLite servidor + Vite; escritorio Wails sigue siendo cliente de báscula.
S2. Menú web (fases): Inicio, Usuarios, Empresa, Choferes, Proveedores, Pesaje/catálogo, Reportes, Config.
S3. Contrato escritorio: `POST /api/auth/token` (`usuario`,`contrasena`); cookies `access_token`/`refresh_token`; `POST /api/pesajes`; `GET|POST /api/pesajes/ultimo-peso`; `GET /api/sync/{recurso}?since=`.
S4. Login Vite réplica visual de `sistema-pesaje-next` (slate, card centrada, branding empresa).

## Tasks

- T1: ODD + go.mod + cmd/web scaffold — route: inline — commit: pending
- T2: DB embed + seed admin + JWT auth — route: inline — commit: pending
- T3: ultimo-peso + POST /api/pesajes — route: inline — commit: pending
- T4: sync GET stub + tests — route: inline — commit: pending
- T5: Vite login + shell menú — route: inline — commit: pending

## Log

L1. arrancarmos — Go nube + Vite, misma funcionalidad que sistema completo; pesaje recibe peso del escritorio; SQLite servidor.
