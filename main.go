package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

type transferInfo struct {
	senderKey     string
	recipientAddr string
	rpcClient     *rpc.Client
	amount        uint64
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}
	// rpcClient := rpc.New(os.Getenv("RPC_URL"))
	rpcClient := rpc.New(rpc.DevNet_RPC)

	if err := createAccount(100000, rpcClient); err != nil {
		log.Fatalf("Error creating account %v", err)
	}

	// t := &transferInfo{
	// 	senderKey:     os.Getenv("PRIVATE_KEY"),
	// 	recipientAddr: os.Getenv("PUBLIC_KEY"),
	// 	rpcClient:     rpcClient,
	// 	amount:        100000,
	// }
	//
	// if err := sendAndConfirmTransaction(t); err != nil {
	// 	log.Fatalf("Error sending transaction: %v", err)
	// }
}

func sendAndConfirmTransaction(t *transferInfo) error {
	accountFrom, err := solana.PrivateKeyFromBase58(t.senderKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("public key:", accountFrom.PublicKey().String())

	accountTo := solana.MustPublicKeyFromBase58(t.recipientAddr)

	recent, err := t.rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(t.amount, accountFrom.PublicKey(), accountTo).
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

	sig, err := t.rpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		panic(fmt.Errorf("error sending transaction: %v", err))
	}

	fmt.Printf("Transaction sent: %s\n", sig)
	return nil
}

func createAccount(amount uint64, rpcClient *rpc.Client) error {
	account := solana.NewWallet()
	fmt.Println("account private key:", account.PrivateKey)
	fmt.Println("account public key:", account.PublicKey())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	out, err := rpcClient.RequestAirdrop(
		ctx,
		account.PublicKey(),
		amount,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return fmt.Errorf("failed to request airdrop: %w", err)
	}

	fmt.Println("Airdrop transaction signature:", out)
	return nil
}
