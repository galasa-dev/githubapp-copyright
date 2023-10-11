# Github copyright checker

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

--debug : An optoinal flag. If used, then HTTP traffic is logged in the log output. Useful for capturing real packets for unit tests.

## Deploying

The key.pem file should be supplied to any deployment as a secret.

