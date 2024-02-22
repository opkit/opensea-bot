package pkg

import (
	"encoding/json"
	wallet "github.com/ethersphere/bee/pkg/crypto"
	"math/big"
	"opensea-bot/pkg/seaport"
	"os"
)

const apiDomain = "https://api.opensea.io"
const testnetApiDomain = "https://testnets-api.opensea.io"

const NftType721 = "erc721"
const NftType1155 = "erc1155"

var rpcURL = map[string]string{
	"sepolia":  "https://sepolia.infura.io/v3/",
	"ethereum": "https://mainnet.infura.io/v3/",
}

var types = `
{
	"EIP712Domain": [{
			"name": "name",
			"type": "string"
		},
		{
			"name": "version",
			"type": "string"
		},
		{
			"name": "chainId",
			"type": "uint256"
		},
		{
			"name": "verifyingContract",
			"type": "address"
		}
	],
	"OrderComponents": [{
			"name": "offerer",
			"type": "address"
		},
		{
			"name": "zone",
			"type": "address"
		},
		{
			"name": "offer",
			"type": "OfferItem[]"
		},
		{
			"name": "consideration",
			"type": "ConsiderationItem[]"
		},
		{
			"name": "orderType",
			"type": "uint8"
		},
		{
			"name": "startTime",
			"type": "uint256"
		},
		{
			"name": "endTime",
			"type": "uint256"
		},
		{
			"name": "zoneHash",
			"type": "bytes32"
		},
		{
			"name": "salt",
			"type": "uint256"
		},
		{
			"name": "conduitKey",
			"type": "bytes32"
		},
		{
			"name": "counter",
			"type": "uint256"
		}
	],
	"OfferItem": [{
			"name": "itemType",
			"type": "uint8"
		},
		{
			"name": "token",
			"type": "address"
		},
		{
			"name": "identifierOrCriteria",
			"type": "uint256"
		},
		{
			"name": "startAmount",
			"type": "uint256"
		},
		{
			"name": "endAmount",
			"type": "uint256"
		}
	],
	"ConsiderationItem": [{
			"name": "itemType",
			"type": "uint8"
		},
		{
			"name": "token",
			"type": "address"
		},
		{
			"name": "identifierOrCriteria",
			"type": "uint256"
		},
		{
			"name": "startAmount",
			"type": "uint256"
		},
		{
			"name": "endAmount",
			"type": "uint256"
		},
		{
			"name": "recipient",
			"type": "address"
		}
	]
}
`

type contractInfo struct {
	Address          string `json:"address"`
	Chain            string `json:"chain"`
	Collection       string `json:"collection"`
	ContractStandard string `json:"contract_standard"`
	Name             string `json:"name"`
}

type OrderParameters struct {
	Offerer                         string              `json:"offerer"`
	Zone                            string              `json:"zone"`
	ZoneHash                        string              `json:"zoneHash"`
	StartTime                       int64               `json:"startTime"`
	EndTime                         int64               `json:"endTime"`
	OrderType                       uint8               `json:"orderType"`
	Salt                            string              `json:"salt"`
	ConduitKey                      string              `json:"conduitKey"`
	Offer                           []OfferItem         `json:"offer"`
	Consideration                   []ConsiderationItem `json:"consideration"`
	TotalOriginalConsiderationItems int                 `json:"totalOriginalConsiderationItems"`
	Counter                         int64               `json:"counter"`
}

type OfferItem struct {
	ItemType             uint8  `json:"itemType"`
	Token                string `json:"token"`
	IdentifierOrCriteria int64  `json:"identifierOrCriteria"`
	StartAmount          int64  `json:"startAmount"`
	EndAmount            int64  `json:"endAmount"`
}

type ConsiderationItem struct {
	ItemType             uint8  `json:"itemType"`
	Token                string `json:"token"`
	IdentifierOrCriteria int64  `json:"identifierOrCriteria"`
	StartAmount          int64  `json:"startAmount"`
	EndAmount            int64  `json:"endAmount"`
	Recipient            string `json:"recipient"`
}

type protocolData struct {
	Parameters      OrderParameters `json:"parameters"`
	Signature       string          `json:"signature"`
	ProtocolAddress string          `json:"protocol_address"`
}

type NFT struct {
	Identifier    string `json:"identifier"`
	Contract      string `json:"contract"`
	TokenStandard string `json:"token_standard"`
}
type AccountNFTsResp struct {
	Nfts []NFT `json:"nfts"`
}

type Account struct {
	signer          wallet.Signer
	contract        *contractInfo
	seaportInstance *seaport.Seaport
	chainID         *big.Int
}

