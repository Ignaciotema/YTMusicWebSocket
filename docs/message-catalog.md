# Catálogo de mensajes — YTMusicWebSocket

Este documento lista **todos los mensajes** que circulan por el WebSocket, quién los puede enviar/recibir, y sirve de **plantilla** para documentar mensajes nuevos a medida que se agreguen. Para entender el protocolo general y el ciclo de vida de la conexión, ver [`websocket-api.md`](./websocket-api.md).

## Componentes / roles del sistema

| Componente          | Rol en el WS  | Archivo                                              |
|---------------------|---------------|-------------------------------------------------------|
| Extensión de Chrome  | `player`      | `chrome-extension/background.js`, `chrome-extension/content.js` |
| Controller (cliente genérico, ej. futuro bot de Telegram) | `controller`  | `client/telegramBot/` (aún no integrado) |
| `ConnectionManager`  | servidor — enruta mensajes | `internal/transport/websocketHandler/connectionManager.go` |
| `Dispatcher`         | servidor — interpreta comandos | `internal/dispatcher/dispatcher.go` |
| `player.Module`      | servidor — arma comandos hacia el player | `internal/player/player.go` |

## Convención de mensaje

Todo mensaje es un JSON con esta forma:

```json
{ "type": "<string>", "data": <string | objeto> }
```

- `data` es **string** en mensajes entrantes (cliente → servidor).
- `data` es **string u objeto** en mensajes salientes (servidor → cliente), según el mensaje.

## Índice de mensajes

| `type`      | Dirección              | Emisor(es)         | Receptor(es)       | Estado        |
|-------------|------------------------|--------------------|---------------------|---------------|
| `register`  | cliente → servidor     | player, controller | `ConnectionManager` | ✅ implementado |
| `command` (entrante) | cliente → servidor | controller     | `Dispatcher`        | ✅ implementado |
| `command` (saliente) | servidor → cliente | `player.Module`   | player              | ✅ implementado |
| *(broadcast a controllers)* | servidor → cliente | `ConnectionManager.Event` | todos los controllers | ⚠️ mecanismo existe, sin ningún mensaje concreto emitido todavía |

---

## `register`

**Dirección:** cliente → servidor
**Emisores:** player, controller
**Receptor:** `ConnectionManager.Register` (`connectionManager.go:52`)
**Cuándo se envía:** obligatoriamente como **primer mensaje** de toda conexión WS. Si no es de este tipo, el servidor cierra la conexión.

**Payload:**
```json
{ "type": "register", "data": "player" }
```
```json
{ "type": "register", "data": "controller" }
```

- `data == "player"` → se guarda como `ConnectionManager.PlayerConn` (reemplaza cualquier player previo).
- `data != "player"` (cualquier otro string) → se agrega a `controllerConns` con un `Id` (uuid) generado por el servidor.

**Respuesta del servidor:** ninguna. No hay ack de registro; el cliente asume éxito si el socket sigue abierto.

---

## `command` (entrante — controller → servidor)

**Dirección:** cliente → servidor
**Emisor:** controller
**Receptor:** `ConnectionManager.HandleMsg` → `Dispatcher.DispatchCommand` (`connectionManager.go:41`, `dispatcher.go:19`)
**Cuándo se envía:** en cualquier momento después del registro, para pedir una acción sobre el player.

**Payload:**
```json
{ "type": "command", "data": "playPause" }
```

**Valores válidos de `data`:**

| valor        | acción                    |
|--------------|---------------------------|
| `playPause`  | play/pause                |
| `next`       | siguiente canción         |
| `previous`   | canción anterior          |

Valores no reconocidos se ignoran silenciosamente (no hay `default` en el `switch`).

**Efecto:** dispara el método correspondiente en `player.Module`, que arma y envía el mensaje `command` saliente al player (ver abajo).

---

## `command` (saliente — servidor → player)

**Dirección:** servidor → cliente
**Emisor:** `player.Module` (`player.go:24-34`), vía `ConnectionManager.SendMessage`
**Receptor:** la conexión registrada como player
**Cuándo se envía:** inmediatamente después de procesar un `command` entrante válido.

**Payload:**
```json
{ "type": "command", "data": { "action": "playPause" } }
```

**Valores válidos de `action`:** `playPause`, `previous`, `next` (mismos tres, generados 1:1 desde el comando entrante).

**Consumido por:** `background.js` (reenvía a la tab de `music.youtube.com`) → `content.js` (hace click en el botón real).

---

## Broadcast a controllers (`ConnectionManager.Event`)

**Dirección:** servidor → cliente
**Emisor:** cualquier módulo del servidor que llame a `ConnectionManager.Event(msgType, data)` (`connectionManager.go:112`)
**Receptor:** **todos** los controllers conectados (no el player)
**Estado:** el método existe y funciona (itera `controllerConns` y usa `SendMessage`), pero **nada lo invoca todavía**. Es el mecanismo pensado para, por ejemplo, notificar a los controllers el estado actual de reproducción (canción sonando, si está en pausa, etc.).

No hay `type`/payload definido aún porque no hay ningún evento real emitido.

---

## Plantilla para documentar un mensaje nuevo

Copiar este bloque y completarlo cada vez que se agregue un `type` nuevo (ya sea entrante o saliente):

```markdown
## `<type>`

**Dirección:** <cliente → servidor | servidor → cliente>
**Emisor(es):** <componente(s) que lo puede mandar>
**Receptor(es):** <componente(s) que lo procesa/consume>
**Handler:** `<archivo:línea o función>`
**Cuándo se envía:** <trigger / condición>

**Payload:**
​```json
{ "type": "<type>", "data": <forma del payload> }
​```

**Campos de `data`:**

| campo | tipo | obligatorio | descripción |
|-------|------|-------------|-------------|
|       |      |             |             |

**Valores válidos / enum (si aplica):**

| valor | acción/significado |
|-------|---------------------|

**Respuesta esperada (si aplica):** <mensaje que dispara en el otro extremo, o "ninguna">

**Estado:** <✅ implementado | 🚧 en desarrollo | ⚠️ definido pero sin uso>
```

### Checklist al agregar un mensaje nuevo

- [ ] ¿El `type` es único y no colisiona con uno existente?
- [ ] ¿Está claro quién lo puede emitir (rol: player / controller / servidor interno)?
- [ ] ¿Quién lo debe procesar, y en qué archivo se agrega el `case`/handler?
- [ ] Si es un comando nuevo hacia el player: ¿hay que agregar el botón/selector correspondiente en `content.js`?
- [ ] Si es un evento nuevo hacia los controllers: ¿qué dispara el llamado a `ConnectionManager.Event`?
- [ ] Actualizar este catálogo y, si cambia el flujo general, `websocket-api.md`.
