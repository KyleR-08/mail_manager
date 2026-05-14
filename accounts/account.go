// Define el modelo de datos de una cuenta de correo registrada en la app.
// Una cuenta guarda: email del usuario, proveedor (gmail/outlook/yahoo/
// institucional), ruta del archivo de token OAuth2 y metadatos básicos.
// Es la representación serializable que se persiste en data/accounts.json.
package accounts

// Account representa una cuenta de correo registrada en la aplicación.
// Se guarda en data/accounts.json como parte de una lista de cuentas.
type Account struct {
	Email        string `json:"email"`         // Correo del usuario (ej. usuario@gmail.com)
	ProviderType string `json:"provider_type"` // Tipo: "gmail", "outlook", "yahoo" o "institutional"
	DisplayName  string `json:"display_name"`  // Nombre amigable para mostrar al usuario
	TokenFile    string `json:"token_file"`    // Ruta al archivo de token OAuth2 de esta cuenta
}
