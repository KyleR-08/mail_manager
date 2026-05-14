// Punto de entrada de la aplicación correo-manager.
// Aquí se inicia el programa: se cargan las cuentas registradas, se elige
// (o se crea) una cuenta, se construye el EmailProvider correspondiente y
// se entrega el control al menú de consola.
package main

import (
	"fmt"

	"mail_manager/accounts"
	"mail_manager/ui"
)

func main() {
	// Mensaje de bienvenida al iniciar la aplicación
	fmt.Println("Iniciando correo-manager...")

	// Cargar las cuentas registradas desde data/accounts.json
	listaCuentas := accounts.LoadAccounts()

	// Mostrar al usuario cuántas cuentas se cargaron
	fmt.Println("Cuentas cargadas:", len(listaCuentas))

	// Mostrar el menú principal y leer la opción del usuario
	opcion := ui.ShowMainMenu()

	// Por ahora solo imprimimos la opción que eligió, sin lógica detrás
	fmt.Println("Opción elegida:", opcion)
}
