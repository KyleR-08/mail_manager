// Configuración del flujo OAuth2 contra Google (Gmail / Google Workspace).
// Versión orientada a un servidor HTTP: el usuario es redirigido al
// navegador para iniciar sesión y Google devuelve el código a nuestro
// endpoint /auth/google/callback. Aquí ya no se pegan códigos a mano.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// GoogleAuthConfig contiene las rutas a los archivos necesarios para
// autenticar contra Google: credentials.json (descargado de Google Cloud
// Console) y la ruta donde se guardará el token del usuario.
type GoogleAuthConfig struct {
	CredentialsFile string // Ruta a credentials.json descargado de Google Cloud
	TokenFile       string // Ruta donde guardar/leer el token OAuth2 del usuario
}

// googleRedirectURL es la URL a la que Google redirige al usuario después
// de iniciar sesión. Debe estar registrada en Google Cloud Console como
// "Authorized redirect URI" para que el flujo OAuth2 funcione.
const googleRedirectURL = "http://localhost:8080/auth/google/callback"

// buildGoogleOAuthConfig construye el oauth2.Config a partir del
// credentials.json. Se reutiliza tanto al generar la URL como al
// intercambiar el código por un token.
func buildGoogleOAuthConfig(cfg GoogleAuthConfig) (*oauth2.Config, error) {
	// Leer el archivo credentials.json descargado de Google Cloud Console
	credBytes, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer credentials.json: %v", err)
	}

	// Convertir el JSON de credenciales en un oauth2.Config con los scopes de Gmail
	oauthConfig, err := google.ConfigFromJSON(credBytes,
		gmail.GmailReadonlyScope, // Permiso para leer correos
		gmail.GmailSendScope,     // Permiso para enviar correos
	)
	if err != nil {
		return nil, fmt.Errorf("no se pudo parsear credentials.json: %v", err)
	}

	// Forzar la URL de redirección a la del servidor HTTP local
	oauthConfig.RedirectURL = googleRedirectURL
	return oauthConfig, nil
}

// GetGoogleAuthURL genera la URL de autorización a la que se debe
// redirigir al usuario para que inicie sesión con Google.
func GetGoogleAuthURL(cfg GoogleAuthConfig) (string, error) {
	// Construir el oauth2.Config con scopes de Gmail
	oauthConfig, err := buildGoogleOAuthConfig(cfg)
	if err != nil {
		return "", err
	}

	// Generar la URL de consentimiento con AccessType=offline para tener refresh token
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return authURL, nil
}

// ExchangeGoogleCode recibe el código que Google envía al callback y lo
// intercambia por un token de acceso real. Guarda el token en disco para
// poder reutilizarlo después sin volver a pedir login al usuario.
func ExchangeGoogleCode(cfg GoogleAuthConfig, code string) (*oauth2.Token, error) {
	// Construir el oauth2.Config con scopes de Gmail
	oauthConfig, err := buildGoogleOAuthConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Intercambiar el código por un token real contra Google
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el código por un token: %v", err)
	}

	// Guardar el token en disco para usarlo en futuras peticiones
	err = SaveToken(cfg.TokenFile, token)
	if err != nil {
		fmt.Println("No se pudo guardar el token de Google:", err)
	}

	// Devolver el token obtenido
	return token, nil
}

// GetGoogleClient devuelve un *http.Client autenticado contra Google,
// listo para llamar a Gmail API. Si NO hay token guardado devuelve (nil, nil)
// para que el servidor sepa que debe pedir login OAuth2 primero.
func GetGoogleClient(cfg GoogleAuthConfig) (*http.Client, error) {
	// Construir el oauth2.Config con scopes de Gmail
	oauthConfig, err := buildGoogleOAuthConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Intentar cargar un token guardado previamente
	token, err := LoadToken(cfg.TokenFile)
	if err != nil {
		// No hay token todavía — el servidor debe redirigir a /auth/google/start
		return nil, nil
	}

	// TokenSource refresca el token automáticamente cuando expira
	tokenSource := oauthConfig.TokenSource(context.Background(), token)

	// Forzar a obtener el token actual (lo refresca si está caducado)
	nuevoToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("no se pudo refrescar el token de Google: %v", err)
	}

	// Si el token fue refrescado, guardarlo en disco también
	if nuevoToken.AccessToken != token.AccessToken {
		err = SaveToken(cfg.TokenFile, nuevoToken)
		if err != nil {
			fmt.Println("No se pudo guardar el token refrescado:", err)
		}
	}

	// Devolver el cliente HTTP autenticado contra Google
	return oauthConfig.Client(context.Background(), nuevoToken), nil
}
