// Administra el ciclo de vida de las cuentas registradas en la app.
// Se encarga de cargar y guardar data/accounts.json, agregar nuevas
// cuentas, eliminarlas, listarlas y devolver la cuenta seleccionada
// para que se construya el EmailProvider adecuado.
package accounts

import (
	"encoding/json"
	"fmt"
	"os"
)

// accountsFile es la ruta del archivo donde se guardan las cuentas.
const accountsFile = "data/accounts.json"

// LoadAccounts lee data/accounts.json y devuelve la lista de cuentas.
// Si el archivo no existe, lo crea vacío ("[]") y devuelve una lista vacía.
// Si hay otro error (JSON inválido, permiso, etc.) se imprime el error.
func LoadAccounts() []Account {
	// Lista vacía por defecto, se devuelve si algo falla
	var listaCuentas []Account

	// Asegurar que la carpeta data/ exista antes de leer o crear el archivo
	err := os.MkdirAll("data", 0755)
	if err != nil {
		fmt.Println("No se pudo crear la carpeta data/:", err)
		return listaCuentas
	}

	// Si el archivo no existe, crearlo con una lista vacía y salir
	_, err = os.Stat(accountsFile)
	if os.IsNotExist(err) {
		fmt.Println("No existe", accountsFile, "- creando uno nuevo vacío.")
		err = os.WriteFile(accountsFile, []byte("[]"), 0644)
		if err != nil {
			fmt.Println("No se pudo crear el archivo de cuentas:", err)
		}
		return listaCuentas
	}

	// Leer todo el contenido del archivo accounts.json
	data, err := os.ReadFile(accountsFile)
	if err != nil {
		fmt.Println("No se pudo leer el archivo de cuentas:", err)
		return listaCuentas
	}

	// Si el archivo está vacío, devolver lista vacía sin error
	if len(data) == 0 {
		return listaCuentas
	}

	// Convertir el JSON leído a la lista de cuentas
	err = json.Unmarshal(data, &listaCuentas)
	if err != nil {
		fmt.Println("No se pudo interpretar el JSON de cuentas:", err)
		return listaCuentas
	}

	// Devolver la lista cargada
	return listaCuentas
}

// SaveAccounts recibe una lista de cuentas y la guarda en data/accounts.json.
// Si hay un error se imprime con fmt.
func SaveAccounts(listaCuentas []Account) {
	// Asegurar que la carpeta data/ exista antes de escribir
	err := os.MkdirAll("data", 0755)
	if err != nil {
		fmt.Println("No se pudo crear la carpeta data/:", err)
		return
	}

	// Convertir la lista de cuentas a JSON con indentación bonita
	data, err := json.MarshalIndent(listaCuentas, "", "  ")
	if err != nil {
		fmt.Println("No se pudo convertir las cuentas a JSON:", err)
		return
	}

	// Escribir el JSON al archivo
	err = os.WriteFile(accountsFile, data, 0644)
	if err != nil {
		fmt.Println("No se pudo guardar el archivo de cuentas:", err)
		return
	}
}
