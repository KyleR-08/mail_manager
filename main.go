// Punto de entrada de la aplicación correo-manager (versión servidor HTTP).
// Arranca un servidor Gin en el puerto 8080 con endpoints REST para
// administrar las cuentas, autenticar con OAuth2 (Google y Microsoft) y
// operar el correo (bandeja, enviar, enviados, buscar).
//
// CONCEPTO CLAVE — POLIMORFISMO:
// La variable global "proveedorActivo" se declara con el tipo de la interfaz
// providers.EmailProvider, pero adentro puede vivir un *GmailProvider,
// un *OutlookProvider o un *InstitutionalProvider. Los handlers HTTP
// llaman a proveedorActivo.GetInbox() / SendMail() sin saber cuál es.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"mail_manager/accounts"
	"mail_manager/auth"
	"mail_manager/providers"
)

// Constantes de credenciales de la app para cada proveedor.
// Solo Gmail tiene credenciales reales (descargadas de Google Cloud Console).
// Las credenciales de Microsoft (Outlook e institucional UIDE) son
// PLACEHOLDERS porque registrar una app en Azure AD requiere permisos de
// administrador del tenant. Cuando se registren las apps reales, basta
// con reemplazar el valor de estas constantes.
const gmailCredentialsFile = "credentials.json"
const outlookClientID = "TU_CLIENT_ID_AQUI"
const outlookClientSecret = "TU_CLIENT_SECRET_AQUI"
const outlookTenantID = "consumers"
const uideClientID = "TU_CLIENT_ID_AQUI"
const uideClientSecret = "TU_CLIENT_SECRET_AQUI"
const uideTenantID = "TU_TENANT_ID_UIDE_AQUI"

// Estado global en memoria del servidor.
// listaCuentas contiene las cuentas registradas, cuentaActiva es la que se
// está usando ahora mismo y proveedorActivo es el EmailProvider concreto
// detrás de la interfaz (polimorfismo).
var (
	mutexEstado     sync.Mutex                // Protege el estado global de accesos concurrentes
	listaCuentas    []accounts.Account        // Cuentas registradas, cargadas desde disco
	cuentaActiva    accounts.Account          // Cuenta seleccionada actualmente
	proveedorActivo providers.EmailProvider   // Proveedor concreto detrás de la interfaz
)

// resolveCredentialsPath devuelve la ruta absoluta a credentials.json a
// partir del directorio de trabajo actual (donde se lanzó la app).
func resolveCredentialsPath() string {
	// Obtener el directorio de trabajo actual
	dir, err := os.Getwd()
	if err != nil {
		// Si falla, caer al nombre relativo como respaldo
		return gmailCredentialsFile
	}
	// Unir el directorio actual con el nombre del archivo de credenciales
	return filepath.Join(dir, gmailCredentialsFile)
}

// detectProvider devuelve el tipo de proveedor según el dominio del email.
// Reglas simples: @gmail.com -> gmail, @hotmail/@outlook -> outlook,
// @uide.edu.ec -> institutional.
func detectProvider(email string) string {
	// Pasar a minúsculas para comparar sin importar mayúsculas
	correo := strings.ToLower(email)

	// Detectar Gmail por su dominio
	if strings.HasSuffix(correo, "@gmail.com") {
		return "gmail"
	}

	// Detectar Outlook/Hotmail/Live por su dominio
	if strings.HasSuffix(correo, "@hotmail.com") || strings.HasSuffix(correo, "@outlook.com") || strings.HasSuffix(correo, "@live.com") {
		return "outlook"
	}

	// Detectar correo institucional UIDE
	if strings.HasSuffix(correo, "@uide.edu.ec") {
		return "institutional"
	}

	// Dominio desconocido
	return ""
}

// googleConfigPara devuelve la configuración OAuth2 de Google para una cuenta.
func googleConfigPara(cuenta accounts.Account) auth.GoogleAuthConfig {
	return auth.GoogleAuthConfig{
		CredentialsFile: resolveCredentialsPath(),
		TokenFile:       cuenta.TokenFile,
	}
}

