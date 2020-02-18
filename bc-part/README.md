# HLF Iot with Hyperledger Fabric 1.4.2, local deployment version

Requirements:
- Docker
- Make
- Python3.6+
- Mako template engine
- pyyaml

```
pip3 install Mako pyyaml
```

Generate artifacts with crypto material, configs and dockercompose templates.
Build API container and web client.
```
make generate
```

Bring up local development network:
```
make up
```

Remove docker containers and volumes:
```
make clean
```


## Members and Components

Network consortium consists of:

- Orderer organization `hlfiot`
- Peer organization `hlfiot` 
- Peer organization `device` 

They transact with each other on the following channel:
- `common` involving all members and with chaincode `hlf_iot_cc` deployed

Each organization starts several docker containers:

- **peer0** (ex.: `peer0.a.hlfiot`) with the anchor [peer](https://github.com/hyperledger/fabric/tree/release/peer) runtime
- **api** `api.a.hlfiot` with [Fabric Rest API Go](https://gitlab.altoros.com/intprojects/fabric-rest-api-go) API server
- **www** `www.a.hlfiot` nginx server to serve web client and reverse proxy to API
- **cli** `cli.a.hlfiot` with tools to run commands during setup

## Local deployment

Deploy docker containers of all member organizations to one host, for development and testing of functionality. 

After all containers are up, web interfaces will be at:

- hlfiot [http://localhost:3001](http://localhost:3001/)
- device [http://localhost:3002](http://localhost:3002/)

## Testing

...