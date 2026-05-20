// Implementación de EmailProvider para cuentas de Gmail.
// Usa la Gmail API v1 directamente con net/http (sin SDK oficial),
// ya autenticada por el *http.Client de OAuth2 que recibe en el constructor.
// Traduce las operaciones genéricas (bandeja, lectura, envío, búsqueda)
// a llamadas REST contra https://gmail.googleapis.com/gmail/v1/users/me/...
package providers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GmailProvider implementa EmailProvider para cuentas de Gmail.
// El http.Client ya viene autenticado con OAuth2 desde auth/google_auth.go.
type GmailProvider struct {
	client *http.Client // Cliente HTTP autenticado contra Gmail API
}

// NewGmailProvider construye un proveedor de Gmail con el cliente OAuth2 ya listo.
func NewGmailProvider(client *http.Client) *GmailProvider {
	return &GmailProvider{client: client}
}

// gmailMessageList representa la respuesta del endpoint /messages
// que devuelve solo ids de mensajes (no su contenido completo).
type gmailMessageList struct {
	Messages []struct {
		Id string `json:"id"` // Id único del mensaje en Gmail
	} `json:"messages"`
}

// gmailMessage representa un mensaje completo devuelto por Gmail API.
type gmailMessage struct {
	Id      string `json:"id"`      // Id del mensaje
	Snippet string `json:"snippet"` // Vista previa corta del cuerpo
	Payload struct {
		Headers []struct {
			Name  string `json:"name"`  // Nombre del header (From, Subject, etc.)
			Value string `json:"value"` // Valor del header
		} `json:"headers"`
	} `json:"payload"`
}

// getHeader busca un header por nombre dentro de un mensaje de Gmail.
// Devuelve string vacío si no lo encuentra.
func getHeader(msg gmailMessage, nombre string) string {
	// Recorrer todos los headers buscando el que coincide
	for i := 0; i < len(msg.Payload.Headers); i++ {
		if msg.Payload.Headers[i].Name == nombre {
			return msg.Payload.Headers[i].Value
		}
	}
	return ""
}

// fetchMessageById descarga un mensaje completo de Gmail por su id.
func (g *GmailProvider) fetchMessageById(id string) (Mail, error) {
	// URL del endpoint que devuelve el mensaje en formato completo
	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages/" + id + "?format=full"

	// Llamar a la API con el cliente autenticado
	respuesta, err := g.client.Get(url)
	if err != nil {
		return Mail{}, err
	}
	defer respuesta.Body.Close()

	// Verificar que la respuesta fue exitosa
	if respuesta.StatusCode != 200 {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		return Mail{}, fmt.Errorf("error de Gmail API (%d): %s", respuesta.StatusCode, string(cuerpo))
	}

	// Decodificar la respuesta JSON
	var msg gmailMessage
	err = json.NewDecoder(respuesta.Body).Decode(&msg)
	if err != nil {
		return Mail{}, err
	}

	// Armar el struct Mail genérico extrayendo los headers
	correo := Mail{
		Id:      msg.Id,
		From:    getHeader(msg, "From"),
		To:      getHeader(msg, "To"),
		Subject: getHeader(msg, "Subject"),
		Body:    msg.Snippet,
		Date:    getHeader(msg, "Date"),
	}
	return correo, nil
}

// listMessages llama al endpoint /messages con la URL indicada y devuelve
// los mensajes encontrados como []Mail. Se usa para inbox, enviados y búsqueda.
func (g *GmailProvider) listMessages(url string) ([]Mail, error) {
	// Lista vacía por defecto
	var correos []Mail

	// Llamar al endpoint que solo devuelve ids
	respuesta, err := g.client.Get(url)
	if err != nil {
		return correos, err
	}
	defer respuesta.Body.Close()

	// Verificar éxito de la respuesta
	if respuesta.StatusCode != 200 {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		return correos, fmt.Errorf("error de Gmail API (%d): %s", respuesta.StatusCode, string(cuerpo))
	}

	// Parsear el JSON con la lista de ids
	var lista gmailMessageList
	err = json.NewDecoder(respuesta.Body).Decode(&lista)
	if err != nil {
		return correos, err
	}

	// Para cada id, descargar el mensaje completo y agregarlo a la lista
	for i := 0; i < len(lista.Messages); i++ {
		correo, err := g.fetchMessageById(lista.Messages[i].Id)
		if err != nil {
			// Si un mensaje falla, lo saltamos y seguimos con el siguiente
			fmt.Println("No se pudo cargar el mensaje:", err)
			continue
		}
		correos = append(correos, correo)
	}
	return correos, nil
}

// GetInbox devuelve los últimos 20 correos de la bandeja de entrada.
func (g *GmailProvider) GetInbox() ([]Mail, error) {
	// URL del endpoint con filtro de etiqueta INBOX
	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages?labelIds=INBOX&maxResults=20"
	return g.listMessages(url)
}

// ReadMail devuelve un correo específico identificado por su id.
func (g *GmailProvider) ReadMail(id string) (Mail, error) {
	return g.fetchMessageById(id)
}

// SendMail envía un correo nuevo desde la cuenta autenticada.
func (g *GmailProvider) SendMail(to string, subject string, body string) error {
	// Construir el mensaje en formato RFC 2822 (texto plano)
	mensaje := "To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body

	// Codificar el mensaje en base64 URL-safe (lo exige Gmail API)
	raw := base64.URLEncoding.EncodeToString([]byte(mensaje))

	// Armar el cuerpo JSON que espera el endpoint send
	cuerpoJSON := map[string]string{"raw": raw}
	cuerpoBytes, err := json.Marshal(cuerpoJSON)
	if err != nil {
		return err
	}

	// URL del endpoint para enviar correos
	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"

	// Hacer la petición POST con el cuerpo JSON
	respuesta, err := g.client.Post(url, "application/json", bytes.NewReader(cuerpoBytes))
	if err != nil {
		return err
	}
	defer respuesta.Body.Close()

	// Verificar que la respuesta fue exitosa (200 OK)
	if respuesta.StatusCode != 200 {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		return fmt.Errorf("error al enviar correo (%d): %s", respuesta.StatusCode, string(cuerpo))
	}
	return nil
}

// GetSent devuelve los últimos 20 correos enviados desde esta cuenta.
func (g *GmailProvider) GetSent() ([]Mail, error) {
	// URL del endpoint con filtro de etiqueta SENT
	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages?labelIds=SENT&maxResults=20"
	return g.listMessages(url)
}

// SearchMail busca correos que coincidan con la consulta indicada.
func (g *GmailProvider) SearchMail(query string) ([]Mail, error) {
	// Escapar espacios simples para el parámetro de búsqueda
	consulta := strings.ReplaceAll(query, " ", "+")
	url := "https://gmail.googleapis.com/gmail/v1/users/me/messages?q=" + consulta + "&maxResults=20"
	return g.listMessages(url)
}
