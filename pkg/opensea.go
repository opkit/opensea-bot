package pkg

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	wallet "github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/crypto/eip712"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"opensea-bot/pkg/seaport"
	"os"
	"strconv"
	"strings"
	"time"
)

const SeaportConduitKey = "0x0000007b02230091a7ed01230072f7006a004d60a8d4e71d599b8104250f0000"

const ProtocolAddress = "0x00000000000000adc04c56bf30ac9d3c0aaf14dc"

var request = gorequest.New()

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	request.Header.Add("x-api-key", os.Getenv("OPENSEA_API_KEY"))
}

func NewAccount(ctx context.Context, contractAddress, chain string) *Account {
	privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	signer := wallet.NewDefaultSigner(privateKey)
	walletAddress, err := signer.EthereumAddress()
	if err != nil {
		log.Fatal("error casting public key to ECDSA", err)
	}

	var info *contractInfo
	req := request.Clone().Get(fmt.Sprintf("%s/api/v2/chain/%s/contract/%s", getOpenSeaAPI(chain), chain, contractAddress))

	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&info)
	if len(errs) > 0 {
		panic(errs[0])
	}
	log.Println(resp)
	log.Println("wallet address: ", walletAddress.Hex())

	cli, err := ethclient.Dial(getRpcURL(chain))
	if err != nil {
		panic(err)
	}

	chainID, err := cli.ChainID(ctx)
	if err != nil {
		panic(err)
	}

	seaportInstance, err := seaport.NewSeaport(common.HexToAddress(ProtocolAddress), cli)
	if err != nil {
		panic(err)
	}

	return &Account{
		signer:          signer,
		contract:        info,
		seaportInstance: seaportInstance,
		chainID:         chainID,
	}
}

func (a *Account) GetCollection(ctx context.Context) (*CollectionResp, error) {
	var data *CollectionResp
	req := request.Clone().Get(fmt.Sprintf("%s/api/v2/collections/%s", getOpenSeaAPI(a.contract.Chain), a.contract.Collection))
	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&data)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	log.Println(resp)
	return data, nil
}

func (a *Account) GetNFTs(ctx context.Context) (*AccountNFTsResp, error) {
	var data *AccountNFTsResp
	req := request.Clone().
		Get(fmt.Sprintf("%s/api/v2/chain/%s/account/%s/nfts", getOpenSeaAPI(a.contract.Chain), a.contract.Chain, a.WalletAddress().Hex())).
		Param("collection", a.contract.Collection)
	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&data)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	log.Println(resp)
	return data, nil
}

func (a *Account) GetBestListingByNFT(ctx context.Context, identifier string) (*BestListingResp, error) {
	var data *BestListingResp
	req := request.Clone().
		Get(fmt.Sprintf("%s/api/v2/listings/collection/%s/nfts/%s/best", getOpenSeaAPI(a.contract.Chain), a.contract.Collection, identifier))
	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&data)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	log.Println(resp)
	return data, nil
}

func (a *Account) CreateListing(ctx context.Context, nft *NFT, price string) error {
	startTime := big.NewInt(time.Now().Local().Unix())
	endTime := big.NewInt(time.Now().Local().Add(10 * time.Minute).Unix())

	paymentToken, err := a.contract.paymentToken(ctx)
	if err != nil {
		return err
	}

	listPrice, err := decimal.NewFromString(price)
	if err != nil {
		return err
	}

	if listPrice.IsZero() {
		return errors.New("price is zero")
	}
	listPrice = listPrice.Shift(int32(paymentToken.Decimals))

	collection, err := a.GetCollection(ctx)
	if err != nil {
		return err
	}

	identifierOrCriteria, _ := strconv.Atoi(nft.Identifier)
	offer := OfferItem{
		ItemType:             nft.nftType(),
		Token:                common.HexToAddress(nft.Contract).Hex(),
		StartAmount:          1,
		EndAmount:            1,
		IdentifierOrCriteria: int64(identifierOrCriteria),
	}

	counter, err := a.seaportInstance.GetCounter(nil, a.WalletAddress())
	if err != nil {
		return err
	}

	considerations := make([]ConsiderationItem, 0)
	var totalFee = decimal.Zero
	for _, fee := range collection.Fees {
		if fee.Required {
			feeAmount := listPrice.Mul(decimal.NewFromFloat(fee.Fee)).Div(decimal.NewFromInt(100))
			totalFee = totalFee.Add(feeAmount)
			considerations = append(considerations, ConsiderationItem{
				ItemType:             0,
				Token:                paymentToken.Address,
				IdentifierOrCriteria: 0,
				StartAmount:          feeAmount.BigInt().Int64(),
				EndAmount:            feeAmount.BigInt().Int64(),
				Recipient:            fee.Recipient,
			})
		}
	}
	considerations = append([]ConsiderationItem{
		{
			ItemType:             0,
			Token:                paymentToken.Address,
			IdentifierOrCriteria: 0,
			StartAmount:          listPrice.Sub(totalFee).BigInt().Int64(),
			EndAmount:            listPrice.Sub(totalFee).BigInt().Int64(),
			Recipient:            a.WalletAddress().Hex(),
		},
	}, considerations...)
	param := OrderParameters{
		Offerer:                         a.WalletAddress().Hex(),
		Zone:                            zeroAddress().Hex(),
		ZoneHash:                        zero32BytesHexString(),
		StartTime:                       startTime.Int64(),
		EndTime:                         endTime.Int64(),
		OrderType:                       0, // FULL_OPEN
		Salt:                            fixedSalt(),
		ConduitKey:                      SeaportConduitKey,
		Offer:                           []OfferItem{offer},
		Consideration:                   considerations,
		TotalOriginalConsiderationItems: len(considerations),
		Counter:                         counter.Int64(),
	}

	data, err := param.signTypedData(a)
	if err != nil {
		return err
	}

	req := request.Clone().
		Post(fmt.Sprintf("%s/api/v2/orders/%s/seaport/listings", getOpenSeaAPI(a.contract.Chain), a.contract.Chain)).Send(data)

	log.Println(req.AsCurlCommand())

	var output *CreateListingResp
	_, _, errs := req.EndStruct(&output)
	if len(errs) > 0 {
		return errs[0]
	}

	log.Println(output)
	return nil
}

