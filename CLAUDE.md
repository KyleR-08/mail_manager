# correo-manager (mail_manager)

> Nota: la carpeta real del proyecto es `C:\proyectos\mail_manager` y el módulo Go se llama `mail_manager`. El nombre conceptual del proyecto sigue siendo **correo-manager**.

## Descripción

Sistema **multi-proveedor** de gestión de correo electrónico en Go (Golang). La aplicación es de consola y permite al usuario operar varias cuentas de correo (Gmail, Outlook/Hotmail, correos institucionales y Yahoo) desde una sola terminal, sin abrir el navegador. Es un proyecto universitario para la materia de Programación Orientada a Objetos.

## Objetivo

Aplicación de consola/terminal donde el usuario:

1. Registra una o varias cuentas de correo de **distintos proveedores**.
2. Selecciona una cuenta activa.
3. Gestiona esa cuenta (ver bandeja, leer, enviar, ver enviados, buscar) a través de un menú numerado.

Toda la interacción ocurre en texto, en la terminal.

## Proveedores soportados

| Proveedor | Tecnología | Detalle |
|---|---|---|
| **Gmail** | Gmail API v1 + OAuth2 | `google.golang.org/api/gmail/v1` y `golang.org/x/oauth2/google` |
| **Outlook / Hotmail / Live** | Microsoft Graph API + OAuth2 | `golang.org/x/oauth2` con endpoints de Microsoft Identity Platform |
| **Institucional** (`@uide.edu.ec`, etc.) | Auto-detección por dominio | Si el dominio es Google Workspace → Gmail API. Si es Microsoft 365 → Graph API. |
| **Yahoo Mail** | IMAP (fallback) | Solo como respaldo, ya que Yahoo no expone una API moderna estable |

## Concepto central — polimorfismo en Go

En `providers/provider.go` se define la interfaz **`EmailProvider`** con los métodos:

- `GetInbox()`
- `ReadMail(id string)`
- `SendMail(to, subject, body string)`
- `GetSent()`
- `SearchMail(query string)`

Cada proveedor (`gmail.go`, `outlook.go`, `yahoo.go`, `institutional.go`) implementa esta interfaz con su propia API. La UI y la sesión solo conocen `EmailProvider`, nunca el tipo concreto. Esto es **polimorfismo en Go** y es el corazón del proyecto desde el punto de vista de POO.

## Funcionalidades planeadas (MVP)

1. Login local en la aplicación
2. Registro y selección de cuentas (multi-proveedor) — persistido en `data/accounts.json`
3. Conexión OAuth2 al proveedor de la cuenta seleccionada (token guardado en `token/`)
4. Ver bandeja de entrada (remitente, asunto, fecha)
5. Leer un correo específico
6. Enviar un correo nuevo
7. Ver correos enviados
8. Buscar correos

## Arquitectura

Arquitectura en capas. Cada capa solo se comunica con las capas adyacentes:

```
Usuario
  ↓
Console UI (ui/)
  ↓
Sesión activa (session/)
  ↓
EmailProvider — interfaz (providers/)
  ↓
Implementación concreta (gmail / outlook / yahoo / institutional)
  ↓
Auth OAuth2 (auth/)  +  Cuentas registradas (accounts/ + data/accounts.json)
  ↓
APIs externas (Gmail API / Microsoft Graph / IMAP de Yahoo)
```

> ⚠️ **Arquitectura anterior reemplazada.** La estructura inicial con `handlers/` y `services/` (y un único `oauth_config.go` solo para Gmail) **ya no aplica**. Fue sustituida por la arquitectura multi-proveedor descrita arriba.

## Estructura del proyecto

