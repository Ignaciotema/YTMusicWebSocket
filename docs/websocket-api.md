# API WebSocket — YTMusicWebSocket

Documentación del protocolo WebSocket que conecta el **player** (la extensión de Chrome corriendo sobre `music.youtube.com`) con uno o más **controllers** (clientes que envían comandos, ej. un bot de Telegram).

## Endpoint

```
ws://<host>:8080/ws
```

Definido en `cmd/main.go:27` y servido por `websocketHandler.HandleWebSocket` (`internal/transport/websocketHandler/websocketHandler.go:31`).

No hay autenticación ni validación de origen (`CheckOrigin` siempre devuelve `true` en `websocketHandler.go:19`), así que cualquier cliente en la red puede conectarse.

## Arquitectura general

```
┌────────────┐   register: player     ┌──────────────────┐
│  Extensión │ ─────────────────────▶ │                   │
│  (player)  │ ◀───────────────────── │ ConnectionManager │
└────────────┘   command (evento)     │                   │
                                       │  - PlayerConn     │
┌────────────┐   register: controller │  - controllerConns│
│ Controller │ ─────────────────────▶ │                   │
│ (Telegram, │                        └─────────┬─────────┘
│  etc.)     │   command: playPause/next/...     │
└────────────┘ ─────────────────────▶     Dispatcher
                                              │
                                          player.Module
                                     (arma el mensaje y lo
                                      manda de vuelta al
                                      player vía WS)
```

Piezas clave:

- **`ConnectionManager`** (`internal/transport/websocketHandler/connectionManager.go`): mantiene la conexión del player (`PlayerConn`, una sola) y la lista de controllers conectados (`controllerConns`, muchos).
- **`Dispatcher`** (`internal/dispatcher/dispatcher.go`): traduce un comando en texto (`"playPause"`, `"previous"`, `"next"`) en una llamada al `player.Module`.
- **`player.Module`** (`internal/player/player.go`): arma el mensaje de salida (`{"type": "command", "data": {"action": ...}}`) y lo envía por WS a la conexión registrada como player.

Solo hay una conexión de player activa a la vez: cada nuevo registro `player` pisa la anterior en `ConnectionManager.PlayerConn`.

## Ciclo de vida de una conexión

1. El cliente abre el WebSocket contra `/ws`.
2. **El primer mensaje que se lee debe ser de registro** (`websocketHandler.go:43`). Si no lo es, o no puede parsearse, el servidor cierra la conexión.
3. Una vez registrado, el servidor queda en loop leyendo mensajes (`websocketHandler.go:52-60`). Solo se procesan mensajes de tipo `"command"`; cualquier otro tipo es ignorado silenciosamente.
4. Si `ReadMessage` falla (cliente desconectado, error de red), el loop corta y la conexión se cierra (`defer CloseConnection(conn)`).

No hay heartbeat/ping-pong ni reconexión automática por parte del servidor — eso queda a cargo del cliente (la extensión, por ejemplo, no reintenta si el socket se cierra).

## Formato de los mensajes

Todos los mensajes son JSON de texto (`websocket.TextMessage`) con esta forma genérica:

```json
{
  "type": "<tipo>",
  "data": "<payload>"
}
```

`data` se tipa como `string` en los mensajes **entrantes** (`IncomingMessage`, `connectionManager.go:22-25`) y como `any` en los mensajes **salientes** del servidor (`Message`, `connectionManager.go:91-94`), ya que ahí sí se manda un objeto (`map[string]string`).

### 1. Registro (`type: "register"`) — cliente → servidor

Primer mensaje obligatorio de toda conexión.

```json
{ "type": "register", "data": "player" }
```

```json
{ "type": "register", "data": "controller" }
```

