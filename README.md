# Ghisa: ultra simple proxy service
Ghisa is a 0-security proxy, implemented for sole purpose of bypassing CORS errors.
Not recommended for production, unless you're a fan of Stephen King's inspired live role playing.

### El nòster ghisa che el fa passà el tràfegh indove che vorèmm nun alter.
![Ghisa](https://img3.stcrm.it/images/18134478/550x/20190124-220035117-6424.jpg)

## 🎮 Try it

Currently deployed on `https://ghisa.forkbomb.eu/`

## 🕹 Run

The service will run on port `5552` to run it just:

```bash
./ghisa
```

## 🐳 Deploy with Coolify

`docker-compose.yml` defines a single `ghisa` service. In Coolify create a Service from this Docker Compose definition, then set the hostname under the component's **Domains** field (e.g. `ghisa.forkbomb.eu`).

- The container listens on internal port `5552`.
- `SERVICE_FQDN_GHISA_5552` makes Coolify route the component's domain to that internal port.
- Coolify's reverse proxy terminates TLS, so the public URL is `https://<domain>` and the proxy forwards plain HTTP to the container. Ports are intentionally not published.

For a local run, publish the port yourself:

```bash
docker build -t ghisa .
docker run --rm -p 5552:5552 ghisa
```

The binary honours the `PORT` environment variable, falling back to `5552`.
