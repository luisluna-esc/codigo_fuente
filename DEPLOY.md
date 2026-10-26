# Despliegue de Música

Música va en el mismo servidor que InvenTrack y FundTrack (`148.230.73.226`, Ubuntu, Docker) y se abre en
**https://musica.148-230-73-226.sslip.io**.

La aplicación exige HTTPS (sin él no deja ingresar). Como no hay dominio propio, se usa sslip.io: ese nombre
apunta solo a la IP del servidor y Let's Encrypt le da certificado gratis. Si más adelante hay dominio, se
cambia `server_name` y se vuelve a ejecutar certbot con el nuevo nombre; los datos no se tocan.

Es el sistema actual (6.2.6) tal como está en producción. La reconstrucción se hace en proyectos aparte.

## Cómo está armado

| Pieza | Dónde | Detalle |
|---|---|---|
| Aplicación | contenedor `musica` | Servidor Go que entrega la interfaz y el API. Puerto local `127.0.0.1:5092`. |
| Datos | `/opt/musica/data` | Montado en `/data`. Configuración, clave de instalación, datos, respaldos y log. **No va a Git.** Se conserva al actualizar. |
| Entrada pública | nginx del servidor, `/etc/nginx/sites-available/musica-https.conf` | HTTPS en el 443 por nombre (`musica.148-230-73-226.sslip.io`), pasa todo a `127.0.0.1:5092`. Parte de `despliegue/nginx-musica.conf`; certbot le agregó el certificado. |
| Certificado | `/etc/letsencrypt/live/musica.148-230-73-226.sslip.io/` | Let's Encrypt. certbot lo renueva solo. |
| Código | `/opt/musica/codigo` | Clon de GitHub desde donde se construye la imagen. |

Los datos se guardan en PostgreSQL, base `musica` (ver **Base de datos**). En `/opt/musica/data` quedan la
configuración, la clave de instalación, las sesiones y el log.

El log está en `/opt/musica/data/ServidorMusica.log` (`docker logs` sale casi vacío).

## Primera instalación

1. Revisa que el puerto local esté libre (no debe salir nada) y que los puertos 80 y 443 sean de nginx:

   ```bash
   ss -ltnp | grep -E ':5092 '
   ss -ltnp | grep -E ':(80|443) '
   ```

   Si el 5092 está ocupado, cámbialo en `despliegue/nginx-musica.conf` y en `deploy.sh`.

2. Baja el código y prepara la carpeta de datos (el contenedor corre con el usuario 10001):

   ```bash
   mkdir -p /opt/musica/data
   git clone <url-del-repo> /opt/musica/codigo
   chown 10001:10001 /opt/musica/data
   chmod 700 /opt/musica/data
   ```

3. Entrada pública en nginx con HTTPS:

   ```bash
   cp /opt/musica/codigo/despliegue/nginx-musica.conf /etc/nginx/sites-available/musica-https.conf
   ln -s /etc/nginx/sites-available/musica-https.conf /etc/nginx/sites-enabled/musica-https.conf
   nginx -t && systemctl reload nginx
   command -v certbot >/dev/null || (apt-get update && apt-get install -y certbot python3-certbot-nginx)
   certbot --nginx -d musica.148-230-73-226.sslip.io --redirect --agree-tos --register-unsafely-without-email -n
   ```

   Si el servidor usa firewall: `ufw allow 80/tcp && ufw allow 443/tcp`.

4. Construye y arranca:

   ```bash
   bash /opt/musica/codigo/deploy.sh
   ```

5. Abre https://musica.148-230-73-226.sslip.io. Pide la clave de instalación:

   ```bash
   cat /opt/musica/data/CLAVE_INSTALACION.txt
   ```

   Crea el Administrador con ella. La clave se borra sola después de usarla.

6. Restaura los datos de producción:
   - Extrae el `.json` del RAR de respaldo **en tu equipo** (no subas el RAR al servidor ni a Git).
   - En la aplicación, abre **Importar datos de Música**, selecciona el `.json`, pulsa **Ver qué se importará**,
     revisa el resumen y confirma.

7. En **Configuración**, pon como dirección pública para los QR: `https://musica.148-230-73-226.sslip.io`.

## Actualizar a una nueva versión

1. Haz un respaldo (ver abajo).
2. Sube los cambios a GitHub.
3. En el servidor:

   ```bash
   bash /opt/musica/codigo/deploy.sh
   ```

   El script baja lo último, construye la imagen y cambia el contenedor. Los datos no se tocan.

4. Revisa:

   ```bash
   docker ps | grep musica
   tail -n 30 /opt/musica/data/ServidorMusica.log
   curl -s -o /dev/null -w "%{http_code}\n" https://musica.148-230-73-226.sslip.io/
   ```

## Respaldos

Desde la aplicación: **Configuración → Respaldo completo**. Se guarda en `/opt/musica/data/data/respaldos/`
y se puede descargar.

Copia de toda la carpeta de datos (detiene Música unos segundos para que los archivos queden consistentes):

```bash
docker stop -t 60 musica
tar -czf /root/musica_$(date +%Y%m%d_%H%M).tar.gz -C /opt/musica data
docker start musica
```

Para volver a una copia: `docker stop musica`, reemplaza `/opt/musica/data` con el contenido del `.tar.gz`,
`chown -R 10001:10001 /opt/musica/data` y `docker start musica`.

## Base de datos

PostgreSQL del servidor, base `musica`, usuario `musica`. Contraseña en `/root/musica-db.env`; la cadena que
usa el contenedor (`MUSICA_DATABASE_URL`) en `/root/musica-app.env`. Ninguno va a Git.

| Esquema | Qué guarda | Quién escribe |
|---|---|---|
| `legacy` | Estado completo del sistema (`sgl_kv`, `sgl_files`), mismo formato que el SQLite de Windows. | El sistema, en cada cambio. |
| `public` | 51 tablas relacionales (cuentas, requisitos, entregas, archivos, plantillas, certificados...). | `despliegue/actualizar-tablas.sh`, cada 10 minutos y sólo si hubo cambios. |

Las tablas de `public` son una copia para consultas, reportes y el sistema nuevo: **no se editan a mano**,
la siguiente actualización las reemplaza. Se crean con `Scrips/dataBase.sql` y `Scrips/seedCatalogs.sql`
del proyecto `MBOS.Hub.Music`, y se llenan con `Scrips/importarRespaldo.py` (copiados en `/root/musica-scrips`).

Si PostgreSQL no responde, el sistema muestra error al guardar en lugar de guardar en otro lugar.

Respaldo de la base:

```bash
sudo -u postgres pg_dump -Fc musica > /root/musica_$(date +%Y%m%d_%H%M).dump
```

Revisar la actualización de tablas: `tail /var/lib/musica-tablas/actualizaciones.log` y `avisos.log`.

## Opcional: Google y recuperación por correo

Crea `/opt/musica/data/acceso-servidor.json` a partir de `acceso-servidor.example.json`, con permisos
`chmod 600` y dueño `10001`, y reinicia con `docker restart musica`. Detalles en `CONFIGURACION_ACCESO.md`.
Google puede rechazar un nombre sslip.io como dirección de retorno; con dominio propio no hay problema.

## Notas

- Si el contenedor se cae, al arrancar borra el bloqueo `musica-server.lock` que dejó la ejecución
  anterior (`despliegue/docker-entrypoint.sh`); sin eso el servidor no volvería a iniciar.
