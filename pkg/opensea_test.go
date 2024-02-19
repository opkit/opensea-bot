package pkg

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccount_CreateListing(t *testing.T) {
	account := NewAccount(context.TODO(), "0x300b105942d6d181cdfe8199fd48eb09d26efd24", "sepolia")

	nfts, err := account.GetNFTs(context.TODO())
	require.Nil(t, err)

	t.Log(account.CreateListing(context.TODO(), &nfts.Nfts[0], "0.289"))
}
