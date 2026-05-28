// Interfaz de usuario en consola.
// Muestra el menú numerado al usuario, lee la opción que elige por teclado
// y la traduce en una llamada a un EmailProvider a través de main.go.
// No contiene lógica de proveedores ni de OAuth: solo entrada y salida.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"mail_manager/accounts"
	"mail_manager/providers"
)

// lector global para leer líneas completas desde el teclado.
// Se declara una sola vez para no recrearlo en cada llamada.
var lector = bufio.NewReader(os.Stdin)

// ShowMainMenu imprime el menú principal junto con la lista de cuentas
// registradas. No lee ninguna opción: solo muestra el menú.
func ShowMainMenu(listaCuentas []accounts.Account) {
	fmt.Println("")
	fmt.Println("===== correo-manager =====")

	// Mostrar las cuentas registradas (si hay)
	if len(listaCuentas) == 0 {
		fmt.Println("(No hay cuentas registradas todavía)")
	} else {
		fmt.Println("Cuentas registradas:")
		for i := 0; i < len(listaCuentas); i++ {
			cuenta := listaCuentas[i]
			fmt.Printf("   - %s (%s)\n", cuenta.Email, cuenta.ProviderType)
		}
	}

	// Opciones disponibles del menú
	fmt.Println("")
	fmt.Println("1) Ver bandeja de entrada")
	fmt.Println("2) Leer correo")
	fmt.Println("3) Enviar correo")
	fmt.Println("4) Ver enviados")
	fmt.Println("5) Buscar correo")
	fmt.Println("6) Agregar cuenta")
	fmt.Println("7) Salir")
}

// GetUserInput muestra un texto y lee una línea completa desde el teclado.
// Se usa para pedir asuntos, destinatarios, búsquedas, emails, etc.
func GetUserInput(prompt string) string {
	// Mostrar el texto que indica qué se está pidiendo
	fmt.Print(prompt)

	// Leer toda la línea que escriba el usuario
	texto, err := lector.ReadString('\n')
	if err != nil {
		fmt.Println("Error al leer la entrada:", err)
		return ""
	}

	// Quitar espacios y saltos de línea al inicio y final
	texto = strings.TrimSpace(texto)
	return texto
}

// ShowAccountList imprime la lista de cuentas registradas numeradas para
// que el usuario pueda elegir una. Si no hay cuentas, avisa.
func ShowAccountList(listaCuentas []accounts.Account) {
	// Si no hay cuentas no hay nada que mostrar
	if len(listaCuentas) == 0 {
		fmt.Println("No hay cuentas registradas.")
		return
	}

	// Imprimir cada cuenta con un índice empezando en 1
	fmt.Println("Cuentas disponibles:")
	for i := 0; i < len(listaCuentas); i++ {
		cuenta := listaCuentas[i]
		fmt.Printf("  %d) %s (%s) - %s\n", i+1, cuenta.Email, cuenta.ProviderType, cuenta.DisplayName)
	}
}

// ShowInbox imprime una lista de correos de la bandeja de entrada.
// Cada correo se muestra con índice, asunto, remitente, fecha e id.
func ShowInbox(correos []providers.Mail) {
	// Encabezado de la sección
	fmt.Println("")
	fmt.Println("--- Bandeja de entrada ---")

	// Si no hay correos, avisar y salir
	if len(correos) == 0 {
		fmt.Println("(No hay correos en la bandeja)")
		return
	}

	// Imprimir cada correo numerado
	for i := 0; i < len(correos); i++ {
		c := correos[i]
		fmt.Printf("[%d] %s\n", i+1, c.Subject)
		fmt.Printf("    De:    %s\n", c.From)
		fmt.Printf("    Fecha: %s\n", c.Date)
		fmt.Printf("    Id:    %s\n", c.Id)
	}
}

// ShowSent imprime una lista de correos enviados con el mismo formato
// que ShowInbox, pero con un encabezado distinto.
func ShowSent(correos []providers.Mail) {
	// Encabezado de la sección de enviados
	fmt.Println("")
	fmt.Println("--- Correos enviados ---")

	// Si no hay correos, avisar y salir
	if len(correos) == 0 {
		fmt.Println("(No hay correos enviados)")
		return
	}

	// Imprimir cada correo numerado
	for i := 0; i < len(correos); i++ {
		c := correos[i]
		fmt.Printf("[%d] %s\n", i+1, c.Subject)
		fmt.Printf("    Para:  %s\n", c.To)
		fmt.Printf("    Fecha: %s\n", c.Date)
		fmt.Printf("    Id:    %s\n", c.Id)
	}
}

// ShowMail imprime un correo completo con todos sus campos formateados.
func ShowMail(correo providers.Mail) {
	// Encabezado del detalle del correo
	fmt.Println("")
	fmt.Println("--- Correo ---")
	fmt.Println("De:     ", correo.From)
	fmt.Println("Para:   ", correo.To)
	fmt.Println("Fecha:  ", correo.Date)
	fmt.Println("Asunto: ", correo.Subject)
	fmt.Println("")
	fmt.Println(correo.Body)
	fmt.Println("--- fin del correo ---")
}

// GetUserChoice pide un número al usuario entre min y max y lo devuelve.
// Si el usuario escribe algo inválido o fuera de rango, vuelve a pedirlo.
func GetUserChoice(prompt string, min int, max int) int {
	// Bucle infinito hasta que el usuario escriba algo válido
	for {
		// Pedir el número usando GetUserInput
		texto := GetUserInput(prompt)

		// Intentar convertir el texto a número entero
		numero, err := strconv.Atoi(texto)
		if err != nil {
			fmt.Println("Por favor escribe un número válido.")
			continue
		}

		// Verificar que esté en el rango permitido
		if numero < min || numero > max {
			fmt.Printf("El número debe estar entre %d y %d.\n", min, max)
			continue
		}

		// Si pasó las dos validaciones, devolverlo
		return numero
	}
}
