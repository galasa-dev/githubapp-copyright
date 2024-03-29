# Github copyright checker

Objective: Checks that valid copyright statements are present in code, checking when a pull request is created/re-opened/updated.

[![build of main branch](https://github.com/galasa-dev/githubapp-copyright/actions/workflows/build.yaml/badge.svg)](https://github.com/galasa-dev/githubapp-copyright/actions/workflows/build.yaml)

# Checks performed

There are two main classes of things checked:

For `.java`, `.go`, `.ts`, `.tsx`, and `.js` files, we expect this:
```
/* 
 * Copyright contributors to the Galasa project
 *
 * SPDX-License-Identifier: EPL-2.0
 */
```

For `.yaml` and `.sh` files, we expect this:
```
#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#
```

# Developing and Deploying

This code builds a docker image, which can be deployed to kubernetes.

## Running the code

The program `copyright` or `copyright-amd64` is invoked with this syntax:

```
copyright --githubAuthKeyFile <key-file-path> [--debug]
```

Parameters:

key-file-path is a mandatory parameter. It holds the path to a file which is a key.pem, in which we hold a 
certificate which can be used to log into github as the `galasa` user.

For example `copyright --githubAuthKeyFile /my/folder/key.pem`

That lets the copyright application authenticate with github, so it can do things like ask for the file content
of a file which is mentioned in a pull request.

--debug : An optional flag. If used, then HTTP traffic is logged in the log output. Useful for capturing real packets for unit tests.

## Deploying

The key.pem file should be supplied to any deployment as a secret.

## Running the docker image
- Create a key.pem file in temp
  - The contents of this file can be lifted from the `pkg/checks/tokenSupplierMock.go` file
```
docker run -p 3000:3000 -v $(pwd)/temp:/temp githubapp-copyright:latest copyright --debug --githubAuthKeyFile /temp/key.pem
```

Then you could hit the tool with curl:
```
curl -X POST http://localhost:3000/githubapp/copyright/event_handler -H "Content-Type: application/json" -d '{"key1":"value1", "key2":"value2"}'
```

## License
See [license file](./LICENSE)

## Contributing
See [contributing guidelines](./CONTRIBUTIONS.md)

