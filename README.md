# mail_manager

> Gestor de correos electrónicos de consola escrito en **Go**. Proyecto final de **Programación Orientada a Objetos**.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Gmail API](https://img.shields.io/badge/Gmail%20API-v1-EA4335?logo=gmail&logoColor=white)
![Microsoft Graph](https://img.shields.io/badge/Microsoft%20Graph-API-0078D4?logo=microsoft&logoColor=white)
![OAuth2](https://img.shields.io/badge/OAuth2-Google%20%7C%20Microsoft-4285F4?logo=oauth&logoColor=white)

---

## Descripción

**mail_manager** es una aplicación de consola escrita en Go que permite gestionar varias cuentas de correo de distintos proveedores desde la terminal, sin necesidad de abrir el navegador. Cada proveedor se integra mediante su API oficial, pero la aplicación los expone a través de **una sola interfaz común**: `EmailProvider`.

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

> Las implementaciones de Outlook e Institucional UIDE están **completas a nivel de código**, pero requieren registrar la app en **Azure Portal** y reemplazar las constantes placeholder en `main.go` (`outlookClientID`, `outlookClientSecret`, `uideClientID`, `uideClientSecret`, `uideTenantID`) por las credenciales reales antes de poder autenticar.

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

Cada proveedor (`gmail.go`, `outlook.go`, `institutional.go`) implementa esta misma interfaz **con su propia lógica interna** — Gmail llama al endpoint REST de Gmail API v1, mientras que Outlook e Institucional llaman a Microsoft Graph. La UI y `main.go` solo conocen la interfaz `EmailProvider`, nunca el tipo concreto:

```go
var proveedor providers.EmailProvider
proveedor = providers.NewGmailProvider(client) // o NewOutlookProvider, o NewInstitutionalProvider
proveedor.GetInbox()                           // misma llamada, comportamiento distinto
```

Eso es **polimorfismo en Go**: una sola firma, múltiples implementaciones.

---

## Estructura del proyecto

```
mail_manager/
├── main.go
├── providers/
│   ├── provider.go       (interfaz EmailProvider + struct Mail)
│   ├── gmail.go          (implementación Gmail API v1)
│   ├── outlook.go        (implementación Microsoft Graph)
│   └── institutional.go  (implementación UIDE - Microsoft Graph)
├── auth/
│   ├── google_auth.go    (OAuth2 Google)
│   └── microsoft_auth.go (OAuth2 Microsoft)
├── accounts/
│   ├── account.go        (struct Account)
│   └── manager.go        (cargar/guardar cuentas)
├── ui/
│   └── menu.go           (menú de terminal)
└── data/
    └── accounts.json     (no se sube a GitHub)
```

---

## Funcionalidades del menú

Al iniciar la app, el usuario registra o selecciona una cuenta y entra al menú principal:

1. **Ver bandeja de entrada**
2. **Leer correo** (por id)
3. **Enviar correo**
4. **Ver enviados**
5. **Buscar correo**
6. **Agregar cuenta**
7. **Salir**

---

## Requisitos previos

- **Go 1.21 o superior**
- **Para Gmail:** descargar `credentials.json` desde [Google Cloud Console](https://console.cloud.google.com/) (habilitar Gmail API, crear OAuth client ID tipo *Aplicación de escritorio*) y colocarlo en la **raíz del proyecto**.
- **Para Outlook / UIDE:** registrar la app en [Azure Portal](https://portal.azure.com/) → Microsoft Entra ID → App registrations, otorgar los permisos `Mail.Read`, `Mail.Send` y `offline_access` de Microsoft Graph, y reemplazar las constantes correspondientes en `main.go`:
  - `outlookClientID`, `outlookClientSecret`
  - `uideClientID`, `uideClientSecret`, `uideTenantID`

---

## Cómo ejecutar

```bash
git clone https://github.com/KyleR-08/mail_manager
cd mail_manager
go run main.go
```

La primera vez se abre el flujo OAuth2 del proveedor de la cuenta que registres. El token resultante se guarda en `token/` y la cuenta queda registrada en `data/accounts.json` para no tener que pedirlo otra vez.

---

## Archivos ignorados por `.gitignore`

Los siguientes archivos contienen información sensible o local y **no se suben a GitHub**:

- `credentials.json` — credenciales OAuth2 de la app de Google.
- `token/` — tokens OAuth2 emitidos por proveedor para cada cuenta.
- `data/accounts.json` — lista local de cuentas registradas (incluye correos del usuario).
- `CLAUDE.md` — bitácora interna de desarrollo.

Cada integrante debe generar sus propias credenciales en local.
