package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/text"
	"github.com/joho/godotenv"
)

var (
	senderKey     string
	recipientAddr string
)

type userInfo struct {
	senderKey     string
	recipientAddr string
	rpcClient     *rpc.Client
	amount        uint64
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}
	// Create a new account:
	// account := solana.NewWallet()
	// fmt.Println("account private key:", account.PrivateKey)
	// fmt.Println("account public key:", account.PublicKey())

	// Create a new RPC client:

	rpcClient := rpc.New(os.Getenv("RPC_URL"))

	amount := uint64(100000)

	u := &userInfo{
		senderKey:     os.Getenv("PRIVATE_KEY"),
		recipientAddr: os.Getenv("PUBLIC_KEY"),
		rpcClient:     rpcClient,
		amount:        amount,
	}

	if err := sendAndConfirmTransaction(u); err != nil {
		log.Fatalf("Error sending transaction: %v", err)
	}
	// out, err := client.RequestAirdrop(
	// 	context.TODO(),
	// 	account.PublicKey(),
	// 	solana.LAMPORTS_PER_SOL*1,
	// 	rpc.CommitmentFinalized,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("airdrop transaction signature:", out)
}

func sendAndConfirmTransaction(u *userInfo) error {
	accountFrom, err := solana.PrivateKeyFromBase58(u.senderKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("public key:", accountFrom.PublicKey().String())

	accountTo := solana.MustPublicKeyFromBase58(u.recipientAddr)

	recent, err := u.rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(u.amount, accountFrom.PublicKey(), accountTo).
				Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(accountFrom.PublicKey()),
	)
	if err != nil {
		panic(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(accountFrom.PublicKey()) {
			return &accountFrom
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}

	spew.Dump(tx)
	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer Sol"))

	sig, err := u.rpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		panic(fmt.Errorf("error sending transaction: %v", err))
	}

	fmt.Printf("Transaction sent: %s\n", sig)
	return nil
}