func (c *contractInfo) paymentToken(ctx context.Context) (*paymentTokenResp, error) {
	var data *paymentTokenResp
	req := request.Clone().
		Get(fmt.Sprintf("%s/api/v2/chain/%s/payment_token/0x0000000000000000000000000000000000000000", getOpenSeaAPI(c.Chain), c.Chain))
	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&data)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	log.Println(resp)
	return data, nil
}

func (c *contractInfo) Stats() {

}

func (a *Account) WalletAddress() common.Address {
	address, _ := a.signer.EthereumAddress()
	return address
}

func (p *OrderParameters) signTypedData(account *Account) (*protocolData, error) {
	name, _ := account.seaportInstance.Name(nil)
	info, _ := account.seaportInstance.Information(nil)
	var data = &eip712.TypedData{
		PrimaryType: "OrderComponents",
		Domain: apitypes.TypedDataDomain{
			Name:              name,
			Version:           info.Version,
			ChainId:           math.NewHexOrDecimal256(account.chainID.Int64()),
			VerifyingContract: ProtocolAddress,
		},
		Message: map[string]interface{}{},
	}
	_ = json.Unmarshal([]byte(types), &data.Types)
	salt, _ := big.NewInt(0).SetString(p.Salt, 0)

	offer := make([]interface{}, 0)
	for _, item := range p.Offer {
		offer = append(offer, map[string]interface{}{
			"itemType":             big.NewInt(int64(item.ItemType)),
			"token":                item.Token,
			"identifierOrCriteria": big.NewInt(item.IdentifierOrCriteria),
			"startAmount":          big.NewInt(item.StartAmount),
			"endAmount":            big.NewInt(item.EndAmount),
		})
	}

	data.Message["offer"] = offer

	consideration := make([]interface{}, 0)
	for _, item := range p.Consideration {
		consideration = append(consideration, map[string]interface{}{
			"itemType":             big.NewInt(int64(item.ItemType)),
			"token":                item.Token,
			"identifierOrCriteria": big.NewInt(item.IdentifierOrCriteria),
			"startAmount":          big.NewInt(item.StartAmount),
			"endAmount":            big.NewInt(item.EndAmount),
			"recipient":            item.Recipient,
		})
	}

	data.Message["consideration"] = consideration
	data.Message["offerer"] = p.Offerer
	data.Message["startTime"] = big.NewInt(p.StartTime)
	data.Message["endTime"] = big.NewInt(p.EndTime)
	data.Message["orderType"] = big.NewInt(int64(p.OrderType))
	data.Message["zone"] = p.Zone
	data.Message["zoneHash"] = hexStringToByte32(p.ZoneHash)
	data.Message["salt"] = salt
	data.Message["conduitKey"] = hexStringToByte32(p.ConduitKey)
	data.Message["counter"] = big.NewInt(p.Counter)

	str, _ := json.Marshal(data)

	log.Println("eip712 data", string(str))

	sign, err := account.signer.SignTypedData(data)
	if err != nil {
		return nil, err
	}

	return &protocolData{
		Parameters:      *p,
		Signature:       hexutil.Encode(sign),
		ProtocolAddress: ProtocolAddress,
	}, nil
}

func (c *contractInfo) lastSaleCost(nft *NFT) (*payment, error) {
	var data *SaleResp
	req := request.Clone().Get(
		fmt.Sprintf("%s/api/v2/events/chain/%s/contract/%s/nfts/%s?event_type=sale&limit=1",
			getOpenSeaAPI(c.Chain), c.Chain, nft.Contract, nft.Identifier))
	log.Println(req.AsCurlCommand())
	resp, _, errs := req.EndStruct(&data)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	log.Println(resp)
	if len(data.AssetEvents) > 0 {
		return &data.AssetEvents[0].Payment, nil
	}
	return nil, nil
}

func hexStringToByte32(hexString string) [32]byte {
	if hexString == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		return [32]byte{0}
	}
	decodedByteArray, err := hex.DecodeString(strings.TrimPrefix(hexString, "0x"))
	if err != nil {
		panic("Unable to convert hex to byte.")
	}
	var arr [32]byte
	copy(arr[:], decodedByteArray)
	return arr
}

func fixedSalt() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func zeroAddress() common.Address {
	addrBytes := [20]byte{0}
	addr := common.BytesToAddress(addrBytes[:])
	return addr
}

func zero32BytesHexString() string {
	return "0x0000000000000000000000000000000000000000000000000000000000000000"
}
