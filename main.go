// Punto de entrada de la aplicación correo-manager.
// Carga las cuentas registradas, permite agregar una cuenta nueva si no hay,
// autentica al usuario contra el proveedor correspondiente, construye el
// EmailProvider concreto y entrega el control al menú de consola en un loop.
//
// CONCEPTO CLAVE — POLIMORFISMO:
// La variable "proveedor" se declara con el tipo de la interfaz
// providers.EmailProvider, pero adentro puede vivir un *GmailProvider,
// un *OutlookProvider o un *InstitutionalProvider. El resto del código
// llama a proveedor.GetInbox() / proveedor.SendMail() sin saber cuál es.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"mail_manager/accounts"
	"mail_manager/auth"
	"mail_manager/providers"
	"mail_manager/ui"
)

// detectProvider devuelve el tipo de proveedor según el dominio del email.
// Reglas simples: @gmail.com -> gmail, @hotmail/@outlook -> outlook,
// @uide.edu.ec -> institutional.
func detectProvider(email string) string {
	// Pasar a minúsculas para comparar sin importar mayúsculas
	correo := strings.ToLower(email)

	// Detectar Gmail por su dominio
	if strings.HasSuffix(correo, "@gmail.com") {
		return "gmail"
	}

	// Detectar Outlook/Hotmail/Live por su dominio
	if strings.HasSuffix(correo, "@hotmail.com") || strings.HasSuffix(correo, "@outlook.com") || strings.HasSuffix(correo, "@live.com") {
		return "outlook"
	}

	// Detectar correo institucional UIDE
	if strings.HasSuffix(correo, "@uide.edu.ec") {
		return "institutional"
	}

	// Dominio desconocido
	return ""
}

// agregarCuenta pide los datos al usuario y guarda la nueva cuenta en disco.
// Devuelve la cuenta creada (o una vacía si algo falló).
func agregarCuenta(listaCuentas []accounts.Account) ([]accounts.Account, accounts.Account) {
	// Pedir el email al usuario
	email := ui.GetUserInput("Escribe el email de la cuenta: ")

	// Detectar el tipo de proveedor por el dominio
	tipo := detectProvider(email)
	if tipo == "" {
		fmt.Println("Dominio no soportado. Solo se acepta @gmail.com, @hotmail.com, @outlook.com, @live.com o @uide.edu.ec.")
		return listaCuentas, accounts.Account{}
	}

	// Pedir un nombre amigable para mostrar
	nombre := ui.GetUserInput("Nombre para mostrar (puede ser tu nombre): ")

	// Construir la cuenta nueva con la ruta del token derivada del email
	rutaToken := "token/" + strings.ReplaceAll(email, "@", "_at_") + ".json"
	cuentaNueva := accounts.Account{
		Email:        email,
		ProviderType: tipo,
		DisplayName:  nombre,
		TokenFile:    rutaToken,
	}

	// Agregar la cuenta a la lista y guardarla en disco
	listaCuentas = append(listaCuentas, cuentaNueva)
	accounts.SaveAccounts(listaCuentas)
	fmt.Println("Cuenta registrada:", email, "(", tipo, ")")
	return listaCuentas, cuentaNueva
}

