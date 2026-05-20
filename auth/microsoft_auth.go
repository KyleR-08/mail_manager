// Configuración del flujo OAuth2 contra Microsoft Identity Platform
// (Outlook / Hotmail / Live / Microsoft 365 institucional). Define los scopes
// de Microsoft Graph (Mail.Read, Mail.Send, offline_access), construye el
// oauth2.Config con los endpoints de Microsoft y guarda el token en disco.
// Sirve tanto para cuentas personales de Outlook como para cuentas
// institucionales de la UIDE.
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

// GetMicrosoftClient devuelve un *http.Client listo para llamar a Microsoft
// Graph. Si ya existe un token en disco lo usa; si no, abre el flujo OAuth2
// en el navegador para que el usuario inicie sesión y otorgue permisos.
func GetMicrosoftClient(cfg MicrosoftAuthConfig) (*http.Client, error) {
	// Construir el oauth2.Config con los endpoints de Microsoft
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		// Scopes necesarios para leer y enviar correo + refresh token
		Scopes: []string{
			"https://graph.microsoft.com/Mail.Read",
			"https://graph.microsoft.com/Mail.Send",
			"offline_access",
		},
		// URL local para recibir el código de autorización
		RedirectURL: "http://localhost:8080/callback",
		// Endpoints específicos del tenant de Microsoft
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/token",
		},
	}

	// Intentar cargar un token previamente guardado en disco
	token, err := LoadToken(cfg.TokenFile)
	if err != nil {
		// Si no hay token guardado, iniciar el flujo OAuth2 desde el navegador
		fmt.Println("No se encontró token guardado, iniciando flujo OAuth2 con Microsoft...")
		token, err = GetMicrosoftTokenFromWeb(cfg)
		if err != nil {
			return nil, fmt.Errorf("no se pudo obtener el token desde el navegador: %v", err)
		}

		// Guardar el token recién obtenido para no pedirlo otra vez
		err = SaveToken(cfg.TokenFile, token)
		if err != nil {
			fmt.Println("No se pudo guardar el token de Microsoft:", err)
		}
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

// GetMicrosoftTokenFromWeb realiza el flujo OAuth2 interactivo: muestra
// la URL al usuario, espera a que pegue el código que recibe Microsoft y
// lo intercambia por un token de acceso.
func GetMicrosoftTokenFromWeb(cfg MicrosoftAuthConfig) (*oauth2.Token, error) {
	// Reconstruir el mismo oauthConfig para usarlo aquí también
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes: []string{
			"https://graph.microsoft.com/Mail.Read",
			"https://graph.microsoft.com/Mail.Send",
			"offline_access",
		},
		RedirectURL: "http://localhost:8080/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + cfg.TenantID + "/oauth2/v2.0/token",
		},
	}

	// Generar la URL que el usuario debe abrir en el navegador
	authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("")
	fmt.Println("Abre la siguiente URL en tu navegador e inicia sesión con Microsoft:")
	fmt.Println(authURL)
	fmt.Print("Pega aquí el código que aparece al terminar: ")

	// Leer el código pegado por el usuario en la consola
	var codigo string
	_, err := fmt.Scan(&codigo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el código: %v", err)
	}

	// Intercambiar el código por un token de acceso real
	token, err := oauthConfig.Exchange(context.Background(), codigo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo intercambiar el código por un token: %v", err)
	}

	// Devolver el token obtenido
	return token, nil
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