// microsoftConfigPara devuelve la configuración OAuth2 de Microsoft para una cuenta.
// Decide los valores según si la cuenta es outlook o institutional.
func microsoftConfigPara(cuenta accounts.Account) auth.MicrosoftAuthConfig {
	// Caso Outlook (personal): tenant "consumers"
	if cuenta.ProviderType == "outlook" {
		return auth.MicrosoftAuthConfig{
			ClientID:     outlookClientID,
			ClientSecret: outlookClientSecret,
			TenantID:     outlookTenantID,
			TokenFile:    cuenta.TokenFile,
		}
	}

	// Caso Institucional UIDE: tenant propio de la UIDE
	return auth.MicrosoftAuthConfig{
		ClientID:     uideClientID,
		ClientSecret: uideClientSecret,
		TenantID:     uideTenantID,
		TokenFile:    cuenta.TokenFile,
	}
}

// microsoftCredencialesValidas verifica que las constantes de Microsoft
// no sean placeholders. Devuelve un error claro si lo son.
func microsoftCredencialesValidas(cfg auth.MicrosoftAuthConfig) error {
	// Detectar placeholders en cualquiera de los tres valores
	if cfg.ClientID == "" || cfg.ClientSecret == "" ||
		strings.HasPrefix(cfg.ClientID, "TU_") || strings.HasPrefix(cfg.ClientSecret, "TU_") {
		return fmt.Errorf("las credenciales de Microsoft son placeholders: registra la app en Azure AD y reemplaza las constantes en main.go")
	}

	// Si el tenant aún es el placeholder, también es inválido
	if strings.HasPrefix(cfg.TenantID, "TU_") {
		return fmt.Errorf("el TenantID de Microsoft es un placeholder: reemplaza la constante en main.go")
	}
	return nil
}

// buildProvider construye el EmailProvider concreto según el tipo de cuenta.
// AQUÍ ES DONDE OCURRE EL POLIMORFISMO: la función devuelve un EmailProvider
// (interfaz), pero adentro pone un proveedor distinto según el caso.
// Si la cuenta aún no tiene token OAuth2 devuelve (nil, nil) para que
// el llamador sepa que hay que pedir login primero.
func buildProvider(cuenta accounts.Account) (providers.EmailProvider, error) {
	// Caso Gmail: usar Google OAuth2 + GmailProvider
	if cuenta.ProviderType == "gmail" {
		cfg := googleConfigPara(cuenta)

		// Obtener el cliente HTTP autenticado (nil si no hay token aún)
		client, err := auth.GetGoogleClient(cfg)
		if err != nil {
			return nil, err
		}
		if client == nil {
			// Aún no se ha completado el flujo OAuth2 — proveedor vacío
			return nil, nil
		}

		// Devolver el proveedor Gmail (asignable a EmailProvider)
		return providers.NewGmailProvider(client), nil
	}

	// Caso Outlook o Institucional: ambos usan Microsoft Graph
	if cuenta.ProviderType == "outlook" || cuenta.ProviderType == "institutional" {
		cfg := microsoftConfigPara(cuenta)

		// Validar que las credenciales no sean placeholders
		err := microsoftCredencialesValidas(cfg)
		if err != nil {
			return nil, err
		}

		// Obtener el cliente HTTP autenticado (nil si no hay token aún)
		client, err := auth.GetMicrosoftClient(cfg)
		if err != nil {
			return nil, err
		}
		if client == nil {
			// Aún no se ha completado el flujo OAuth2 — proveedor vacío
			return nil, nil
		}

		// Devolver el proveedor concreto según el tipo
		if cuenta.ProviderType == "outlook" {
			return providers.NewOutlookProvider(client), nil
		}
		return providers.NewInstitutionalProvider(client), nil
	}

	// Tipo desconocido
	return nil, fmt.Errorf("tipo de proveedor desconocido: %s", cuenta.ProviderType)
}

