// Implementación de EmailProvider para correos institucionales (ej. @uide.edu.ec).
// Detecta automáticamente, a partir del dominio del correo, si la cuenta
// pertenece a Google Workspace (usa Gmail API) o a Microsoft 365 (usa
// Microsoft Graph API), y delega en el proveedor concreto correspondiente.
package providers
