// Implementación de EmailProvider para cuentas de Yahoo Mail.
// Yahoo no expone una API moderna estable para terceros, por lo que se
// utiliza IMAP como mecanismo de respaldo (fallback) para leer y buscar
// correos. El envío se hace por SMTP autenticado.
package providers
