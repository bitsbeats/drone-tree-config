# Drone Tree Config

This is a Drone extension to support mono repositories with multiple `.drone.yml`.

The extension checks each changed file and looks for a `.drone.yml` in the directory of the file or any parent directory. Drone will either use the first `.drone.yml` that matches or optionally run all of them in a multi-machine build.

There is an official Docker image: https://hub.docker.com/r/bitsbeats/drone-tree-config

## Limitations

Currently supports

* Github
* Gitlab
* Bitbucket [#4](https://github.com/bitsbeats/drone-tree-config/pull/4)

## Usage

#### Environment variables

* `PLUGIN_CONCAT`: Concats all found configs to a multi-machine build. Defaults to `false`.
* `PLUGIN_FALLBACK`: Rebuild all .drone.yml if no changes where made. Defaults to `false`.
* `PLUGIN_MAXDEPTH`: Max depth to search for `.drone.yml`, only active in fallback mode. Defaults to `2` (would still find `/a/b/.drone.yml`).
* `PLUGIN_DEBUG`: Set this to `true` to enable debug messages.
* `PLUGIN_ADDRESS`: Listen address for the plugins webserver. Defaults to `:3000`.
* `PLUGIN_SECRET`: Shared secret with drone. You can generate the token using `openssl rand -hex 16`.
* `PLUGIN_ALLOW_LIST_FILE`: (Optional) Path to regex pattern file. Matches the repo slug(s) against a list of regex patterns. Defaults to `""`, match everything.
* `PLUGIN_CONSIDER_FILE`: (Optional) Consider file name. Only consider the `.drone.yml` files listed in this file. When defined, all enabled repos must contain a consider file.

Backend specific options

* `SERVER`: Custom SCM server (also used by Gitlab / Bitbucket)
* GitHub:
  * `GITHUB_TOKEN`: Github personal access token. Only needs repo rights. See [here][1].
* GitLab:
  * `GITLAB_TOKEN`: Gitlab personal access token. Only needs `read_repository` rights. See [here][2]
* Bitbucket
  * `BITBUCKET_AUTH_SERVER`: Custom auth server (uses SERVER if empty)
  * `BITBUCKET_CLIENT`: Credentials for Bitbucket access
  * `BITBUCKET_SECRET`: Credentials for Bitbucket access

If `PLUGIN_CONCAT` is not set, the first found `.drone.yml` will be used.

#### Example docker-compose

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
    restart: always
```

Edit the Secrets (`***`), `<SECRET>` and `<GITHUB_TOKEN>` to your needs. `<SECRET>` is used between Drone and drone-tree-config.

#### Enable repos via regex matching

By default, this plugin matches against ALL repo slugs. If you want to enable the plugin for specific repos only, turn on
regex matching by specifying a `PLUGIN_ALLOW_LIST_FILE`.

* Regex match rules must comply with [re2][3] syntax.
* Each line is a single rule.
* Empty lines are ignored.
* Lines which start with `#` are treated as comments (ignored).

Updated docker-compose:

```yaml
  drone-tree-config:
    image: bitsbeats/drone-tree-config
    environment:
      - PLUGIN_DEBUG=true
      - PLUGIN_CONCAT=true
      - PLUGIN_FALLBACK=true
      - PLUGIN_SECRET=<SECRET>
      - GITHUB_TOKEN=<GITHUB_TOKEN>
      - PLUGIN_ALLOW_LIST_FILE=/drone-tree-config-matchfile
    restart: always
    volumes:
      - /var/lib/drone/drone-tree-config-matchfile:/drone-tree-config-matchfile
```

File: drone-tree-config-matchfile:

```text
^bitbeats/.*$
^myorg/myrepo$
```

* Matches against all repos in the `bitbeats` org
* Matches against `myorg/myrepo`

[1]: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line
[2]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html
[3]: https://github.com/google/re2/wiki/Syntax

#### Consider file

 If a `PLUGIN_CONSIDER_FILE` is defined, drone-tree-config will first read the content of the target file and will only consider
 the `.drone.yml` files specified, when matching.

Depending on the size and the complexity of the repository, using a "consider file" can
significantly reduce the number of API calls made to the provider (github, bitbucket, other). The reduction in API calls
reduces the risk of being rate limited and can result in less processing time for drone-tree-config.

Given the config;

```yaml
   - PLUGIN_CONSIDER_FILE=.drone-consider
```

A local git repo clone;

```shell
$ tree -a my-repo-clone/
 my-repo-clone/
 ├── .drone-consier
 ├── foo
 │   └── .drone.yml
 ├── bar
 │   └── .drone.yml
 └── baz

```

Content of the .drone-consider to check in;

```shell
$ cat my-repo-clone/.drone-consider
foo/.drone.yml
bar/.drone.yml
```

The downside of a "consider file" is that it has to be kept in sync. As a suggestion, to help with this, a step can be
added to each `.drone.yml` which verifies the "consider file" is in sync with the actual content of the repo. For
example, this can be accomplished by comparing the output of `find ./ -name .drone.yml` with the content of the "consider file".
