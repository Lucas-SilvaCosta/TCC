## TCC
Arquivos de configuração e chaincodes para reproduzir do ensaio prático

## Setup

Para o funcionamento correto dos comandos a seguir é ideal executar o clone deste repositório e do repositório do projeto [Hyperledger Cacti](https://github.com/hyperledger-cacti/cacti) na home do seu computador.

Caso tenha problemas durante a execução utilize o [guia original](https://hyperledger-cacti.github.io/cacti/weaver/getting-started/test-network/setup-local-docker/) disponibilizado no repositório.

## Passo-a-passo

### Pre-requisitos
- curl
- Git (latest)
- Docker (lastest)
- Docker-compose (version 2 or higher)
- Golang (version 1.20 or higher)
- Node.js and NPM (version 16 supported)
- Yarn
- Protoc
    ```
    apt-get install protobuf-compiler
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
    ```

### Protos
Compilar protobufs:
```
cd ~/cacti/weaver/common/protos-js
make build

cd ~/cacti/weaver/common/protos-go
make build

cd ~/cacti/weaver/common/protos-rs
make build
```

### Fabric Interoperation SDK
Compila o sdk:
```
cd ~/cacti/weaver/sdks/fabric/interoperation-node-sdk
make build-local
```

### Fabric Network
Copia chaincodes desenvolvidas para a pasta adequada do repositório:
```
cd ~/cacti/weaver/samples/fabric
cp -r ~/TCC/chaincodes/* ./
```

Inicia as redes da educação e da saúde com suas respectivas chaincodes:
```
cd ~/cacti/weaver/tests/network-setups/fabric/dev
make start-interop-network1-local CHAINCODE_NAME=crm
make start-interop-network2-local CHAINCODE_NAME=prescription
```

### Relay
Builda e inicia os relays de ambas as redes, copiando as configurações definidas:
```
cd ~/cacti/weaver/core/relay
cp ~/TCC/relay/* ./docker/testnet-envs
make build-server-local
make convert-compose-method2
make start-server COMPOSE_ARG='--env-file docker/testnet-envs/.env.n1'
make start-server COMPOSE_ARG='--env-file docker/testnet-envs/.env.n2'
make convert-compose-method1
```

### Driver
Builda e inicia os drivers de ambas as redes, copiando as configurações definidas:
```
cd ~/cacti/weaver/core/drivers/fabric-driver
cp ~/TCC/driver/* ./docker-testnet-envs
make build-image-local
make deploy COMPOSE_ARG='--env-file docker-testnet-envs/.env.n1' NETWORK_NAME=$(grep NETWORK_NAME docker-testnet-envs/.env.n1 | cut -d '=' -f 2)
make deploy COMPOSE_ARG='--env-file docker-testnet-envs/.env.n2' NETWORK_NAME=$(grep NETWORK_NAME docker-testnet-envs/.env.n2 | cut -d '=' -f 2)
```

### IIN Agent
Builda e inicia os agentes IIN de ambas as redes, copiando as configurações definidas:
```
cd ~/cacti/weaver/core/identity-management/iin-agent
cp ~/TCC/IIN/* ./docker-testnet/envs
make build-image-local
make deploy COMPOSE_ARG='--env-file docker-testnet/envs/.env.n1.org1' DLT_SPECIFIC_DIR=$(grep DLT_SPECIFIC_DIR docker-testnet/envs/.env.n1.org1 | cut -d '=' -f 2)
make deploy COMPOSE_ARG='--env-file docker-testnet/envs/.env.n2.org1' DLT_SPECIFIC_DIR=$(grep DLT_SPECIFIC_DIR docker-testnet/envs/.env.n2.org1 | cut -d '=' -f 2)
```

### Fabric-cli
Atualiza e monta o fabric-cli, copiando as configurações definidas:
```
cd ~/cacti/weaver/samples/fabric/fabric-cli
cp ~/TCC/fabric-cli/all.ts ./src/commands/configure
cp -r ~/TCC/fabric-cli/crm ./src/data/interop/
make build-local
```

Prepara CLI para execução de funções, passando configurações necessárias:
```
cp ~/TCC/fabric-cli/config.json ./
cp ~/TCC/fabric-cli/chaincode.json ./
cp ~/TCC/fabric-cli/.env ./
./bin/fabric-cli env set-file ./.env

./bin/fabric-cli configure all network1 network2
```
Após executar este comando, aguarde no mínimo 5 minutos para que os agentes IIN de ambas as redes se atualizem automaticamente.

### Populando a rede
Populando rede da educação:
```
./bin/fabric-cli chaincode invoke mychannel crm create '["123","Dermatology",true]' --local-network=network1
./bin/fabric-cli chaincode invoke mychannel crm create '["456","Cardiology",false]' --local-network=network1
```

Populando rede da saúde:
```
./bin/fabric-cli chaincode invoke mychannel prescription create '["1","123","Dermatology","NorthbridgeMedicalCenter","Isotretinoin"]' --local-network=network2
./bin/fabric-cli chaincode invoke mychannel prescription create '["2","123","Endocrinology","NorthbridgeMedicalCenter","Metformin"]' --local-network=network2
./bin/fabric-cli chaincode invoke mychannel prescription create '["3","456","Cardiology","CrestwoodGeneralHospital","Lisinopril"]' --local-network=network2
```

### Fluxo de interoperabilidade 1
Um médico com CRM válido, emite uma receita dentro do escopo de sua especialidade.

Antes de executar o comando abra, o arquivo ```~/cacti/weaver/samples/fabric/fabric-cli/.env``` e troque a variável ```DEFAULT_APPLICATION_CHAINCODE``` para ```prescription```. Então execute:
```
./bin/fabric-cli interop --local-network=network2 --requesting-org=Org1MSP relay-network1:9080/network1/mychannel:crm:Check:123
./bin/fabric-cli chaincode query mychannel prescription read '["1"]' --local-network=network2
```

### Fluxo de interoperabilidade 2
Um médico com CRM válido, emite uma receita fora do escopo de sua especialidade.

Antes de executar o comando, abra o arquivo ```~/cacti/weaver/samples/fabric/fabric-cli/chaincode.json``` e troque a o primeiro valor do array em ```prescription:Check:args``` para "2". Então execute:
```
./bin/fabric-cli interop --local-network=network2 --requesting-org=Org1MSP relay-network1:9080/network1/mychannel:crm:Check:123
./bin/fabric-cli chaincode query mychannel prescription read '["2"]' --local-network=network2
```

### Fluxo de interopetabilidade 3
Um médico com CRM inválido, emite uma receita.

Antes de executar o comando, abra o arquivo ```~/cacti/weaver/samples/fabric/fabric-cli/chaincode.json``` e troque a o primeiro valor do array em ```prescription:Check:args``` para "3". Então execute:
```
./bin/fabric-cli interop --local-network=network2 --requesting-org=Org1MSP relay-network1:9080/network1/mychannel:crm:Check:456
./bin/fabric-cli chaincode query mychannel prescription read '["3"]' --local-network=network2
```