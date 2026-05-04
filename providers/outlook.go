// Implementación de EmailProvider para cuentas de Outlook / Hotmail / Live.
// Se conecta a Microsoft Graph API mediante OAuth2 con los endpoints de
// Microsoft Identity Platform y traduce las operaciones genéricas a las
// llamadas REST correspondientes de Graph (/me/messages, /me/sendMail, etc).
package providers
