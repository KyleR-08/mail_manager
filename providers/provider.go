// Define la interfaz EmailProvider que abstrae cualquier proveedor de correo.
// Es el contrato común que implementan Gmail, Outlook, Yahoo e Institucional.
// Gracias a esta interfaz se aplica polimorfismo: la UI llama métodos
// (GetInbox, ReadMail, SendMail, GetSent, SearchMail) sin saber qué
// proveedor concreto está detrás.
package providers

// Mail representa un correo electrónico en forma simple.
// Todos los campos son strings para que sea fácil de mostrar en consola.
type Mail struct {
	Id      string // Identificador único del correo dentro del proveedor
	From    string // Remitente del correo
	To      string // Destinatario(s) del correo
	Subject string // Asunto del correo
	Body    string // Cuerpo o contenido del correo
	Date    string // Fecha en que se recibió o envió el correo
}

// EmailProvider es el contrato que todos los proveedores deben cumplir.
// Cada proveedor (Gmail, Outlook, Yahoo, Institucional) implementa estos
// métodos a su manera, pero la UI solo conoce esta interfaz.
type EmailProvider interface {
	// GetInbox devuelve la lista de correos de la bandeja de entrada.
	GetInbox() ([]Mail, error)

	// ReadMail devuelve un correo específico identificado por su id.
	ReadMail(id string) (Mail, error)

	// SendMail envía un correo nuevo al destinatario indicado.
	SendMail(to string, subject string, body string) error

	// GetSent devuelve la lista de correos enviados.
	GetSent() ([]Mail, error)

	// SearchMail busca correos que coincidan con la consulta.
	SearchMail(query string) ([]Mail, error)
}
