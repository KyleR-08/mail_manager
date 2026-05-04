// Implementación de EmailProvider para cuentas de Gmail.
// Se conecta a la Gmail API v1 usando OAuth2 (google.golang.org/api/gmail/v1)
// y traduce las operaciones genéricas (bandeja, lectura, envío, búsqueda)
// a llamadas concretas de la API de Google.
package providers
