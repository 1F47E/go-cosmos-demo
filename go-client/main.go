package main

import (
	"context"
	"encoding/json"
	"fmt"

	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	nodeEndpoint = "127.0.0.1:9090"
)

type Client struct {
	conn       *grpc.ClientConn
	bankClient banktypes.QueryClient
}

type Balance struct {
	Coin  string
	Value string
}

func NewClient(nodeEndpoint string) (*Client, error) {
	grpcConn, err := grpc.Dial(
		nodeEndpoint,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		return nil, err
	}

	bankClient := banktypes.NewQueryClient(grpcConn)

	return &Client{conn: grpcConn, bankClient: bankClient}, nil
}

func (c *Client) GetBalance(address types.AccAddress) ([]Balance, error) {
	req := &banktypes.QueryAllBalancesRequest{
		Address: address.String(),
	}
	resp, err := c.bankClient.AllBalances(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var balances []Balance
	for _, coin := range resp.Balances {
		balances = append(balances, Balance{Coin: coin.Denom, Value: coin.Amount.String()})
	}

	return balances, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <validator address>", os.Args[0])
	}
	addr := os.Args[1]
	myAddress, err := types.AccAddressFromBech32(addr)
	if err != nil {
		log.Fatalf("Failed to parse address: %v", err)
	}

	client, err := NewClient(nodeEndpoint)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// log.Printf("Fetching balance for %s...\n", myAddress)

	balances, err := client.GetBalance(myAddress)
	if err != nil {
		log.Fatalf("Failed to fetch balance: %v", err)
	}

	// convert balances to json
	ret := new(BalanceResponse)
	ret.Balances = balances
	jsonBytes, err := json.Marshal(ret)
	if err != nil {
		log.Fatalf("Failed to marshal json: %v", err)
	}
	fmt.Println(string(jsonBytes))
}
