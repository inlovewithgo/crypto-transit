# 🟣 Crypto-Transit — Terminal Litecoin Wallet

Crypto-Transit is a **secure, user-friendly command-line Litecoin wallet** built in Go.  
It allows you to generate, save, manage, and transact using non-custodial Litecoin wallets—right from your terminal.

```
╔══════════════════════════════════════════════╗
║               LITECOIN WALLET               ║
╚══════════════════════════════════════════════╝
```

## 🚀 Features

- **Generate new wallets** (with alias)
- **Save/load wallets** to local encrypted database
- **Show balance, wallet overview, and transaction history**
- **Send LTC (including "send all" minus fee)**
- **Receive LTC with address QR code**
- **Move funds between your own wallets**
- **Change or delete wallet alias**
- **Resync balance, export transactions as CSV**
- **Save your address QR code as PNG**
- **Generate vanity addresses, bulk wallet tools**
- **Clipboard support for addresses**
- **Bulk Wallet Generation**

## 📦 Installation

1. **Clone this repo:**

    ```sh
    git clone https://github.com/inlovewithgo/crypto-transit.git
    cd crypto-transit
    ```

2. **Install Go (if you haven't):**  
   https://go.dev/dl/

3. **Install dependencies:**

    ```sh
    go mod tidy
    ```

4. **Build and run:**

    ```sh
    go run cmd/wallet/main.go
    ```

## 🛠 Usage

### Main Menu
- `1. Generate new wallet` —— Create a new Litecoin wallet with public/private keys.
- `2. Load wallet from disk` —— Load a previously saved wallet by number.

### After loading or generating:

- `1. Wallet overview` — Shows balance, total received/sent, tx count, etc.
- `2. Transaction history` — Shows recent incoming & outgoing txns.
- `3. Send transaction` — LTC transfer to anyone (supports “all”/max send).
- `4. Receive` — Show your address + QR code for others to send LTC to you.
- `5. Move funds` — Move coins between your local wallets.
- `6. Change alias` — Rename a wallet.
- `7. Delete this wallet` — Removes wallet from storage (confirmation required).
- `8. Resync balance` — Updates wallet details from blockchain.
- `9. Export transactions as CSV` — Export your tx history as a .csv file.
- `10. Save address QR as PNG` — Saves your public address QR code as a .png file.
- `11. Vanity address generator` — Mine a pretty-looking LTC address.
- `12. Bulk wallet generator` — Make/seal multiple wallets at once.
- `13. Logout` — Return to main menu.
- `0. Exit` — Safe app shutdown.

## 📷 Some Shots

```
╔══════════════════════════════════════════════╗
║               LITECOIN WALLET               ║
╚══════════════════════════════════════════════╝
╔────────────────────────────────────────────╗
║MAIN MENU                                   ║
╠════════════════════════════════════════════╣
║ 1. Generate new wallet                     ║
║ 2. Load wallet from disk                   ║
╚────────────────────────────────────────────╝
```

```
Receive Litecoin
Share your public address or QR below for payments.
Address: LULzDxb1UJtAp8b53r5BSnX...
████ ███ █  ████  ... (QR code)
Copy address to clipboard (y/N)?
```

## ⚠️ Security Warnings

- **Never share your private key or wallet file!**
- **Back up your wallet keys**; if you lose your .db or keys, your coins are lost.
- QR export, clipboard, and CSV files are saved locally. Treat them as sensitive.

## 💡 Credits

- Blockchain API: [BlockCypher](https://www.blockcypher.com/dev/ltc/)
- QR code: [github.com/mdp/qrterminal](https://github.com/mdp/qrterminal), [github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode)

## 🤝 Contributing

PRs and issues are welcome! Please open an issue or submit a pull request on [GitHub](https://github.com/inlovewithgo/crypto-transit).


**Let me know if you want a Markdown section for troubleshooting, detailed developer info, or a badge for Go version, etc.!**