- Si `data == "player"` → la conexión se guarda como `PlayerConn` (única, se reemplaza si ya había una).
- Cualquier otro valor de `data` → se registra como **controller** y se agrega a la lista `controllerConns` con un `Id` (uuid) generado por el servidor.
- Cualquier mensaje de registro con `type` distinto de `"register"` hace que `Register` devuelva `false` y el servidor cierre la conexión.

### 2. Comando (`type: "command"`) — controller → servidor

```json
{ "type": "command", "data": "playPause" }
```

Valores válidos de `data` (definidos en `dispatcher.go:23-29`):

| comando     | acción                          |
|-------------|----------------------------------|
| `playPause` | play/pause del reproductor       |
| `next`      | siguiente canción                |
| `previous`  | canción anterior                 |

Cualquier otro valor no hace nada (el `switch` del dispatcher no tiene `default`).

Este mensaje lo procesa `ConnectionManager.HandleMsg` → `Dispatcher.DispatchCommand` → método correspondiente en `player.Module`.

### 3. Comando reenviado al player — servidor → player

Cuando el dispatcher ejecuta una acción, `player.Module` arma y envía este mensaje **únicamente a la conexión registrada como player**:

```json
{ "type": "command", "data": { "action": "playPause" } }
```

(`action` puede ser `"playPause"`, `"previous"` o `"next"`, ver `player.go:24-34`).

La extensión de Chrome escucha este mensaje en `background.js:18-29`, lo reenvía a la pestaña de `music.youtube.com` vía `chrome.tabs.sendMessage`, y `content.js:36-51` hace click en el botón correspondiente del reproductor web.

### 4. Eventos broadcast — servidor → todos los controllers

`ConnectionManager.Event(msgType, data)` (`connectionManager.go:112-116`) envía un mensaje a **todos** los controllers conectados (no al player). Actualmente no hay ningún llamado a `Event` en el código — es un mecanismo disponible para, por ejemplo, notificar a todos los controllers el estado actual de reproducción, pero todavía no está cableado a ninguna fuente de eventos.

## Roles resumidos

| Rol          | Cuántas conexiones | Quién lo usa hoy                  | Qué manda                      | Qué recibe                          |
|--------------|---------------------|------------------------------------|----------------------------------|--------------------------------------|
| `player`     | 1 (la última que se registra) | Extensión de Chrome (`background.js`) | `register: player`             | `command` con la acción a ejecutar   |
| `controller` | N                    | Ninguno implementado aún (el bot de Telegram en `client/telegramBot/` es solo un tutorial sin integrar) | `register: controller`, luego `command: playPause/next/previous` | Nada por ahora (`Event` no se usa)   |

## Endpoint HTTP auxiliar

Además del WS hay un endpoint HTTP en `/` (`internal/transport/httpHandler/httpHandler.go`):

```
GET /?command=<comando>
```

Actualmente solo valida que el query param `command` no esté vacío y responde `200 OK` — la línea que despacharía el comando al dispatcher está comentada (`httpHandler.go:25`), así que este endpoint no tiene efecto real todavía.

## Ejemplo de sesión completa

**Extensión (player):**
```json
→ {"type": "register", "data": "player"}
```

**Controller (ej. futuro bot):**
```json
→ {"type": "register", "data": "controller"}
→ {"type": "command", "data": "next"}
```

**Servidor → player:**
```json
← {"type": "command", "data": {"action": "next"}}
```

## Puntos a tener en cuenta / limitaciones actuales

- **Sin autenticación**: cualquiera que alcance el puerto 8080 puede registrarse como player o controller.
- **Un solo player**: si dos extensiones se registran como `player`, la segunda reemplaza a la primera silenciosamente.
- **Sin manejo de errores hacia el cliente**: mensajes mal formados, comandos desconocidos o fallos de registro se logean en el servidor pero no generan una respuesta de error al cliente.
- **`Event` no está en uso**: la infraestructura para notificar a los controllers existe pero ningún flujo la dispara todavía.
- **Sin reconexión/heartbeat**: la responsabilidad de reconectar ante un cierre de socket es del cliente.