// buscarCuentaPorEmail devuelve la cuenta con el email dado, o false si no existe.
func buscarCuentaPorEmail(email string) (accounts.Account, bool) {
	// Recorrer la lista de cuentas en memoria
	for i := 0; i < len(listaCuentas); i++ {
		if strings.EqualFold(listaCuentas[i].Email, email) {
			return listaCuentas[i], true
		}
	}
	return accounts.Account{}, false
}

// activarCuenta establece la cuenta activa y construye su proveedor.
// Devuelve error si la cuenta no se puede activar (por ejemplo, credenciales inválidas).
func activarCuenta(cuenta accounts.Account) error {
	// Intentar construir el proveedor (puede devolver nil si falta token)
	prov, err := buildProvider(cuenta)
	if err != nil {
		return err
	}

	// Actualizar el estado en memoria
	cuentaActiva = cuenta
	proveedorActivo = prov
	return nil
}

// --- Handlers HTTP ---

// healthHandler devuelve un OK para chequeos del servidor.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// listAccountsHandler devuelve la lista de cuentas registradas.
func listAccountsHandler(c *gin.Context) {
	mutexEstado.Lock()
	defer mutexEstado.Unlock()

	// Responder con la cuenta activa para que la UI sepa cuál es
	c.JSON(http.StatusOK, gin.H{
		"accounts": listaCuentas,
		"active":   cuentaActiva.Email,
	})
}

// createAccountRequest es el cuerpo JSON esperado por POST /accounts.
type createAccountRequest struct {
	Email       string `json:"email"`        // Email a registrar
	DisplayName string `json:"display_name"` // Nombre amigable
}

// addAccountHandler agrega una cuenta nueva a partir del JSON recibido.
func addAccountHandler(c *gin.Context) {
	// Parsear el cuerpo de la petición
	var req createAccountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	// Validar el email
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el email es obligatorio"})
		return
	}

	// Detectar el tipo de proveedor por el dominio
	tipo := detectProvider(req.Email)
	if tipo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dominio no soportado (solo @gmail.com, @hotmail.com, @outlook.com, @live.com o @uide.edu.ec)",
		})
		return
	}

	mutexEstado.Lock()
	defer mutexEstado.Unlock()

	// Verificar que no exista ya
	_, existe := buscarCuentaPorEmail(req.Email)
	if existe {
		c.JSON(http.StatusConflict, gin.H{"error": "la cuenta ya está registrada"})
		return
	}

	// Construir la cuenta nueva con la ruta del token derivada del email
	rutaToken := "token/" + strings.ReplaceAll(req.Email, "@", "_at_") + ".json"
	cuentaNueva := accounts.Account{
		Email:        req.Email,
		ProviderType: tipo,
		DisplayName:  req.DisplayName,
		TokenFile:    rutaToken,
	}

	// Agregar la cuenta a la lista y guardarla en disco
	listaCuentas = append(listaCuentas, cuentaNueva)
	accounts.SaveAccounts(listaCuentas)

	// Responder con la cuenta creada
	c.JSON(http.StatusCreated, cuentaNueva)
}

// activateAccountHandler establece la cuenta activa por su email.
func activateAccountHandler(c *gin.Context) {
	// Leer el email del parámetro de la URL
	email := c.Param("email")

	mutexEstado.Lock()
	defer mutexEstado.Unlock()

	// Buscar la cuenta en la lista
	cuenta, existe := buscarCuentaPorEmail(email)
	if !existe {
		c.JSON(http.StatusNotFound, gin.H{"error": "cuenta no encontrada"})
		return
	}

	// Intentar activarla (construye el proveedor)
	err := activarCuenta(cuenta)
	if err != nil {
		// Aun así actualizamos la cuenta activa para que el frontend pueda
		// disparar el flujo OAuth2 luego.
		cuentaActiva = cuenta
		proveedorActivo = nil
		c.JSON(http.StatusOK, gin.H{
			"active":  cuenta.Email,
			"warning": err.Error(),
		})
		return
	}

	// Responder con la cuenta activa y si ya tiene proveedor listo
	c.JSON(http.StatusOK, gin.H{
		"active":          cuenta.Email,
		"provider_ready":  proveedorActivo != nil,
	})
}

