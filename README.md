## opensea bot


```shell
# opensea api key
export OPENSEA_API_KEY=apikey
# infura rpc key
export INFURA_KEY=key
# waller private key
export PRIVATE_KEY=0x0000000000000
```


### idea
Automatically place orders to sell NFT held by the wallet according to the input configuration

The interface has been implemented in pkg/opensea.go

- Query wallet holding NFTs
- Query NFT pending order sales information
- Query NFT contract information
- EIP712 signature order process


### next

Improve main.go, you can automatically buy and sell NFT according to the configuration through the cli method, and implement the opensea trading bot
