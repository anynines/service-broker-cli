# Service Broker CLI
Service Broker CLI is a command-line tool to interact with a Cloud Foundry like Service Broker.
It was written to reduce to time to test changes on a Service Broker in the development phase.

## Table of Contents
This documentation contains the following topics.

- [Table of Contents](#table-of-contents)
- [Download, Build and Install](#download-build-and-install)
- [Publish a Release](#publish-a-release)
- [Usage](#usage)
- [Restrictions](#restrictions)
- [Target and login](#target-and-login)

## Download, Build and Install
Service Broker CLI is written in [`Go`](https://golang.org). It is designed to use only standard libraries.

It provides also a Makefile, so it's quite easy to build and install.

Downlad the repo with `go get github.com/phartz/service-broker-cli`, then change the current path to the sources `cd $GOPATH/src/github.com/phartz/service-broker-cli`.

Build
```
make
```

Install
```
make install
```

Either add the `$GOPATH/bin` `export PATH=$PATH:$GOPATH/bin` to your `$PATH` or copy the cli to your bin folder `cp sb /usr/local/bin`

## Publish a Release

Releases are published automatically by GitHub Actions when a new Git tag is pushed.
The workflow builds binaries for Linux, Windows and macOS (Intel and Apple Silicon),
creates a GitHub Release, generates release notes and uploads the binaries as assets.

### Steps

1. Make sure your changes are merged and pushed.
2. Create a new tag (example):

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
```

3. Push the tag:

```bash
git push origin v1.2.3
```

4. Open the repository in GitHub and wait for the `Release` workflow to finish.
5. Open the new GitHub Release to verify attached binaries.

### Produced binaries

The workflow uploads these target binaries to the release:

- `sb_<tag>_linux_amd64`
- `sb_<tag>_windows_amd64.exe`
- `sb_<tag>_darwin_amd64`
- `sb_<tag>_darwin_arm64`

## Usage
```
$sb help

NAME:
   sb - A command line tool to interact with a service broker

USAGE:
   sb [global options] command [arguments...] [command options]

COMMANDS

    COMMAND               SHORTCUT    DESCRIPTION
    help                  h           Show help

    api                               Set or view target api url
    login                 l           Login to the target
    logout                lo          Logout from the target
    auth                              Authenticate to the target
    version               -v          Print the version

    marketplace           m           List available offerings in the marketplace
    services              s           List all service instances in the target space
    service                           Show service instance info
    find-service          fs          Lookup a service with a given deployment name

    create-service        cs          Create a service instance
    update-service                    Update a service instance
    delete-service        ds          Delete a service instance

    create-service-key    csk         Create key for a service instance
    service-keys          sk          List keys for a service instance
    delete-service-key    dsk         Delete a service key

```

To get specific help use `sb help <command>`


## Restrictions

There are some restrictions:
* The Service Broker doesn't store any user credentials so it is not possible to get information about a service key.
In contrast to to Cloud Foundry CLI the `sb create-service-key` returns the credentials.
* The Service Broker also doesn't store any information about the real service names. The broker only works with UUIDs.


## Target and login

Use `sb target` to target a Service Broker.

```bash
$ sb target https://localhost:3001
```

SB CLI supports TLS encrypted communication.
To login you can use either `sb login` as an interactive operation or `sb auth` for scripting.

`sb login` also supports the flags `-a` for API, `-u` for the username and `-p` for the password.
So you can use

```bash
$sb login -a https://redis-broker.service.dc1.consul.dsf2:3001 -u admin -p <password>
```

The credentials will be stored in JSON format in the file `.sb`. Those files can be located either in the current working directory or in its parent directory. If the file was not found, sb is looking also in the users root path.

```
--> /some/where/on/your/disk
--> /some/where/on/your
--> /some/where/on
--> /some/where
--> /some
--> /
--> ~
```

### Alternative Way to Use the CLI

The Service-Broker-CLI is able to read the credentials and the host from environment variables. No login is then needed.

```
$ SB_HOST="http://redis-service-broker.service.dc1.consul:3000/" SB_USERNAME="admin" SB_PASSWORD="mytopsecretpwd" sb ......
```

| Env Var | Description |
|---|---|
| SB_HOST | The host name, it can be the IP, the FQHN, with or without the scheme and port |
| SB_USERNAME | The username |
| SB_PASSWORD | The password |
| SB_SKIP_SSL_VERIFY | If true SSL verification is skipped |
| SB_TIMEOUT | HTTP Timeout in seconds (default: 15) |
| SB_ORGANIZATION_GUID | Overrides the `organization_guid` sent on provision, update, and service-key requests (both as a top-level field and inside `context`). When unset, `create-service` generates a random UUID and `update-service` / `create-service-key` reuse the value the broker stored on the instance. |
| SB_SPACE_GUID | Overrides the `space_guid` sent on provision, update, and service-key requests. Same fallback semantics as `SB_ORGANIZATION_GUID`. |


## Logging

Service-Broker-CLI does not provide a logging in the classical sense. But you can trace the Service Broker API requests and answers.

```
$ SB_TRACE=ON && sb services
```

