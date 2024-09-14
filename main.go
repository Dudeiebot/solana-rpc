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
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gagliardetto/solana-go/text"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}
	// Create a new account:
	// account := solana.NewWallet()
	// fmt.Println("account private key:", account.PrivateKey)
	// fmt.Println("account public key:", account.PublicKey())

	// Create a new RPC client:
	client := rpc.New(os.Getenv("RPC_URL"))

	wsClient, err := ws.Connect(context.Background(), os.Getenv("WS_URL"))
	if err != nil {
		panic(err)
	}

	accountFrom, err := solana.PrivateKeyFromBase58(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}
	fmt.Println("public key:", accountFrom.PublicKey().String())
	// Airdrop 1 SOL to the new account:

	accountTo := solana.MustPublicKeyFromBase58(os.Getenv("PUBLIC_KEY"))
	amount := uint64(100000)

	recent, err := client.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(amount, accountFrom.PublicKey(), accountTo).
				Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(accountFrom.PublicKey()),
	)
	if err != nil {
		panic(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if accountFrom.PublicKey().Equals(key) {
			return &accountFrom
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}

	spew.Dump(tx)
	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer Sol"))

	sig, err := confirm.SendAndConfirmTransaction(context.Background(), client, wsClient, tx)
	if err != nil {
		panic(err)
	}
	spew.Dump(sig)

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