```
mail_manager/
├── main.go                          # Punto de entrada
├── providers/
│   ├── provider.go                  # Interfaz EmailProvider (contrato común)
│   ├── gmail.go                     # Implementación Gmail (Gmail API v1)
│   ├── outlook.go                   # Implementación Outlook (Microsoft Graph)
│   ├── yahoo.go                     # Implementación Yahoo (IMAP fallback)
│   └── institutional.go             # Auto-detección por dominio (Workspace / M365)
├── accounts/
│   ├── account.go                   # Modelo de una cuenta registrada
│   └── manager.go                   # Cargar / guardar / listar / agregar cuentas
├── session/
│   └── session.go                   # Cuenta y proveedor actualmente activos
├── auth/
│   ├── google_auth.go               # OAuth2 contra Google (Gmail / Workspace)
│   └── microsoft_auth.go            # OAuth2 contra Microsoft Identity Platform
├── ui/
│   └── menu.go                      # Menú numerado de consola
├── data/
│   └── accounts.json                # Cuentas registradas (ignorado por git)
└── token/                           # Tokens OAuth2 por cuenta (ignorado por git)
```

## Paquetes a usar

- Estándar: `net/http`, `fmt`, `bufio`, `os`, `encoding/json`, `strings`, `errors`
- OAuth2: `golang.org/x/oauth2`, `golang.org/x/oauth2/google`, `golang.org/x/oauth2/microsoft`
- Google: `google.golang.org/api/gmail/v1`
- Microsoft Graph: llamadas REST con `net/http` (o un cliente Graph para Go)
- IMAP (Yahoo): `github.com/emersion/go-imap` (cuando se implemente)

> Nota: las dependencias se incorporan a `go.mod` cuando se implementa cada proveedor. El esqueleto actual no las incluye porque ningún archivo las importa todavía (Go las quitaría con `go mod tidy`).

## Reglas

- Comentarios en **español** en cada línea importante
- Cada archivo tiene **una sola responsabilidad**
- Nunca mezclar lógica de un proveedor con la lógica de UI
- `data/accounts.json`, `token/` y `credentials*.json` **deben estar en `.gitignore`**
- El código debe ser **amigable para principiantes**
- Editor recomendado: **VS Code**

## Equipo

- Kyle Reinoso
- Ariel Esparza
- Alejandro Zambrano

## Bitácora de progreso

### 2026-04-27 — Esqueleto inicial (arquitectura antigua, reemplazada)
- **Construido:** estructura inicial con `handlers/`, `services/`, `auth/oauth_config.go` y `ui/menu.go` orientada solo a Gmail. `go mod init mail_manager`. `.gitignore` para `token/token.json`. `CLAUDE.md` y `README.md`.
- **Estado:** ❌ **Reemplazado el 2026-05-04** por la arquitectura multi-proveedor.

### 2026-05-04 — Refactor a arquitectura multi-proveedor
- **Construido:** nueva estructura por capas con polimorfismo basado en la interfaz `EmailProvider`.
- **Eliminado:** `handlers/`, `services/`, `auth/oauth_config.go`.
- **Creado:**
  - `providers/provider.go`, `providers/gmail.go`, `providers/outlook.go`, `providers/yahoo.go`, `providers/institutional.go`
  - `accounts/account.go`, `accounts/manager.go`
  - `session/session.go`
  - `data/accounts.json` (vacío `[]`)
  - `auth/google_auth.go`, `auth/microsoft_auth.go`
- **Actualizado:** `main.go` con nuevo comentario de propósito, `ui/menu.go` adaptado a sesión + EmailProvider, `.gitignore` ampliado para tokens de varios proveedores y `data/accounts.json`, `CLAUDE.md` y `README.md` rescritos.
- **Cada archivo `.go`:** solo declaración de paquete + comentario en español. **Aún no hay lógica.**
- **Siguiente paso:** implementar la interfaz `EmailProvider` en `providers/provider.go`, luego el flujo OAuth2 en `auth/google_auth.go` y la primera implementación concreta en `providers/gmail.go`. Después construir `accounts/manager.go` y el menú real en `ui/menu.go`.
