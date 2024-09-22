# URDT-USSD SSH server

An SSH server entry point for the vise engine.


## Adding public keys for access

Map your (client) public key to a session identifier (e.g. phone number)

```
go run -v -tags logtrace ./cmd/ssh/sshkey/main.go -i <session_id> [--dbdir <dbpath>] <publickey_filepath>
```


## Create a private key for the server

```
ssh-keygen -N "" -f <privatekey_filepath>
```


## Run the server


```
go run -v -tags logtrace ./cmd/ssh/main.go -h <host> -p <port> [--dbdir <dbpath>] <privatekey_filepath>
```


## Connect to the server

```
ssh -T -p <port> <host>
```
