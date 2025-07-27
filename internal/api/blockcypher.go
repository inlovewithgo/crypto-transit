package api

import (
    "bytes"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"

    "litecoin-wallet/internal/models"
    "github.com/btcsuite/btcd/btcec/v2"
    "github.com/btcsuite/btcd/btcec/v2/ecdsa"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
)

const BlockCypherBaseURL = "https://api.blockcypher.com/v1/ltc/main"

type BlockCypherClient struct {
    BaseURL string
    Client  *http.Client
}

func NewBlockCypherClient() *BlockCypherClient {
    return &BlockCypherClient{
        BaseURL: BlockCypherBaseURL,
        Client:  &http.Client{},
    }
}

func (bc *BlockCypherClient) GetAddressInfo(address string) (models.AddressOverview, error) {
    url := fmt.Sprintf("%s/addrs/%s?limit=10", bc.BaseURL, address)
    resp, err := bc.Client.Get(url)
    if err != nil {
        return models.AddressOverview{}, err
    }
    defer resp.Body.Close()
    var info models.AddressOverview
    err = json.NewDecoder(resp.Body).Decode(&info)
    return info, err
}

func (bc *BlockCypherClient) GetBalance(address string) (int64, error) {
    url := fmt.Sprintf("%s/addrs/%s/balance", bc.BaseURL, address)
    resp, err := bc.Client.Get(url)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    var response struct {
        Balance int64 `json:"balance"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return 0, err
    }
    return response.Balance, nil
}

func (bc *BlockCypherClient) SendTransaction(privateKeyHex, fromAddress, toAddress string, amount int64, sendAll bool) (string, error) {
    var txReq map[string]interface{}
    if sendAll {
        txReq = map[string]interface{}{
            "inputs":  []map[string]interface{}{{"addresses": []string{fromAddress}}},
            "outputs": []map[string]interface{}{{"addresses": []string{toAddress}}},
        }
    } else {
        txReq = map[string]interface{}{
            "inputs":  []map[string]interface{}{{"addresses": []string{fromAddress}}},
            "outputs": []map[string]interface{}{{"addresses": []string{toAddress}, "value": amount}},
        }
    }
    jsonData, _ := json.Marshal(txReq)
    url := fmt.Sprintf("%s/txs/new", bc.BaseURL)
    resp, err := bc.Client.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusCreated {
        var res map[string]interface{}
        json.Unmarshal(body, &res)
        if errs, ok := res["errors"].([]interface{}); ok {
            for _, e := range errs {
                if errStr, ok := e.(map[string]interface{})["error"].(string); ok {
                    switch {
                    case strings.Contains(errStr, "Insufficient funds"):
                        return "", fmt.Errorf("You have insufficient balance for this operation.")
                    case strings.Contains(errStr, "can't have zero for value"):
                        return "", fmt.Errorf("Cannot send zero coins. Enter a valid amount.")
                    case strings.Contains(errStr, "Unable to find a transaction to spend"):
                        return "", fmt.Errorf("No Litecoin ever deposited to this wallet. It cannot spend.")
                    }
                    return "", fmt.Errorf(errStr)
                }
            }
        }
        return "", fmt.Errorf("Service error: %s", string(body))
    }
    var txSkeleton struct {
        ToSign []string `json:"tosign"`
        Tx     struct {
            Hash string `json:"hash"`
        } `json:"tx"`
        TxFee int64 `json:"fees"`
    }
    if err := json.Unmarshal(body, &txSkeleton); err != nil {
        return "", err
    }
    privBytes, _ := hex.DecodeString(privateKeyHex)
    priv, _ := btcec.PrivKeyFromBytes(privBytes)
    sigs := make([]string, len(txSkeleton.ToSign))
    for i, tosign := range txSkeleton.ToSign {
        hash, _ := hex.DecodeString(tosign)
        msgHash, _ := chainhash.NewHash(hash)
        signature := ecdsa.Sign(priv, msgHash[:])
        sigs[i] = hex.EncodeToString(signature.Serialize())
    }
    signedTx := map[string]interface{}{
        "signatures": sigs,
        "pubkeys":    []string{hex.EncodeToString(priv.PubKey().SerializeCompressed())},
        "tx":         txSkeleton.Tx,
    }
    signedTxJson, _ := json.Marshal(signedTx)
    url = fmt.Sprintf("%s/txs/send", bc.BaseURL)
    resp2, err := bc.Client.Post(url, "application/json", bytes.NewBuffer(signedTxJson))
    if err != nil {
        return "", err
    }
    defer resp2.Body.Close()
    var result struct{ Tx struct{ Hash string `json:"hash"` } `json:"tx"` }
    if err := json.NewDecoder(resp2.Body).Decode(&result); err != nil {
        body2, _ := io.ReadAll(resp2.Body)
        return "", fmt.Errorf("broadcast error: %s", string(body2))
    }
    return result.Tx.Hash, nil
}