// googleStartHandler redirige al usuario a la URL de login de Google.
// Usa la cuenta activa para saber dónde guardar el token.
func googleStartHandler(c *gin.Context) {
	mutexEstado.Lock()
	cuenta := cuentaActiva
	mutexEstado.Unlock()

	// Validar que haya cuenta activa de tipo gmail
	if cuenta.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay cuenta activa"})
		return
	}
	if cuenta.ProviderType != "gmail" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "la cuenta activa no es de Gmail"})
		return
	}

	// Construir la URL de autorización
	url, err := auth.GetGoogleAuthURL(googleConfigPara(cuenta))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirigir al navegador del usuario al consentimiento de Google
	c.Redirect(http.StatusFound, url)
}

// googleCallbackHandler recibe el código de Google y lo intercambia por un token.
func googleCallbackHandler(c *gin.Context) {
	// Leer el código de la query string
	codigo := c.Query("code")
	if codigo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se recibió código de Google"})
		return
	}

	mutexEstado.Lock()
	cuenta := cuentaActiva
	mutexEstado.Unlock()

	// Validar que haya cuenta activa de tipo gmail
	if cuenta.Email == "" || cuenta.ProviderType != "gmail" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay cuenta activa de Gmail para el callback"})
		return
	}

	// Intercambiar el código por un token (se guarda en disco automáticamente)
	_, err := auth.ExchangeGoogleCode(googleConfigPara(cuenta), codigo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reconstruir el proveedor para que ya quede listo en memoria
	mutexEstado.Lock()
	err = activarCuenta(cuenta)
	mutexEstado.Unlock()
	if err != nil {
		fmt.Println("Aviso: token guardado pero no se pudo activar el proveedor:", err)
	}

	// Volver a la página principal
	c.Redirect(http.StatusFound, "/")
}

// microsoftStartHandler redirige al usuario a la URL de login de Microsoft.
func microsoftStartHandler(c *gin.Context) {
	mutexEstado.Lock()
	cuenta := cuentaActiva
	mutexEstado.Unlock()

	// Validar que haya cuenta activa de tipo Microsoft
	if cuenta.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay cuenta activa"})
		return
	}
	if cuenta.ProviderType != "outlook" && cuenta.ProviderType != "institutional" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "la cuenta activa no es de Microsoft"})
		return
	}

	// Validar credenciales antes de generar la URL
	cfg := microsoftConfigPara(cuenta)
	err := microsoftCredencialesValidas(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Construir la URL de autorización
	url, err := auth.GetMicrosoftAuthURL(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirigir al navegador del usuario al consentimiento de Microsoft
	c.Redirect(http.StatusFound, url)
}

// microsoftCallbackHandler recibe el código de Microsoft y lo intercambia por un token.
func microsoftCallbackHandler(c *gin.Context) {
	// Leer el código de la query string
	codigo := c.Query("code")
	if codigo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se recibió código de Microsoft"})
		return
	}

	mutexEstado.Lock()
	cuenta := cuentaActiva
	mutexEstado.Unlock()

	// Validar que la cuenta activa sea de Microsoft
	if cuenta.Email == "" || (cuenta.ProviderType != "outlook" && cuenta.ProviderType != "institutional") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay cuenta activa de Microsoft para el callback"})
		return
	}

	// Intercambiar el código por un token (se guarda en disco automáticamente)
	cfg := microsoftConfigPara(cuenta)
	_, err := auth.ExchangeMicrosoftCode(cfg, codigo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reconstruir el proveedor para dejarlo listo en memoria
	mutexEstado.Lock()
	err = activarCuenta(cuenta)
	mutexEstado.Unlock()
	if err != nil {
		fmt.Println("Aviso: token guardado pero no se pudo activar el proveedor:", err)
	}

	// Volver a la página principal
	c.Redirect(http.StatusFound, "/")
}

// requireProveedor obtiene el proveedor activo o responde 400 si no existe.
func requireProveedor(c *gin.Context) (providers.EmailProvider, bool) {
	mutexEstado.Lock()
	prov := proveedorActivo
	mutexEstado.Unlock()

	// Si no hay proveedor, avisarle al cliente
	if prov == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay proveedor activo (selecciona una cuenta y completa OAuth2)"})
		return nil, false
	}
	return prov, true
}

