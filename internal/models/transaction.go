package models

type Transaction struct {
    Hash          string   `json:"hash"`
    Confirmations int      `json:"confirmations"`
    Value         int64    `json:"value"`
    Received      string   `json:"received"`
    Addresses     []string `json:"addresses"`
}
type AddressOverview struct {
    Balance        int64         `json:"balance"`
    TotalReceived  int64         `json:"total_received"`
    TotalSent      int64         `json:"total_sent"`
    NTx            int           `json:"n_tx"`
    Txrefs         []Transaction `json:"txrefs"`
    UnconfirmedBalance int64     `json:"unconfirmed_balance"`
}
