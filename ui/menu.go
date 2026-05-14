// Interfaz de usuario en consola.
// Muestra el menú numerado al usuario, lee la opción que elige por teclado
// y la traduce en una llamada a un EmailProvider a través de la sesión activa.
// No contiene lógica de proveedores ni de OAuth: solo entrada y salida.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ShowMainMenu muestra el menú principal y devuelve la opción que escribió
// el usuario como texto. Por ahora solo imprime y lee, sin lógica detrás.
func ShowMainMenu() string {
	// Imprimir el título y las opciones disponibles
	fmt.Println("")
	fmt.Println("===== correo-manager =====")
	fmt.Println("1) Agregar cuenta")
	fmt.Println("2) Seleccionar cuenta")
	fmt.Println("3) Salir")
	fmt.Print("Elige una opción: ")

	// Crear un lector para leer texto desde el teclado
	lector := bufio.NewReader(os.Stdin)

	// Leer la línea que escribió el usuario
	opcion, err := lector.ReadString('\n')
	if err != nil {
		fmt.Println("Error al leer la opción:", err)
		return ""
	}

	// Quitar espacios y saltos de línea al inicio y final
	opcion = strings.TrimSpace(opcion)

	// Devolver la opción tal cual la escribió el usuario
	return opcion
}