// inboxHandler devuelve la bandeja de entrada como JSON.
func inboxHandler(c *gin.Context) {
	// Asegurar que haya proveedor activo
	prov, ok := requireProveedor(c)
	if !ok {
		return
	}

	// Llamar al método polimórfico GetInbox()
	correos, err := prov.GetInbox()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, correos)
}

// sendRequest es el cuerpo JSON esperado por POST /send.
type sendRequest struct {
	To      string `json:"to"`      // Destinatario
	Subject string `json:"subject"` // Asunto
	Body    string `json:"body"`    // Cuerpo del correo
}

// sendHandler envía un correo usando el proveedor activo.
func sendHandler(c *gin.Context) {
	// Asegurar que haya proveedor activo
	prov, ok := requireProveedor(c)
	if !ok {
		return
	}

	// Parsear el cuerpo de la petición
	var req sendRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	// Validar campos mínimos
	if req.To == "" || req.Subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to y subject son obligatorios"})
		return
	}

	// Enviar usando el método polimórfico SendMail()
	err = prov.SendMail(req.To, req.Subject, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "enviado"})
}

// sentHandler devuelve los correos enviados como JSON.
func sentHandler(c *gin.Context) {
	// Asegurar que haya proveedor activo
	prov, ok := requireProveedor(c)
	if !ok {
		return
	}

	// Llamar al método polimórfico GetSent()
	correos, err := prov.GetSent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, correos)
}

// searchHandler busca correos que coincidan con el parámetro q.
func searchHandler(c *gin.Context) {
	// Asegurar que haya proveedor activo
	prov, ok := requireProveedor(c)
	if !ok {
		return
	}

	// Leer el parámetro de búsqueda de la query string
	consulta := c.Query("q")
	if consulta == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parámetro 'q' obligatorio"})
		return
	}

	// Llamar al método polimórfico SearchMail()
	correos, err := prov.SearchMail(consulta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, correos)
}

// main arranca el servidor HTTP.
func main() {
	// Mensaje de bienvenida
	fmt.Println("Iniciando correo-manager (servidor HTTP)...")

	// Cargar las cuentas registradas desde disco
	listaCuentas = accounts.LoadAccounts()
	fmt.Println("Cuentas cargadas:", len(listaCuentas))

	// Si hay una sola cuenta, activarla automáticamente (mejor UX)
	if len(listaCuentas) == 1 {
		err := activarCuenta(listaCuentas[0])
		if err != nil {
			fmt.Println("Aviso: no se pudo activar la cuenta única:", err)
			cuentaActiva = listaCuentas[0]
		}
	}

	// Crear el router Gin con middleware por defecto (logger + recovery)
	router := gin.Default()

	// Servir el archivo index.html y archivos estáticos del frontend
	router.StaticFile("/", "./static/index.html")
	router.Static("/static", "./static")

	// Endpoint de salud para chequeos
	router.GET("/health", healthHandler)

	// Endpoints de cuentas
	router.GET("/accounts", listAccountsHandler)
	router.POST("/accounts", addAccountHandler)
	router.POST("/accounts/:email/activate", activateAccountHandler)

	// Endpoints de autenticación OAuth2 (Google y Microsoft)
	router.GET("/auth/google/start", googleStartHandler)
	router.GET("/auth/google/callback", googleCallbackHandler)
	router.GET("/auth/microsoft/start", microsoftStartHandler)
	router.GET("/auth/microsoft/callback", microsoftCallbackHandler)

	// Endpoints de correo (usan el proveedor activo polimórficamente)
	router.GET("/inbox", inboxHandler)
	router.POST("/send", sendHandler)
	router.GET("/sent", sentHandler)
	router.GET("/search", searchHandler)

	// Arrancar el servidor en el puerto 8080
	fmt.Println("Servidor escuchando en http://localhost:8080")
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error al arrancar el servidor:", err)
	}
}
