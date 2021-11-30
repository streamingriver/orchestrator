# Opensource IPTV orchestrator

json to push:
```json
{
 "name": "test",
 "image": "alpine:latest",
 "ports": {
  "127.0.0.1:8080": "8080/udp"
 },
 "env": [
  "HELLO=WORLD"
 ],
 "cmd": [
  "echo",
  "hello from other side"
 ],
 "auth": {
  "username": "user",
  "password": "password"
 },
 "ServiceConfig": "[base64encoded_data]"
}
```
