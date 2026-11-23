# Sistema de pesaje — servidor web (Go + Vite)

API compatible con **pesaje-go** (escritorio Wails): token, sync, pesajes y peso en vivo.

## Desarrollo

```bash
cd /Users/jaru/dev/sistema-pesaje-go-web
go run ./cmd/web
```

Por defecto escucha `:8080` y usa `./data/servidor.db`. En el primer arranque imprime la contraseña del admin `admin@sistema.com`.

### Frontend

```bash
cd web
npm install
npm run dev
```

Vite (puerto 5173) hace proxy de `/api` al servidor Go.

### Producción

```bash
cd web && npm run build
go run ./cmd/web   # sirve web/dist si existe
```

### Variables

| Variable | Default |
|----------|---------|
| `LISTEN_ADDR` | `:8080` |
| `DATABASE_PATH` | `./data/servidor.db` |
| `JWT_SECRET` | (cambiar en producción) |
| `WEB_DIST` | `./web/dist` |

## Escritorio

En pesaje-go configurá la URL del servidor (ej. `http://localhost:8080`) y las credenciales del usuario web.
