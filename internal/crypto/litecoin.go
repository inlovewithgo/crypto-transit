package crypto

import (
    "encoding/hex"
    "fmt"

    "github.com/btcsuite/btcd/btcec/v2"
    "github.com/btcsuite/btcd/btcutil"
    "github.com/btcsuite/btcd/chaincfg"
)

type LitecoinWallet struct {
    PrivateKey string
    PublicKey  string
    Address    string
}

func GenerateLitecoinWallet() (*LitecoinWallet, error) {
    priv, err := btcec.NewPrivateKey()
    if err != nil {
        return nil, err
    }
    pub := priv.PubKey()
    pkh := btcutil.Hash160(pub.SerializeCompressed())
    addr, err := btcutil.NewAddressPubKeyHash(pkh, &chaincfg.MainNetParams)
    if err != nil {
        return nil, err
    }
    return &LitecoinWallet{
        PrivateKey: hex.EncodeToString(priv.Serialize()),
        PublicKey:  hex.EncodeToString(pub.SerializeCompressed()),
        Address:    addr.EncodeAddress(),
    }, nil
}

func LoadLitecoinWallet(privateKeyHex string) (*LitecoinWallet, error) {
    privBytes, err := hex.DecodeString(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("invalid private key")
    }
    _, pub := btcec.PrivKeyFromBytes(privBytes)
    pkh := btcutil.Hash160(pub.SerializeCompressed())
    addr, err := btcutil.NewAddressPubKeyHash(pkh, &chaincfg.MainNetParams)
    if err != nil {
        return nil, err
    }
    return &LitecoinWallet{
        PrivateKey: privateKeyHex,
        PublicKey:  hex.EncodeToString(pub.SerializeCompressed()),
        Address:    addr.EncodeAddress(),
    }, nil
}