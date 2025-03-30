package api

import (
	"encoding/json"
	"net/http"
	"unknownberrytrip/internal/blockchain"
	"unknownberrytrip/internal/transaction"
)

// StartAPI starts the HTTP server for interacting with the blockchain
func StartAPI(bc *blockchain.Blockchain) {
	http.HandleFunc("/sendTransaction", func(w http.ResponseWriter, r *http.Request) {
		var tx transaction.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := bc.AddTransactionToPool(tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction added to pool"))
	})

	go http.ListenAndServe(":8080", nil)
}
