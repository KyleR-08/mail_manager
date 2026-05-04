// Define la interfaz EmailProvider que abstrae cualquier proveedor de correo.
// Es el contrato común que implementan Gmail, Outlook, Yahoo e Institucional.
// Gracias a esta interfaz se aplica polimorfismo: la UI llama métodos
// (GetInbox, ReadMail, SendMail, GetSent, SearchMail) sin saber qué
// proveedor concreto está detrás.
package providers