type paymentTokenResp struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Chain    string `json:"chain"`
	Image    string `json:"image"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
	EthPrice string `json:"eth_price"`
	UsdPrice string `json:"usd_price"`
}
type CollectionResp struct {
	Collection              string `json:"collection"`
	TraitOffersEnabled      bool   `json:"trait_offers_enabled"`
	CollectionOffersEnabled bool   `json:"collection_offers_enabled"`
	Contracts               []struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
	} `json:"contracts"`
	Fees []struct {
		Fee       float64 `json:"fee"`
		Recipient string  `json:"recipient"`
		Required  bool    `json:"required"`
	} `json:"fees"`
}

type BestListingListResp struct {
	Listings []BestListingResp `json:"listings"`
}
type BestListingResp struct {
	OrderHash string `json:"order_hash"`
	Chain     string `json:"chain"`
	Type      string `json:"type"`
	Price     struct {
		Current struct {
			Currency string `json:"currency"`
			Decimals int    `json:"decimals"`
			Value    string `json:"value"`
		} `json:"current"`
	} `json:"price"`
	ProtocolData struct {
		Parameters Parameters `json:"parameters"`
	} `json:"protocol_data"`
	ProtocolAddress string `json:"protocol_address"`
}
type Parameters struct {
	Offerer string `json:"offerer"`
	Offer   []struct {
		ItemType             int    `json:"itemType"`
		Token                string `json:"token"`
		IdentifierOrCriteria string `json:"identifierOrCriteria"`
		StartAmount          string `json:"startAmount"`
		EndAmount            string `json:"endAmount"`
	} `json:"offer"`
	Consideration []struct {
		ItemType             int    `json:"itemType"`
		Token                string `json:"token"`
		IdentifierOrCriteria string `json:"identifierOrCriteria"`
		StartAmount          string `json:"startAmount"`
		EndAmount            string `json:"endAmount"`
		Recipient            string `json:"recipient"`
	} `json:"consideration"`
	StartTime                       string `json:"startTime"`
	EndTime                         string `json:"endTime"`
	OrderType                       int    `json:"orderType"`
	Zone                            string `json:"zone"`
	ZoneHash                        string `json:"zoneHash"`
	Salt                            string `json:"salt"`
	ConduitKey                      string `json:"conduitKey"`
	TotalOriginalConsiderationItems int    `json:"totalOriginalConsiderationItems"`
	Counter                         int    `json:"counter"`
}
type CreateListingResp struct {
	Order struct {
		CreatedDate    string `json:"created_date"`
		ClosingDate    string `json:"closing_date"`
		ListingTime    int    `json:"listing_time"`
		ExpirationTime int    `json:"expiration_time"`
		OrderHash      string `json:"order_hash"`
		ProtocolData   struct {
			Parameters Parameters `json:"parameters"`
			Signature  string     `json:"signature"`
		} `json:"protocol_data"`
		ProtocolAddress string `json:"protocol_address"`
		CurrentPrice    string `json:"current_price"`
	} `json:"order"`
}

type AssetEvents struct {
	EventType      string  `json:"event_type"`
	Chain          string  `json:"chain"`
	Nft            NFT     `json:"nft"`
	Quantity       int     `json:"quantity"`
	Seller         string  `json:"seller"`
	Buyer          string  `json:"buyer"`
	Payment        payment `json:"payment"`
	Transaction    string  `json:"transaction"`
	EventTimestamp int     `json:"event_timestamp"`
}
type SaleResp struct {
	AssetEvents []AssetEvents `json:"asset_events"`
}

type payment struct {
	Quantity     string `json:"quantity"`
	TokenAddress string `json:"token_address"`
	Decimals     int    `json:"decimals"`
	Symbol       string `json:"symbol"`
}

func (v *CreateListingResp) String() string {
	s, _ := json.Marshal(v)
	return string(s)
}
func (n *NFT) nftType() uint8 {
	switch n.TokenStandard {
	case NftType1155:
		return 3
	case NftType721:
		return 2
	}
	return 0
}

func (v *AccountNFTsResp) Get(identifier string) *NFT {
	for _, nft := range v.Nfts {
		if nft.Identifier == identifier {
			return &nft
		}
	}
	return nil
}
func getOpenSeaAPI(chain string) string {
	if chain == "ethereum" {
		return apiDomain
	}
	return testnetApiDomain
}

func getRpcURL(chain string) string {
	return rpcURL[chain] + os.Getenv("INFURA_KEY")
}
