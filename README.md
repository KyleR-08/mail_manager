# mail_manager

> Microservicio HTTP en **Go** para gestionar cuentas de correo de varios proveedores. Proyecto final de **Programación Orientada a Objetos**.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-HTTP%20Framework-008ECF?logo=go&logoColor=white)
![Gmail API](https://img.shields.io/badge/Gmail%20API-v1-EA4335?logo=gmail&logoColor=white)
![Microsoft Graph](https://img.shields.io/badge/Microsoft%20Graph-API-0078D4?logo=microsoft&logoColor=white)
![OAuth2](https://img.shields.io/badge/OAuth2-Google%20%7C%20Microsoft-4285F4?logo=oauth&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)

---

## Descripción

**mail_manager** es un **servidor HTTP** escrito en Go con el framework **Gin** que permite gestionar varias cuentas de correo de distintos proveedores desde una interfaz web. Cada proveedor se integra mediante su API oficial, pero la aplicación los expone a través de **una sola interfaz común**: `EmailProvider`.

La aplicación arranca en `http://localhost:8080` y sirve:

- Una interfaz web simple en `/` (HTML + CSS + JavaScript puro).
- Una API REST en JSON para integraciones externas.
- Endpoints de OAuth2 con redirección de callback para conectar Gmail y Outlook sin pegar códigos a mano.

El proyecto fue desarrollado como trabajo final para la materia de **Programación Orientada a Objetos**.

---

## Integrantes

- **Kyle Reinoso**
- **Ariel Esparza**
- **Alejandro Zambrano**

---

## Proveedores soportados

| Proveedor | Dominios | Tecnología | Estado |
|---|---|---|---|
| **Gmail** | `@gmail.com` | Gmail API v1 + OAuth2 de Google | ✅ **Funcional** |
| **Outlook / Hotmail / Live** | `@hotmail.com`, `@outlook.com`, `@live.com` | Microsoft Graph API + OAuth2 | 🟡 Implementado — requiere credenciales de Azure |
| **Institucional UIDE** | `@uide.edu.ec` | Microsoft Graph API + OAuth2 (tenant UIDE) | 🟡 Implementado — requiere credenciales de Azure |

> Las implementaciones de Outlook e Institucional UIDE están **completas a nivel de código**, pero requieren registrar la app en **Azure Portal** y reemplazar las constantes placeholder en `main.go` antes de poder autenticar.

---

## Concepto OOP principal — Polimorfismo

El corazón del proyecto es la interfaz **`EmailProvider`** definida en [`providers/provider.go`](providers/provider.go):

```go
type EmailProvider interface {
    GetInbox() ([]Mail, error)
    ReadMail(id string) (Mail, error)
    SendMail(to, subject, body string) error
    GetSent() ([]Mail, error)
    SearchMail(query string) ([]Mail, error)
}
```

Cada proveedor (`gmail.go`, `outlook.go`, `institutional.go`) implementa esta misma interfaz **con su propia lógica interna**. Los handlers HTTP solo conocen la interfaz `EmailProvider`, nunca el tipo concreto:

```go
var proveedorActivo providers.EmailProvider
proveedorActivo = providers.NewGmailProvider(client) // o NewOutlookProvider, o NewInstitutionalProvider
proveedorActivo.GetInbox()                           // misma llamada, comportamiento distinto
```

Eso es **polimorfismo en Go**: una sola firma, múltiples implementaciones.

---

## Endpoints REST

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/health` | Chequeo de salud (`{"status":"ok"}`) |
| `GET` | `/accounts` | Lista las cuentas registradas y la activa |
| `POST` | `/accounts` | Agrega una cuenta — body: `{"email":"...","display_name":"..."}` |
| `POST` | `/accounts/:email/activate` | Marca una cuenta como activa |
| `GET` | `/auth/google/start` | Redirige a la URL de login de Google |
| `GET` | `/auth/google/callback` | Recibe el código de Google y guarda el token |
| `GET` | `/auth/microsoft/start` | Redirige a la URL de login de Microsoft |
| `GET` | `/auth/microsoft/callback` | Recibe el código de Microsoft y guarda el token |
| `GET` | `/inbox` | Bandeja de entrada (usa la cuenta activa) |
| `POST` | `/send` | Envía un correo — body: `{"to":"...","subject":"...","body":"..."}` |
| `GET` | `/sent` | Correos enviados |
| `GET` | `/search?q=texto` | Busca correos por texto |

---

## Estructura del proyecto

```
mail_manager/
├── main.go                  (servidor Gin + handlers HTTP)
├── providers/
│   ├── provider.go          (interfaz EmailProvider + struct Mail)
│   ├── gmail.go             (implementación Gmail API v1)
│   ├── outlook.go           (implementación Microsoft Graph)
│   └── institutional.go     (implementación UIDE - Microsoft Graph)
├── auth/
│   ├── google_auth.go       (OAuth2 Google con callback HTTP)
│   └── microsoft_auth.go    (OAuth2 Microsoft con callback HTTP)
├── accounts/
│   ├── account.go           (struct Account)
│   └── manager.go           (cargar/guardar cuentas)
├── static/
│   └── index.html           (interfaz web)
├── data/
│   └── accounts.json        (no se sube a GitHub)
├── Dockerfile
└── docker-compose.yml
```

---

## Requisitos previos

- **Go 1.21 o superior** (para correrlo sin Docker).
- **Docker** y **docker-compose** (para correrlo containerizado).
- **Para Gmail:** descargar `credentials.json` desde [Google Cloud Console](https://console.cloud.google.com/), habilitar Gmail API, crear un OAuth client ID tipo *Aplicación Web* y registrar el **Authorized redirect URI** `http://localhost:8080/auth/google/callback`. Colocar el archivo en la **raíz del proyecto**.
- **Para Outlook / UIDE:** registrar la app en [Azure Portal](https://portal.azure.com/) → Microsoft Entra ID → App registrations, otorgar los permisos `Mail.Read`, `Mail.Send` y `offline_access` de Microsoft Graph, registrar el redirect URI `http://localhost:8080/auth/microsoft/callback` y reemplazar las constantes en `main.go`:
  - `outlookClientID`, `outlookClientSecret`
  - `uideClientID`, `uideClientSecret`, `uideTenantID`

---

## Cómo ejecutar (modo desarrollo)

```bash
git clone https://github.com/KyleR-08/mail_manager
cd mail_manager
go run main.go
```

Abre [http://localhost:8080](http://localhost:8080) en el navegador. Desde ahí puedes:

1. Agregar una cuenta de correo.
2. Activarla.
3. Pulsar **Conectar Gmail** o **Conectar Outlook / UIDE** — el servidor te redirige al consentimiento del proveedor.
4. Al terminar, el navegador vuelve a `/` y la cuenta queda lista para usar.

---

## Cómo ejecutar con Docker

```bash
docker compose up --build
```

Esto construye la imagen, monta `data/`, `token/` y `credentials.json` como volúmenes y expone el puerto 8080. La app queda disponible en [http://localhost:8080](http://localhost:8080).

Para detenerla:

```bash
docker compose down
```

---

## Archivos ignorados por `.gitignore`

Los siguientes archivos contienen información sensible o local y **no se suben a GitHub**:

- `credentials.json` — credenciales OAuth2 de la app de Google.
- `token/` — tokens OAuth2 emitidos por proveedor para cada cuenta.
- `data/accounts.json` — lista local de cuentas registradas (incluye correos del usuario).
- `CLAUDE.md` y `.claude/` — bitácora interna y configuración local de Claude.

Cada integrante debe generar sus propias credenciales en local.
