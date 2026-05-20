// Implementación de EmailProvider para cuentas de Outlook / Hotmail / Live.
// Se conecta a Microsoft Graph API mediante OAuth2 con los endpoints de
// Microsoft Identity Platform y traduce las operaciones genéricas a las
// llamadas REST correspondientes de Graph (/me/messages, /me/sendMail, etc).
package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OutlookProvider implementa EmailProvider para cuentas personales de Outlook.
// El http.Client viene autenticado desde auth/microsoft_auth.go.
type OutlookProvider struct {
	client *http.Client // Cliente HTTP autenticado contra Microsoft Graph
}

// NewOutlookProvider construye un proveedor de Outlook con el cliente OAuth2 listo.
func NewOutlookProvider(client *http.Client) *OutlookProvider {
	return &OutlookProvider{client: client}
}

// graphMessage representa un mensaje devuelto por Microsoft Graph API.
type graphMessage struct {
	Id      string `json:"id"`      // Id único del mensaje
	Subject string `json:"subject"` // Asunto del correo
	From    struct {
		EmailAddress struct {
			Address string `json:"address"` // Dirección del remitente
			Name    string `json:"name"`    // Nombre del remitente
		} `json:"emailAddress"`
	} `json:"from"`
	ToRecipients []struct {
		EmailAddress struct {
			Address string `json:"address"` // Dirección del destinatario
		} `json:"emailAddress"`
	} `json:"toRecipients"`
	Body struct {
		Content     string `json:"content"`     // Cuerpo del correo
		ContentType string `json:"contentType"` // text o html
	} `json:"body"`
	ReceivedDateTime string `json:"receivedDateTime"` // Fecha en ISO 8601
}

// graphMessageList representa la respuesta de un listado de mensajes.
type graphMessageList struct {
	Value []graphMessage `json:"value"` // Microsoft Graph envuelve listas en "value"
}

// graphToMail convierte un graphMessage en el struct Mail genérico.
func graphToMail(m graphMessage) Mail {
	// Concatenar destinatarios separados por coma
	var destinatarios string
	for i := 0; i < len(m.ToRecipients); i++ {
		if i > 0 {
			destinatarios = destinatarios + ", "
		}
		destinatarios = destinatarios + m.ToRecipients[i].EmailAddress.Address
	}

	// Armar el struct Mail con los campos extraídos
	correo := Mail{
		Id:      m.Id,
		From:    m.From.EmailAddress.Address,
		To:      destinatarios,
		Subject: m.Subject,
		Body:    m.Body.Content,
		Date:    m.ReceivedDateTime,
	}
	return correo
}

// doGraphGet hace una petición GET a Microsoft Graph y devuelve el cuerpo.
func (o *OutlookProvider) doGraphGet(url string) ([]byte, error) {
	// Hacer la llamada GET con el cliente autenticado
	respuesta, err := o.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer respuesta.Body.Close()

	// Leer todo el cuerpo de la respuesta
	cuerpo, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return nil, err
	}

	// Verificar éxito de la respuesta
	if respuesta.StatusCode != 200 {
		return nil, fmt.Errorf("error de Microsoft Graph (%d): %s", respuesta.StatusCode, string(cuerpo))
	}
	return cuerpo, nil
}

// listMessagesGraph hace GET a un endpoint que devuelve un listado de mensajes.
func (o *OutlookProvider) listMessagesGraph(url string) ([]Mail, error) {
	// Lista vacía por defecto
	var correos []Mail

	// Pedir el listado a Graph
	cuerpo, err := o.doGraphGet(url)
	if err != nil {
		return correos, err
	}

	// Decodificar la respuesta JSON con la lista de mensajes
	var lista graphMessageList
	err = json.Unmarshal(cuerpo, &lista)
	if err != nil {
		return correos, err
	}

	// Convertir cada mensaje de Graph al struct Mail
	for i := 0; i < len(lista.Value); i++ {
		correos = append(correos, graphToMail(lista.Value[i]))
	}
	return correos, nil
}

// GetInbox devuelve los últimos 20 correos de la bandeja de entrada.
func (o *OutlookProvider) GetInbox() ([]Mail, error) {
	// URL del endpoint inbox con campos seleccionados para reducir respuesta
	url := "https://graph.microsoft.com/v1.0/me/mailFolders/inbox/messages?$top=20&$select=id,subject,from,toRecipients,body,receivedDateTime"
	return o.listMessagesGraph(url)
}

// ReadMail descarga un correo específico por su id.
func (o *OutlookProvider) ReadMail(id string) (Mail, error) {
	// URL del endpoint que devuelve un mensaje por id
	url := "https://graph.microsoft.com/v1.0/me/messages/" + id

	// Pedir el mensaje a Graph
	cuerpo, err := o.doGraphGet(url)
	if err != nil {
		return Mail{}, err
	}

	// Decodificar el JSON en un graphMessage
	var msg graphMessage
	err = json.Unmarshal(cuerpo, &msg)
	if err != nil {
		return Mail{}, err
	}

	// Convertir y devolver
	return graphToMail(msg), nil
}

// SendMail envía un nuevo correo a través de Microsoft Graph.
func (o *OutlookProvider) SendMail(to string, subject string, body string) error {
	// Armar la estructura JSON que espera el endpoint sendMail de Graph
	cuerpoMensaje := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]string{
				"contentType": "Text",
				"content":     body,
			},
			"toRecipients": []map[string]interface{}{
				{
					"emailAddress": map[string]string{
						"address": to,
					},
				},
			},
		},
		"saveToSentItems": "true",
	}

	// Serializar el cuerpo a JSON
	jsonBytes, err := json.Marshal(cuerpoMensaje)
	if err != nil {
		return err
	}

	// URL del endpoint sendMail
	url := "https://graph.microsoft.com/v1.0/me/sendMail"

	// Hacer la petición POST con el JSON
	respuesta, err := o.client.Post(url, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	defer respuesta.Body.Close()

	// Graph devuelve 202 Accepted cuando acepta el envío
	if respuesta.StatusCode != 202 {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		return fmt.Errorf("error al enviar correo (%d): %s", respuesta.StatusCode, string(cuerpo))
	}
	return nil
}

// GetSent devuelve los últimos 20 correos enviados desde esta cuenta.
func (o *OutlookProvider) GetSent() ([]Mail, error) {
	// URL del endpoint de elementos enviados
	url := "https://graph.microsoft.com/v1.0/me/mailFolders/sentitems/messages?$top=20&$select=id,subject,from,toRecipients,body,receivedDateTime"
	return o.listMessagesGraph(url)
}

// SearchMail busca correos que coincidan con la consulta usando $search de Graph.
func (o *OutlookProvider) SearchMail(query string) ([]Mail, error) {
	// Reemplazar comillas dobles para evitar romper la consulta
	consulta := strings.ReplaceAll(query, "\"", "")
	url := "https://graph.microsoft.com/v1.0/me/messages?$search=\"" + consulta + "\"&$top=20"
	return o.listMessagesGraph(url)
}
