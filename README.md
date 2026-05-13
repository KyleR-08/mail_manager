# correo-manager

> Gestor **multi-proveedor** de correo electrónico de consola en Go. Conecta Gmail, Outlook/Hotmail, correos institucionales y Yahoo desde una sola terminal.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![Gmail API](https://img.shields.io/badge/Gmail%20API-v1-EA4335?logo=gmail&logoColor=white)
![Microsoft Graph](https://img.shields.io/badge/Microsoft%20Graph-API-0078D4?logo=microsoft&logoColor=white)
![OAuth2](https://img.shields.io/badge/OAuth2-Google%20%7C%20Microsoft-4285F4?logo=oauth&logoColor=white)
![IMAP](https://img.shields.io/badge/IMAP-Yahoo%20fallback-720E9E?logo=yahoo&logoColor=white)

---

## Descripción

**correo-manager** es una aplicación de consola escrita en Go que permite gestionar **varias cuentas de correo de distintos proveedores** desde la terminal, sin necesidad de abrir el navegador. Cada proveedor se integra mediante su API oficial (o IMAP cuando no existe API), pero la aplicación los expone a través de **una sola interfaz común**: `EmailProvider`.

Este proyecto fue desarrollado como trabajo universitario para la materia de **Programación Orientada a Objetos** por:

- Kyle Reinoso
- Ariel Esparza
- Alejandro Zambrano

Editor recomendado: **VS Code**.

---

## Proveedores soportados

| Proveedor | Integración | Librerías clave |
|---|---|---|
| **Gmail** | Gmail API v1 + OAuth2 de Google | `google.golang.org/api/gmail/v1`, `golang.org/x/oauth2/google` |
| **Outlook / Hotmail / Live** | Microsoft Graph API + OAuth2 | `golang.org/x/oauth2` con endpoints de Microsoft Identity Platform |
| **Institucional** (`@uide.edu.ec` y similares) | Auto-detección por dominio | Workspace → Gmail API · Microsoft 365 → Graph API |
| **Yahoo Mail** | IMAP (fallback) | `github.com/emersion/go-imap` |

---

## Funcionalidades

1. **Inicio de sesión local** en la aplicación.
2. **Registrar varias cuentas** (multi-proveedor) y elegir la cuenta activa.
3. **Conexión OAuth2** con el proveedor de la cuenta seleccionada.
4. **Ver bandeja de entrada** con remitente, asunto y fecha.
5. **Leer un correo específico** seleccionado del listado.
6. **Enviar un correo nuevo** desde la terminal.
7. **Ver correos enviados** desde la cuenta activa.
8. **Buscar correos** por palabra clave o remitente.

---

## Arquitectura

El sistema sigue una **arquitectura en capas con polimorfismo**. La UI nunca habla directamente con Gmail o Microsoft Graph; habla con la interfaz `EmailProvider`, y cada proveedor concreto la implementa a su manera.

```
Usuario
  ↓
Console UI (ui/)
  ↓
Sesión activa (session/)
  ↓
EmailProvider — interfaz (providers/)
  ↓
gmail.go · outlook.go · yahoo.go · institutional.go
  ↓
Auth OAuth2 (auth/)  +  Cuentas registradas (accounts/ + data/accounts.json)
  ↓
APIs externas (Gmail API · Microsoft Graph · IMAP de Yahoo)
```

### `EmailProvider` — el contrato común

Definido en [`providers/provider.go`](providers/provider.go):

```go
GetInbox()
ReadMail(id string)
SendMail(to, subject, body string)
GetSent()
SearchMail(query string)
```

Cada implementación concreta traduce estos métodos a llamadas reales contra su API. Esta separación garantiza que la UI no dependa del proveedor.

---

## Tecnologías utilizadas

- **Lenguaje:** Go (Golang)
- **APIs:** Gmail API v1, Microsoft Graph API, IMAP (fallback Yahoo)
- **Autenticación:** OAuth2 (Google y Microsoft Identity Platform)
- **Paquetes principales:**
  - `golang.org/x/oauth2`, `golang.org/x/oauth2/google`, `golang.org/x/oauth2/microsoft`
  - `google.golang.org/api/gmail/v1`
  - `net/http` (para llamadas REST a Microsoft Graph)
  - `github.com/emersion/go-imap` (Yahoo)
  - Estándar: `fmt`, `bufio`, `os`, `encoding/json`, `strings`, `errors`

---

## Instalación y uso

1. **Clonar el repositorio:**
   ```bash
   git clone <url-del-repositorio>
   cd mail_manager
   ```

2. **Obtener las credenciales OAuth2:**
   - **Google (Gmail / Workspace):** crea un proyecto en [Google Cloud Console](https://console.cloud.google.com/), habilita **Gmail API**, genera un *OAuth client ID* tipo Aplicación de escritorio y guarda el JSON como `credentials_google.json`.
   - **Microsoft (Outlook / Hotmail / M365):** registra una app en [Azure Portal](https://portal.azure.com/) → Microsoft Entra ID → App registrations, agrega los permisos `Mail.Read` y `Mail.Send` de Microsoft Graph y guarda el client ID/secret como `credentials_microsoft.json`.
   - **Yahoo:** genera una contraseña de aplicación desde la configuración de seguridad de tu cuenta Yahoo.

3. **Instalar dependencias:**
   ```bash
   go mod tidy
   ```

4. **Ejecutar la aplicación:**
   ```bash
   go run main.go
   ```

   La primera vez se abrirá el flujo OAuth2 del proveedor de la cuenta que registres. El token resultante se guarda en `token/` y la cuenta queda registrada en `data/accounts.json`.

---

## Estructura del proyecto

```
mail_manager/
├── main.go                          # Punto de entrada
├── providers/
│   ├── provider.go                  # Interfaz EmailProvider (contrato común)
│   ├── gmail.go                     # Implementación Gmail (Gmail API v1)
│   ├── outlook.go                   # Implementación Outlook (Microsoft Graph)
│   ├── yahoo.go                     # Implementación Yahoo (IMAP fallback)
│   └── institutional.go             # Auto-detección por dominio
├── accounts/
│   ├── account.go                   # Modelo de una cuenta registrada
│   └── manager.go                   # Cargar / guardar / listar / agregar cuentas
├── session/
│   └── session.go                   # Cuenta y proveedor activos en memoria
├── auth/
│   ├── google_auth.go               # OAuth2 contra Google
│   └── microsoft_auth.go            # OAuth2 contra Microsoft Identity Platform
├── ui/
│   └── menu.go                      # Menú numerado de consola
├── data/
│   └── accounts.json                # Cuentas registradas (ignorado por git)
└── token/                           # Tokens OAuth2 por cuenta (ignorado por git)
```

---

## Estado del proyecto

-  **Etapa 1 — Planeación inicial (solo Gmail):** completada y reemplazada.
-  **Etapa 2 — Refactor a arquitectura multi-proveedor:** completada (esqueleto listo, sin lógica todavía).
-  **Etapa 3 — Implementación:** en progreso (interfaz, OAuth2 de Google, primer proveedor concreto).

---

> **Nota de seguridad:** los archivos `credentials_google.json`, `credentials_microsoft.json`, todo `token/` y `data/accounts.json` contienen información sensible y **están excluidos del repositorio** mediante `.gitignore`. Cada usuario debe generar los suyos de forma local.