// buildProvider construye el EmailProvider concreto según el tipo de cuenta.
// AQUÍ ES DONDE OCURRE EL POLIMORFISMO: la función devuelve un EmailProvider
// (interfaz), pero adentro pone un proveedor distinto según el caso.
func buildProvider(cuenta accounts.Account) (providers.EmailProvider, error) {
	// Caso Gmail: usar Google OAuth2 + GmailProvider
	if cuenta.ProviderType == "gmail" {
		// Configurar la autenticación de Google
		cfg := auth.GoogleAuthConfig{
			CredentialsFile: "credentials.json",
			TokenFile:       cuenta.TokenFile,
		}

		// Obtener el cliente HTTP autenticado
		client, err := auth.GetGoogleClient(cfg)
		if err != nil {
			return nil, err
		}

		// Devolver el proveedor Gmail (asignable a EmailProvider)
		return providers.NewGmailProvider(client), nil
	}

	// Caso Outlook: usar Microsoft OAuth2 con tenant "consumers"
	if cuenta.ProviderType == "outlook" {
		// Leer credenciales de la app desde variables de entorno
		cfg := auth.MicrosoftAuthConfig{
			ClientID:     os.Getenv("MS_CLIENT_ID"),
			ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
			TenantID:     "consumers",
			TokenFile:    cuenta.TokenFile,
		}

		// Obtener el cliente HTTP autenticado
		client, err := getMicrosoftClientSafe(cfg)
		if err != nil {
			return nil, err
		}

		// Devolver el proveedor Outlook
		return providers.NewOutlookProvider(client), nil
	}

	// Caso Institucional: mismo flujo que Outlook pero con tenant UIDE
	if cuenta.ProviderType == "institutional" {
		// Tenant institucional configurable por variable de entorno
		tenant := os.Getenv("UIDE_TENANT_ID")
		if tenant == "" {
			// Si no está definido, se intenta usar "organizations" como genérico
			tenant = "organizations"
		}

		// Configurar la autenticación contra el tenant institucional
		cfg := auth.MicrosoftAuthConfig{
			ClientID:     os.Getenv("MS_CLIENT_ID"),
			ClientSecret: os.Getenv("MS_CLIENT_SECRET"),
			TenantID:     tenant,
			TokenFile:    cuenta.TokenFile,
		}

		// Obtener el cliente HTTP autenticado
		client, err := getMicrosoftClientSafe(cfg)
		if err != nil {
			return nil, err
		}

		// Devolver el proveedor Institucional (también satisface EmailProvider)
		return providers.NewInstitutionalProvider(client), nil
	}

	// Tipo desconocido
	return nil, fmt.Errorf("tipo de proveedor desconocido: %s", cuenta.ProviderType)
}

// getMicrosoftClientSafe envuelve auth.GetMicrosoftClient con validación
// previa de las credenciales requeridas en variables de entorno.
func getMicrosoftClientSafe(cfg auth.MicrosoftAuthConfig) (*http.Client, error) {
	// Verificar que el ClientID y el ClientSecret estén configurados
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("faltan variables MS_CLIENT_ID o MS_CLIENT_SECRET")
	}
	return auth.GetMicrosoftClient(cfg)
}

// mostrarCorreos imprime una lista de correos numerada en consola.
func mostrarCorreos(correos []providers.Mail) {
	// Si no hay correos, avisar al usuario
	if len(correos) == 0 {
		fmt.Println("No hay correos para mostrar.")
		return
	}

	// Imprimir cada correo con un índice
	for i := 0; i < len(correos); i++ {
		c := correos[i]
		fmt.Printf("[%d] %s\n", i+1, c.Subject)
		fmt.Printf("    De: %s\n", c.From)
		fmt.Printf("    Fecha: %s\n", c.Date)
		fmt.Printf("    Id: %s\n", c.Id)
	}
}

