// Configuración del flujo OAuth2 contra Google (Gmail / Google Workspace).
// Define los scopes de Gmail, carga el credentials.json de la app de Google
// y construye el oauth2.Config necesario para autenticar al usuario, abrir
// el flujo de consentimiento y guardar el token resultante en disco.
package auth

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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

// GetGoogleClient devuelve un *http.Client autenticado contra Google,
// listo para llamar a Gmail API. Si hay token guardado lo usa, si no
// abre el flujo OAuth2 en el navegador.
// Nota: LoadToken y SaveToken se reutilizan de microsoft_auth.go (mismo
// paquete auth), porque la lógica de leer/escribir un oauth2.Token desde
// JSON es idéntica para ambos proveedores.
func GetGoogleClient(cfg GoogleAuthConfig) (*http.Client, error) {
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

	// Intentar cargar un token guardado previamente
	token, err := LoadToken(cfg.TokenFile)
	if err != nil {
		// Si no hay token guardado, abrir el flujo OAuth2 desde el navegador
		fmt.Println("No se encontró token guardado, iniciando flujo OAuth2 con Google...")
		token, err = GetGoogleTokenFromWeb(oauthConfig)
		if err != nil {
			return nil, fmt.Errorf("no se pudo obtener el token desde el navegador: %v", err)
		}

		// Guardar el token recién obtenido para no pedirlo otra vez
		err = SaveToken(cfg.TokenFile, token)
		if err != nil {
			fmt.Println("No se pudo guardar el token de Google:", err)
		}
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

// GetGoogleTokenFromWeb realiza el flujo OAuth2 interactivo: imprime la
// URL en consola, espera a que el usuario inicie sesión, copie SOLO el
// valor del parámetro "code" de la URL de redirección y lo pegue acá;
// luego intercambia ese código por un token de acceso real.
func GetGoogleTokenFromWeb(oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	// Generar la URL de autorización con AccessType=offline para tener refresh token
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("")
	fmt.Println("Abre la siguiente URL en tu navegador e inicia sesión con Google:")
	fmt.Println(authURL)
	fmt.Println("")
	fmt.Println("Después del login serás redirigido a una URL tipo:")
	fmt.Println("  http://localhost/?state=state-token&code=4/0ABC123&scope=...")
	fmt.Println("")
	fmt.Print("Pega aquí SOLO el código de autorización (el valor después de 'code=' en la URL): ")

	// Leer toda la línea pegada por el usuario para soportar códigos largos con
	// caracteres especiales (las URL OAuth2 traen "/", "-", "_", etc.).
	lector := bufio.NewReader(os.Stdin)
	codigo, err := lector.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el código: %v", err)
	}

	// Quitar espacios y saltos de línea al inicio y al final
	codigo = strings.TrimSpace(codigo)

	// Validar que no esté vacío después de limpiar
	if codigo == "" {
		return nil, fmt.Errorf("el código de autorización está vacío")
	}

	// Intercambiar el código por un token real
	token, err := oauthConfig.Exchange(context.Background(), codigo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el código por un token: %v", err)
	}

	// Devolver el token obtenido
	return token, nil
}
