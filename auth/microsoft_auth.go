// Configuración del flujo OAuth2 contra Microsoft Identity Platform
// (Outlook / Hotmail / Live / Microsoft 365 institucional). Versión orientada
// a un servidor HTTP: el usuario es redirigido al navegador y Microsoft
// devuelve el código a nuestro endpoint /auth/microsoft/callback.
// Define los scopes de Microsoft Graph (Mail.Read, Mail.Send, offline_access)
// y guarda el token en disco.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// MicrosoftAuthConfig contiene los datos necesarios para autenticar contra
// Microsoft Identity Platform. ClientID y ClientSecret se obtienen al
// registrar la app en Azure AD. TenantID es "consumers" para cuentas
// personales (Outlook/Hotmail) o el id real del tenant para institucionales.
type MicrosoftAuthConfig struct {
	ClientID     string // Id de la aplicación registrada en Azure AD
	ClientSecret string // Secreto de cliente generado en Azure AD
	TenantID     string // "consumers" o el tenant id institucional (ej. UIDE)
	TokenFile    string // Ruta donde guardar/leer el token OAuth2 del usuario
}

// microsoftRedirectURL es la URL a la que Microsoft redirige al usuario
// después de iniciar sesión. Debe estar registrada en Azure AD como
// "Redirect URI" para que el flujo OAuth2 funcione.
const microsoftRedirectURL = "http://localhost:8080/auth/microsoft/callback"

// buildMicrosoftOAuthConfig construye el oauth2.Config para Microsoft
// con los endpoints específicos del tenant indicado.
func buildMicrosoftOAuthConfig(cfg MicrosoftAuthConfig) *oauth2.Config {
	// Armar la configuración OAuth2 con scopes de Microsoft Graph
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		// Scopes necesarios para leer y enviar correo + refresh token
		Scopes: []string{
			"https://graph.microsoft.com/Mail.Read",
			"https://graph.microsoft.com/Mail.Send",
			"offline_access",
		},
		// URL del servidor HTTP local para recibir el código de autorización
		RedirectURL: microsoftRedirectURL,
		// Endpoints específicos del tenant de Microsoft
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/token",
		},
	}
}

// GetMicrosoftAuthURL genera la URL de autorización a la que se debe
// redirigir al usuario para que inicie sesión con Microsoft.
func GetMicrosoftAuthURL(cfg MicrosoftAuthConfig) (string, error) {
	// Construir el oauth2.Config con los endpoints del tenant
	oauthConfig := buildMicrosoftOAuthConfig(cfg)

	// Generar la URL de consentimiento con AccessType=offline para refresh token
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return authURL, nil
}

// ExchangeMicrosoftCode recibe el código que Microsoft envía al callback
// y lo intercambia por un token de acceso real. Guarda el token en disco
// para reutilizarlo después.
func ExchangeMicrosoftCode(cfg MicrosoftAuthConfig, code string) (*oauth2.Token, error) {
	// Construir el oauth2.Config con los endpoints del tenant
	oauthConfig := buildMicrosoftOAuthConfig(cfg)

	// Intercambiar el código por un token real contra Microsoft
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el código por un token: %v", err)
	}

	// Guardar el token en disco para usarlo en futuras peticiones
	err = SaveToken(cfg.TokenFile, token)
	if err != nil {
		fmt.Println("No se pudo guardar el token de Microsoft:", err)
	}

	// Devolver el token obtenido
	return token, nil
}

// GetMicrosoftClient devuelve un *http.Client listo para llamar a Microsoft
// Graph. Si NO hay token guardado devuelve (nil, nil) para que el servidor
// sepa que debe pedir login OAuth2 primero.
func GetMicrosoftClient(cfg MicrosoftAuthConfig) (*http.Client, error) {
	// Construir el oauth2.Config con los endpoints del tenant
	oauthConfig := buildMicrosoftOAuthConfig(cfg)

	// Intentar cargar un token previamente guardado en disco
	token, err := LoadToken(cfg.TokenFile)
	if err != nil {
		// No hay token todavía — el servidor debe redirigir a /auth/microsoft/start
		return nil, nil
	}

	// El TokenSource refresca el token automáticamente cuando expira
	tokenSource := oauthConfig.TokenSource(context.Background(), token)

	// Pedir el token actualizado (si caducó, esto lo refresca)
	nuevoToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("no se pudo refrescar el token de Microsoft: %v", err)
	}

	// Si el token fue refrescado, guardarlo de nuevo en disco
	if nuevoToken.AccessToken != token.AccessToken {
		err = SaveToken(cfg.TokenFile, nuevoToken)
		if err != nil {
			fmt.Println("No se pudo guardar el token refrescado:", err)
		}
	}

	// Devolver un cliente HTTP que ya incluye el token en cada petición
	return oauthConfig.Client(context.Background(), nuevoToken), nil
}

// LoadToken lee un token OAuth2 previamente guardado en un archivo JSON.
// Devuelve un error si el archivo no existe o el JSON está mal formado.
func LoadToken(filename string) (*oauth2.Token, error) {
	// Abrir el archivo de token en modo lectura
	archivo, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer archivo.Close()

	// Decodificar el contenido JSON dentro de un oauth2.Token
	token := &oauth2.Token{}
	err = json.NewDecoder(archivo).Decode(token)
	if err != nil {
		return nil, err
	}

	// Devolver el token cargado
	return token, nil
}

// SaveToken guarda un token OAuth2 en disco como JSON, con permisos
// restringidos (0600) porque contiene credenciales sensibles.
func SaveToken(filename string, token *oauth2.Token) error {
	// Asegurar que la carpeta del token exista
	dir := filepathDir(filename)
	if dir != "" && dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	// Crear o sobrescribir el archivo donde se guarda el token
	archivo, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer archivo.Close()

	// Codificar el token como JSON dentro del archivo
	err = json.NewEncoder(archivo).Encode(token)
	if err != nil {
		return err
	}

	// Todo salió bien
	return nil
}

// filepathDir devuelve la carpeta padre de una ruta. Implementación
// simple para evitar importar "path/filepath" solo para esto.
func filepathDir(p string) string {
	// Buscar el último separador de directorio
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return ""
}