func main() {
	// Mensaje de bienvenida
	fmt.Println("Iniciando correo-manager...")

	// Cargar las cuentas registradas desde disco
	listaCuentas := accounts.LoadAccounts()
	fmt.Println("Cuentas cargadas:", len(listaCuentas))

	// Si no hay cuentas, pedir al usuario que agregue una
	if len(listaCuentas) == 0 {
		fmt.Println("No hay cuentas registradas. Vamos a agregar una.")
		nuevaLista, _ := agregarCuenta(listaCuentas)
		listaCuentas = nuevaLista

		// Si igual no se pudo agregar, salir
		if len(listaCuentas) == 0 {
			fmt.Println("No se agregó ninguna cuenta. Saliendo.")
			return
		}
	}

	// Usar la primera cuenta como cuenta activa
	cuentaActiva := listaCuentas[0]
	fmt.Println("Cuenta activa:", cuentaActiva.Email)

	// Construir el proveedor concreto detrás de la interfaz EmailProvider.
	// POLIMORFISMO: 'proveedor' es del tipo de la interfaz, no del tipo concreto.
	var proveedor providers.EmailProvider
	proveedor, err := buildProvider(cuentaActiva)
	if err != nil {
		fmt.Println("No se pudo construir el proveedor:", err)
		fmt.Println("Se seguirá mostrando el menú, pero las operaciones de correo fallarán.")
	}

	// Loop principal del menú
	for {
		// Mostrar el menú principal con la lista de cuentas
		ui.ShowMainMenu(listaCuentas)

		// Pedir la opción al usuario (entre 1 y 7)
		opcion := ui.GetUserChoice("Elige una opción: ", 1, 7)

		// Opción 1: Ver bandeja de entrada
		if opcion == 1 {
			if proveedor == nil {
				fmt.Println("No hay proveedor disponible.")
				continue
			}
			correos, err := proveedor.GetInbox()
			if err != nil {
				fmt.Println("Error al obtener la bandeja de entrada:", err)
				continue
			}
			mostrarCorreos(correos)
			continue
		}

		// Opción 2: Leer correo por id
		if opcion == 2 {
			if proveedor == nil {
				fmt.Println("No hay proveedor disponible.")
				continue
			}
			id := ui.GetUserInput("Id del correo a leer: ")
			correo, err := proveedor.ReadMail(id)
			if err != nil {
				fmt.Println("Error al leer el correo:", err)
				continue
			}
			fmt.Println("--- Correo ---")
			fmt.Println("De:     ", correo.From)
			fmt.Println("Para:   ", correo.To)
			fmt.Println("Fecha:  ", correo.Date)
			fmt.Println("Asunto: ", correo.Subject)
			fmt.Println("")
			fmt.Println(correo.Body)
			continue
		}

		// Opción 3: Enviar correo
		if opcion == 3 {
			if proveedor == nil {
				fmt.Println("No hay proveedor disponible.")
				continue
			}
			destino := ui.GetUserInput("Para: ")
			asunto := ui.GetUserInput("Asunto: ")
			cuerpo := ui.GetUserInput("Cuerpo: ")
			err := proveedor.SendMail(destino, asunto, cuerpo)
			if err != nil {
				fmt.Println("Error al enviar el correo:", err)
				continue
			}
			fmt.Println("Correo enviado con éxito.")
			continue
		}

		// Opción 4: Ver enviados
		if opcion == 4 {
			if proveedor == nil {
				fmt.Println("No hay proveedor disponible.")
				continue
			}
			correos, err := proveedor.GetSent()
			if err != nil {
				fmt.Println("Error al obtener enviados:", err)
				continue
			}
			mostrarCorreos(correos)
			continue
		}

		// Opción 5: Buscar correo
		if opcion == 5 {
			if proveedor == nil {
				fmt.Println("No hay proveedor disponible.")
				continue
			}
			consulta := ui.GetUserInput("Texto a buscar: ")
			correos, err := proveedor.SearchMail(consulta)
			if err != nil {
				fmt.Println("Error al buscar:", err)
				continue
			}
			mostrarCorreos(correos)
			continue
		}

		// Opción 6: Agregar una cuenta nueva
		if opcion == 6 {
			nuevaLista, cuentaNueva := agregarCuenta(listaCuentas)
			listaCuentas = nuevaLista

			// Si se agregó algo, intentar construir su proveedor y activarlo
			if cuentaNueva.Email != "" {
				cuentaActiva = cuentaNueva
				proveedor, err = buildProvider(cuentaActiva)
				if err != nil {
					fmt.Println("No se pudo construir el proveedor para la cuenta nueva:", err)
				} else {
					fmt.Println("Cuenta activa cambiada a:", cuentaActiva.Email)
				}
			}
			continue
		}

		// Opción 7: Salir del programa
		if opcion == 7 {
			fmt.Println("Saliendo de correo-manager. ¡Hasta luego!")
			return
		}
	}
}
