# Compilar Música 6.2.6

Se incluye el código de la interfaz en app y el servidor Go en src. Los directorios ya están separados. Requiere Go 1.22 o posterior y Python 3; no necesita dependencias externas de Go.

Desde codigo_fuente ejecute python3 tools/compilar.py --platform windows para generar el servidor, el iniciador, el desinstalador y el instalador de Windows, con las fuentes y configuraciones de texto incluidas, sin fondos ni sellos. En Windows puede usar python en lugar de python3. El script crea el paquete interno que no se distribuye duplicado.

Para Linux amd64 ejecute python3 tools/compilar.py --platform linux. Use la carpeta app y los ejemplos de despliegue. Configure un dominio HTTPS y almacenamiento persistente.

Ejecute primero tools/compilar.py --platform windows para generar el paquete interno que necesita el instalador; después ejecute las pruebas desde src con go test ./.... El ZIP omite ese paquete generado para evitar duplicarlo. No suba data, respaldos, claves ni configuraciones privadas a GitHub. Los ejemplos de configuración están incluidos; el paquete no contiene cuentas ni datos personales.

El proceso ejecuta tools/preparar_web.py para regenerar los paquetes web y sus archivos gzip. El archivo src/cmd/server/defaults618.json contiene únicamente las plantillas de texto de fábrica. Nunca coloque el respaldo del usuario en el código fuente ni en el paquete de instalación.

Pruebas de la proyección de cuentas y permisos de la interfaz: node tools/probar_interfaz.cjs (Node.js 18 o posterior). No es un requisito del equipo destinatario.
