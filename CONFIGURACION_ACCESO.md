# Configuración del acceso por correo y Google

El acceso con usuario o correo y contraseña funciona con las cuentas del sistema. Google y los mensajes de recuperación requieren credenciales propias del servidor. No incluya credenciales reales en GitHub ni en un ZIP que vaya a compartir.

## Archivo de configuración privada

Copie acceso-servidor.example.json como acceso-servidor.json junto a server-config.json. Complete únicamente los servicios que utilizará y reinicie el servidor. Las variables de entorno con los mismos nombres prevalecen sobre ese archivo. En Linux, limite el archivo al usuario del servicio con chmod 600. En Windows, limite su acceso al usuario que ejecuta Música.

MUSICA_PUBLIC_URL debe ser el origen público de la aplicación, por ejemplo https://musica.su-organizacion.org, sin ruta, consulta ni barra final. Se usa en los enlaces de recuperación y como dirección de los QR cuando no hay otra definida en Configuración. Administración y Secretaría pueden definir la dirección pública para QR desde la aplicación; ese campo no sustituye la configuración de Google y correo. Para una prueba local se admite http://127.0.0.1:8796. No use una dirección local para enviar enlaces que abrirán otras personas.

## Continuar con Google

Cree un proyecto de Google Cloud y una aplicación OAuth de tipo Web. Configure la pantalla de consentimiento y, si el proyecto sigue en modo de pruebas, agregue los usuarios de prueba autorizados.

Registre exactamente este URI de redirección, sustituyendo el dominio:

https://musica.su-organizacion.org/api/auth610/google/callback

No agregue /api/musica al URI registrado: el servidor admite el callback público directamente.

Complete MUSICA_GOOGLE_CLIENT_ID y MUSICA_GOOGLE_CLIENT_SECRET. Se solicitan únicamente openid, email y profile. No se habilita Google Drive ni se crean usuarios desde Google.

El correo debe coincidir con una única cuenta activa y habilitada en Música. El primer vínculo directo admite Gmail y organizaciones de Google verificadas. Para un correo externo que no sea de una organización verificada, use correo y contraseña del sistema y la recuperación por correo. Los cambios de correo invalidan el vínculo previo de identidad.

El flujo usa código de autorización, PKCE, estado y nonce. El servidor comprueba la firma RS256 y los datos de emisor, destinatario, vencimiento y correo verificado. La sesión se entrega al almacenamiento de la pestaña; no se pone el token de sesión en la URL.

Documentación oficial: https://developers.google.com/identity/openid-connect/openid-connect

## Recuperación por correo

Complete MUSICA_SMTP_HOST, MUSICA_SMTP_PORT, MUSICA_SMTP_FROM y, si su servidor lo requiere, MUSICA_SMTP_USER y MUSICA_SMTP_PASSWORD. El puerto predeterminado es 587. El servidor exige STARTTLS con verificación del certificado; no admite SMTP sin cifrar ni el modo TLS implícito del puerto 465.

MUSICA_SMTP_FROM debe ser una dirección remitente autorizada por el proveedor. Si el proveedor exige una contraseña de aplicación o credencial SMTP específica, use esa credencial. El usuario puede tener Gmail, Outlook u otro correo registrado: el proveedor de su cuenta no obliga a usar el mismo proveedor para el envío SMTP.

Los enlaces duran 30 minutos y se utilizan una sola vez. Se guarda únicamente la huella del token, vinculada a la cuenta, su correo, la contraseña vigente y la generación de los datos. El restablecimiento invalida las sesiones previas. Se limita la frecuencia de solicitudes. La respuesta no confirma si un correo específico pertenece a una cuenta.

La configuración debe verificarse en el servidor real con una cuenta de prueba ya registrada: solicitar el enlace, recibirlo, cambiar la contraseña y comprobar que el enlace no vuelve a funcionar. Esta entrega no contiene credenciales ni certifica la conexión con un proveedor real de Google o SMTP.

Referencia de recuperación de contraseñas: https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
