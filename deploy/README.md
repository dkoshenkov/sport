# Sport Deploy

Direct server deploy without GitHub Actions.

## Topology

```text
sport.register.im
  -> Caddy on host :443
  -> 127.0.0.1:3001
  -> docker compose client nginx
  -> api container for /v1/* and /healthz
  -> postgres container
```

`client` binds only to `127.0.0.1:3001`, so it is not exposed publicly without
Caddy.

## Deploy

```bash
./deploy/deploy.sh
```

Defaults:

- host: `87.242.119.199`
- user: `project-kdo`
- key: `~/.ssh/cloud_key`
- remote app dir: `/opt/sport`

Optional overrides:

```bash
SERVER_USER=project-kdo APP_DIR=/opt/sport ./deploy/deploy.sh 87.242.119.199 22
```

The script creates `.env.prod` on the server and reuses the same generated
Postgres password on later deploys. To force a password on first deploy:

```bash
POSTGRES_PASSWORD='long_random_password' ./deploy/deploy.sh
```

## Caddy

The script installs this snippet to `/etc/caddy/conf.d/sport.register.im.caddy`:

```caddyfile
sport.register.im {
    reverse_proxy 127.0.0.1:3001
}
```

If `/etc/caddy/Caddyfile` does not already import `/etc/caddy/conf.d/*.caddy`,
the script backs it up and appends:

```caddyfile
import /etc/caddy/conf.d/*.caddy
```

Then it runs `caddy validate` and reloads Caddy.
