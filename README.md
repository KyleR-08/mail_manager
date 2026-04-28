# correo-manager

> Gestor de correo electrónico de consola en Go que se conecta a Gmail mediante OAuth2.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Gmail API](https://img.shields.io/badge/Gmail%20API-v1-EA4335?logo=gmail&logoColor=white)
![OAuth2](https://img.shields.io/badge/OAuth2-Google-4285F4?logo=google&logoColor=white)

---

## Descripción

**correo-manager** es una aplicación de consola escrita en Go que permite gestionar una cuenta de Gmail directamente desde la terminal, sin necesidad de abrir el navegador. La aplicación se autentica mediante OAuth2 contra la API oficial de Gmail y ofrece un menú numerado para ejecutar todas las acciones.

Este proyecto fue desarrollado como trabajo universitario para la materia de **Programación Orientada a Objetos** por:

- Kyle Reinoso
- Ariel Esparza
- Alejandro Zambrano

---

## Funcionalidades

1. **Inicio de sesión local** en la aplicación.
2. **Conexión con Gmail vía OAuth2** (el token se guarda en `token/token.json`).
3. **Ver bandeja de entrada** con remitente, asunto y fecha.
4. **Leer un correo específico** seleccionado del listado.
5. **Enviar un correo nuevo** desde la terminal.
6. **Ver correos enviados** desde la cuenta.
7. **Buscar correos** por palabra clave o remitente.

---

## Arquitectura

El sistema sigue una **arquitectura en capas**, donde cada capa solo se comunica con las capas adyacentes:

```
Usuario  →  Consola (UI)  →  Backend Go  →  Gmail API  →  Servidores de Google
```

- **Consola (UI):** muestra el menú y captura las opciones del usuario.
- **Backend Go:** contiene los handlers y servicios que aplican la lógica.
- **Gmail API:** capa externa con la que se comunica el servicio de Gmail.
- **Servidores de Google:** infraestructura que aloja los correos.

Esta separación garantiza que ninguna capa mezcle responsabilidades de otra.

---

## Tecnologías utilizadas

- **Lenguaje:** Go (Golang)
- **API:** Gmail API v1
- **Autenticación:** OAuth2 de Google
- **Paquetes principales:**
  - `golang.org/x/oauth2`
  - `golang.org/x/oauth2/google`
  - `google.golang.org/api/gmail/v1`
  - `net/http`, `fmt`, `bufio`, `os`, `encoding/json` (librería estándar)

---

## Instalación y uso

1. **Clonar el repositorio:**
   ```bash
   git clone <url-del-repositorio>
   cd mail_manager
   ```

2. **Obtener las credenciales de Google:**
   - Entra a [Google Cloud Console](https://console.cloud.google.com/).
   - Crea un proyecto y habilita la **Gmail API**.
   - Genera credenciales de tipo *OAuth client ID* (aplicación de escritorio).
   - Descarga el archivo y guárdalo como `credentials.json` en la raíz del proyecto.

3. **Instalar dependencias:**
   ```bash
   go mod tidy
   ```

4. **Ejecutar la aplicación:**
   ```bash
   go run main.go
   ```

   La primera vez se abrirá el flujo OAuth2 para autorizar el acceso a la cuenta. El token resultante se guardará automáticamente en `token/token.json` para próximas ejecuciones.

---

## Estructura del proyecto

```
mail_manager/
├── main.go                      # Punto de entrada de la aplicación
├── handlers/
│   ├── auth.go                  # Coordina el login local y el flujo OAuth2
│   ├── gmail.go                 # Recibe peticiones de la UI y llama al servicio
│   └── email.go                 # Manejo de envío y composición de correos
├── services/
│   └── gmail_service.go         # Única capa que habla directamente con la Gmail API
├── auth/
│   └── oauth_config.go          # Configuración de OAuth2 (scopes, client, token)
├── ui/
│   └── menu.go                  # Menú de consola y captura de opciones del usuario
└── token/
    └── token.json               # Token OAuth2 generado tras autenticarse (ignorado por git)
```

---

## Estado del proyecto

- ✅ **Etapa 1 — Planeación:** completada (arquitectura, estructura, esqueleto de archivos).
- 🚧 **Etapa 2 — Implementación:** en progreso (OAuth2, menú y operaciones contra Gmail).

---

> **Nota de seguridad:** los archivos `credentials.json` y `token/token.json` contienen información sensible y **están excluidos del repositorio** mediante `.gitignore`. Cada usuario debe generar los suyos de forma local.
