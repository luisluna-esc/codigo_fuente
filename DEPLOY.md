# Despliegue de Música

Música va en el mismo servidor que InvenTrack y FundTrack (`148.230.73.226`, Ubuntu, Docker) y se abre en
**http://148.230.73.226:8092**.

Es el sistema actual (6.2.6) tal como está en producción. La reconstrucción se hace en proyectos aparte.

## Cómo está armado

| Pieza | Dónde | Detalle |
|---|---|---|
| Aplicación | contenedor `musica` | Servidor Go que entrega la interfaz y el API. Puerto local `127.0.0.1:5092`. |
| Datos | `/opt/musica/data` | Montado en `/data`. Configuración, clave de instalación, datos, respaldos y log. **No va a Git.** Se conserva al actualizar. |
| Entrada pública | nginx del servidor, `/etc/nginx/sites-available/musica.conf` | Escucha en el 8092 y pasa todo a `127.0.0.1:5092`. Copia de `despliegue/nginx-musica.conf`. |
| Código | `/opt/musica/codigo` | Clon de GitHub desde donde se construye la imagen. |

No usa PostgreSQL: el sistema guarda todo en `/opt/musica/data/data/`.

El log está en `/opt/musica/data/ServidorMusica.log` (`docker logs` sale casi vacío).

## Primera instalación

1. Revisa que los puertos estén libres (no debe salir nada):

   ```bash
   ss -ltnp | grep -E ':(8092|5092) '
   ```

   Si alguno está ocupado, cambia `8092` en `despliegue/nginx-musica.conf` y `5092` ahí y en `deploy.sh`.

2. Baja el código y prepara la carpeta de datos (el contenedor corre con el usuario 10001):

   ```bash
   mkdir -p /opt/musica/data
   git clone <url-del-repo> /opt/musica/codigo
   chown 10001:10001 /opt/musica/data
   chmod 700 /opt/musica/data
   ```

3. Entrada pública en nginx:

   ```bash
   cp /opt/musica/codigo/despliegue/nginx-musica.conf /etc/nginx/sites-available/musica.conf
   ln -s /etc/nginx/sites-available/musica.conf /etc/nginx/sites-enabled/musica.conf
   nginx -t && systemctl reload nginx
   ```

   Si el servidor usa firewall: `ufw allow 8092/tcp`.

4. Construye y arranca:

   ```bash
   bash /opt/musica/codigo/deploy.sh
   ```

5. Abre http://148.230.73.226:8092. Pide la clave de instalación:

   ```bash
   cat /opt/musica/data/CLAVE_INSTALACION.txt
   ```

   Crea el Administrador con ella. La clave se borra sola después de usarla.

6. Restaura los datos de producción:
   - Extrae el `.json` del RAR de respaldo **en tu equipo** (no subas el RAR al servidor ni a Git).
   - En la aplicación, abre **Importar datos de Música**, selecciona el `.json`, pulsa **Ver qué se importará**,
     revisa el resumen y confirma.

7. En **Configuración**, pon como dirección pública para los QR: `http://148.230.73.226:8092`.

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
   curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8092/
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

## Opcional: Google y recuperación por correo

Crea `/opt/musica/data/acceso-servidor.json` a partir de `acceso-servidor.example.json`, con permisos
`chmod 600` y dueño `10001`, y reinicia con `docker restart musica`. Detalles en `CONFIGURACION_ACCESO.md`.
Google exige HTTPS con dominio; con la IP y el puerto solo funciona usuario y contraseña.

## Notas

- Sin dominio ni HTTPS las contraseñas viajan sin cifrar, igual que en InvenTrack y FundTrack por IP.
  Cuando haya dominio, conviene poner HTTPS en el nginx del servidor.
- Si el contenedor se cae, al arrancar borra el bloqueo `musica-server.lock` que dejó la ejecución
  anterior (`despliegue/docker-entrypoint.sh`); sin eso el servidor no volvería a iniciar.
