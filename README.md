# Drone Tree Config

This is a Drone extension to support mono repositories with multiple `.drone.yml`.

The extension checks each changed file and looks for a `.drone.yml` in the directory of the file or any parent directory. Drone will either use the first `.drone.yml` that matches or optionally run all of them in a multi-machine build.

There is an official Docker image: https://hub.docker.com/r/bitsbeats/drone-tree-config

## Limitations

Currently supports only Github.

## Usage

Environment variables:

- `PLUGIN_CONCAT`: Concats all found configs to a multi-machine build. Defaults to `false`.
- `PLUGIN_FALLBACK`: Rebuild all .drone.yml if no changes where made. Defaults to `false`.
- `PLUGIN_MAXDEPTH`: Max depth to search for `drone.yml`, only active in fallback mode. Defaults to `2` (would still find `/a/b/.drone.yml`).
- `PLUGIN_DEBUG`: Set this to `true` to enable debug messages.
- `PLUGIN_ADDRESS`: Listen address for the plugins webserver. Defaults to `:3000`.
- `PLUGIN_SECRET`: Shared secret with drone. You can generate the token using `openssl rand -hex 16`.
- `GITHUB_TOKEN`: Github personal access token. Only needs repo rights. See [here][1].
- `GITHUB_SERVER`: Custom Github server for Github Enterprise

If `PLUGIN_CONCAT` is not set, the first `.drone.yml` will be used.

Example docker-compose:

```yaml
version: '2'
services:
  drone-server:
    image: drone/drone
    ports:
      - 8000:80
    volumes:
      - /var/lib/drone:/data
      - /var/run/docker.sock:/var/run/docker.sock
    links:
      - drone-tree-config
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_SERVER_PROTO=https
      - DRONE_SERVER_HOST=***
      - DRONE_GITHUB=true
      - DRONE_GITHUB_SERVER=https://github.com
      - DRONE_GITHUB_CLIENT_ID=***
      - DRONE_GITHUB_CLIENT_SECRET=***
      - DRONE_GIT_ALWAYS_AUTH=true
      - DRONE_SECRET=***
      - DRONE_RUNNER_CAPACITY=2

      - DRONE_YAML_ENDPOINT=http://drone-tree-config:3000
      - DRONE_YAML_SECRET=<SECRET>

  drone-tree-config:
    image: bitsbeats/drone-tree-config
    environment:
      - PLUGIN_DEBUG=true
      - PLUGIN_CONCAT=true
      - PLUGIN_FALLBACK=true
      - PLUGIN_SECRET=<SECRET>
      - GITHUB_TOKEN=<GITHUB_TOKEN>
```

Edit the Secrets (`***`), `<SECRET>` and `<GITHUB_TOKEN>` to your needs. `<SECRET>` is used between Drone and drone-tree-config.

[1]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
