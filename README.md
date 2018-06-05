#Sigapi

Sigapi es un simple servidor HTTP para servir la tabla `student` de SIGENU. La consulta que hace es `SELECT id_student, t,identification,name,middle_name,last_name,address,phone FROM student LIMIT X OFFSET Y` donde X y Y deben ser proporcionados por el usuario. Las restricciones `LIMIT X` y `OFFSET Y` sirven para paginar. Con la API corriendo en `localhost:8080` la URL para consultar es `http://localhost:8080/?offset=Y&size=X`, donde `X` e `Y` deben reemplazarse por números.

## Despliegue

```conf
[Unit]
Description=Sigapi Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/dir
ExecStart=/usr/local/bin/sigapi -a postgresql.server/database -u usuario -p contraseña -s :8081 -l doc.html
Restart=on-abort

[Install]
WantedBy=multi-user.target
```
