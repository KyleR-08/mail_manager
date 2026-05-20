// InstitutionalProvider maneja el correo institucional de la UIDE (@uide.edu.ec) usando Microsoft 365.
// Funcionalmente es idéntico a OutlookProvider (ambos usan Microsoft Graph API)
// pero existe como tipo separado por claridad semántica: representa una
// cuenta institucional, no una personal de Outlook/Hotmail. La diferencia
// real en producción está en el TenantID que usa el flujo OAuth2 (el de UIDE
// en lugar de "consumers"), no en la implementación del proveedor.
package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// InstitutionalProvider implementa EmailProvider para cuentas institucionales
// de la UIDE (@uide.edu.ec), que también corren sobre Microsoft 365.
type InstitutionalProvider struct {
	client *http.Client // Cliente HTTP autenticado contra Microsoft Graph (tenant UIDE)
}

// NewInstitutionalProvider construye un proveedor institucional con el cliente OAuth2 listo.
func NewInstitutionalProvider(client *http.Client) *InstitutionalProvider {
	return &InstitutionalProvider{client: client}
}

// doGraphGet hace una petición GET autenticada a Microsoft Graph.
func (p *InstitutionalProvider) doGraphGet(url string) ([]byte, error) {
	// Llamar al endpoint de Graph con el cliente autenticado
	respuesta, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer respuesta.Body.Close()

	// Leer el cuerpo de la respuesta
	cuerpo, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return nil, err
	}

	// Verificar respuesta exitosa
	if respuesta.StatusCode != 200 {
		return nil, fmt.Errorf("error de Microsoft Graph (%d): %s", respuesta.StatusCode, string(cuerpo))
	}
	return cuerpo, nil
}

// listMessagesGraph parsea un listado de mensajes desde un endpoint de Graph.
// Reutiliza los structs graphMessage / graphMessageList definidos en outlook.go
// (mismo paquete providers).
func (p *InstitutionalProvider) listMessagesGraph(url string) ([]Mail, error) {
	// Lista vacía por defecto
	var correos []Mail

	// Pedir el listado a Graph
	cuerpo, err := p.doGraphGet(url)
	if err != nil {
		return correos, err
	}

	// Decodificar la respuesta JSON
	var lista graphMessageList
	err = json.Unmarshal(cuerpo, &lista)
	if err != nil {
		return correos, err
	}

	// Convertir cada mensaje de Graph al struct Mail genérico
	for i := 0; i < len(lista.Value); i++ {
		correos = append(correos, graphToMail(lista.Value[i]))
	}
	return correos, nil
}

// GetInbox devuelve los últimos 20 correos de la bandeja de entrada institucional.
func (p *InstitutionalProvider) GetInbox() ([]Mail, error) {
	// URL del endpoint inbox de Microsoft Graph
	url := "https://graph.microsoft.com/v1.0/me/mailFolders/inbox/messages?$top=20&$select=id,subject,from,toRecipients,body,receivedDateTime"
	return p.listMessagesGraph(url)
}

// ReadMail descarga un correo institucional específico por su id.
func (p *InstitutionalProvider) ReadMail(id string) (Mail, error) {
	// URL del endpoint que devuelve un mensaje por id
	url := "https://graph.microsoft.com/v1.0/me/messages/" + id

	// Pedir el mensaje a Graph
	cuerpo, err := p.doGraphGet(url)
	if err != nil {
		return Mail{}, err
	}

	// Decodificar el JSON
	var msg graphMessage
	err = json.Unmarshal(cuerpo, &msg)
	if err != nil {
		return Mail{}, err
	}

	// Convertir al struct Mail genérico
	return graphToMail(msg), nil
}

// SendMail envía un correo desde la cuenta institucional vía Microsoft Graph.
func (p *InstitutionalProvider) SendMail(to string, subject string, body string) error {
	// Construir el JSON que pide el endpoint sendMail
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

	// Hacer la petición POST
	respuesta, err := p.client.Post(url, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	defer respuesta.Body.Close()

	// Graph responde 202 Accepted cuando acepta el envío
	if respuesta.StatusCode != 202 {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		return fmt.Errorf("error al enviar correo (%d): %s", respuesta.StatusCode, string(cuerpo))
	}
	return nil
}

// GetSent devuelve los últimos 20 correos enviados desde la cuenta institucional.
func (p *InstitutionalProvider) GetSent() ([]Mail, error) {
	// URL del endpoint de elementos enviados
	url := "https://graph.microsoft.com/v1.0/me/mailFolders/sentitems/messages?$top=20&$select=id,subject,from,toRecipients,body,receivedDateTime"
	return p.listMessagesGraph(url)
}

// SearchMail busca correos institucionales que coincidan con la consulta.
func (p *InstitutionalProvider) SearchMail(query string) ([]Mail, error) {
	// Quitar comillas para que no rompan el parámetro $search
	consulta := strings.ReplaceAll(query, "\"", "")
	url := "https://graph.microsoft.com/v1.0/me/messages?$search=\"" + consulta + "\"&$top=20"
	return p.listMessagesGraph(url)
}
