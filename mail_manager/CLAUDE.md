# correo-manager (mail_manager)

> Nota: la carpeta real del proyecto es `C:\proyectos\mail_manager` y el módulo Go se llama `mail_manager`. El nombre conceptual del proyecto sigue siendo **correo-manager**.

## Descripción

Sistema de gestión de correo electrónico en Go (Golang) que se conecta a Gmail mediante la API oficial de Gmail y autenticación OAuth2. Es un proyecto universitario para la materia de Programación Orientada a Objetos.

## Objetivo

Aplicación de consola/terminal donde el usuario puede gestionar su cuenta de Gmail sin abrir el navegador. Toda la interacción se hace con un menú numerado de texto en la terminal.

## Funcionalidades planeadas (MVP)

1. Login local en la aplicación
2. Conexión a Gmail vía OAuth2 (token guardado en `token/token.json`)
3. Ver bandeja de entrada (remitente, asunto, fecha)
4. Leer un correo específico
5. Enviar un correo nuevo
6. Ver correos enviados
7. Buscar correos

## Arquitectura

Arquitectura en capas. Cada capa solo se comunica con las capas adyacentes:

```
Usuario → Console UI → Backend Go → Gmail API → Servidores de Google
```

## Estructura del proyecto

```
mail_manager/
├── main.go
├── handlers/
│   ├── auth.go
│   ├── gmail.go
│   └── email.go
├── services/
│   └── gmail_service.go
├── auth/
│   └── oauth_config.go
├── ui/
│   └── menu.go
└── token/
    └── token.json
```

## Paquetes a usar

- `net/http`
- `fmt`
- `bufio`
- `os`
- `encoding/json`
- `golang.org/x/oauth2`
- `golang.org/x/oauth2/google`
- `google.golang.org/api/gmail/v1`

## Reglas

- Comentarios en español explicando qué hace cada parte
- Cada archivo tiene una sola responsabilidad
- Nunca mezclar lógica de capas distintas en el mismo archivo
- `token/token.json` debe estar en `.gitignore`
- El código debe ser amigable para principiantes

## Equipo

- Kyle Reinoso
- Ariel Esparza
- Alejandro Zambrano

## Bitácora de progreso

### 2026-04-27 — Esqueleto inicial
- **Construido:** estructura de carpetas y archivos `.go` vacíos con declaración de paquete y comentario en español. `go mod init mail_manager` ejecutado. `.gitignore` creado para excluir `token/token.json`. `CLAUDE.md` creado con todo el contexto del proyecto.
- **Archivos creados:**
  - `CLAUDE.md`
  - `.gitignore`
  - `go.mod`
  - `main.go`
  - `handlers/auth.go`, `handlers/gmail.go`, `handlers/email.go`
  - `services/gmail_service.go`
  - `auth/oauth_config.go`
  - `ui/menu.go`

### 2026-04-27 — Documentación pública (README)
- **Construido:** archivo `README.md` en español con título, badges (Go, Gmail API, OAuth2), descripción del proyecto y autores, lista de las 7 funcionalidades, explicación de la arquitectura por capas, tecnologías y paquetes utilizados, pasos de instalación y uso, árbol de estructura comentado, estado del proyecto (Etapa 1 completada, Etapa 2 en progreso) y nota de seguridad sobre `credentials.json` y `token/token.json`.
- **Archivos modificados:** `README.md` (nuevo), `CLAUDE.md` (esta bitácora).
- **Siguiente paso:** implementar la configuración OAuth2 en `auth/oauth_config.go` y el flujo de autenticación inicial en `handlers/auth.go` para obtener y guardar el primer `token.json`. Luego construir el menú numerado en `ui/menu.go`.
