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
// Si hay un error (por ejemplo el archivo no existe) se imprime y se
// devuelve una lista vacía.
func LoadAccounts() []Account {
	// Lista vacía por defecto, se devuelve si algo falla
	var listaCuentas []Account

	// Leer todo el contenido del archivo accounts.json
	data, err := os.ReadFile(accountsFile)
	if err != nil {
		fmt.Println("No se pudo leer el archivo de cuentas:", err)
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
