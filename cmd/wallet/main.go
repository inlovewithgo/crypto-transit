package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"

    "github.com/mdp/qrterminal/v3"
    "litecoin-wallet/internal/api"
    "litecoin-wallet/internal/crypto"
    "litecoin-wallet/internal/db"
    "litecoin-wallet/internal/ui"
    "litecoin-wallet/internal/wallet"
	qrcode "github.com/skip2/go-qrcode"
)

var lastSyncTime string
var lastBalance float64

func main() {
    _, ierr := db.InitDB()
    if ierr != nil {
        ui.PrintError("Could not create or open database: " + ierr.Error())
        os.Exit(1)
    }
    scanner := bufio.NewScanner(os.Stdin)
    w := &wallet.Wallet{}
    apiClient := api.NewBlockCypherClient()

    for {
        if w.PrivateKey == "" {
            ui.PrintBanner()
            items := []string{"1. Generate new wallet", "2. Load wallet from disk"}
            ui.PrintMenu("MAIN MENU", items)
            ui.PrintPrompt("Select option: ")
            scanner.Scan()
            choice := strings.TrimSpace(scanner.Text())
            switch choice {
            case "1":
                generateWallet(w, scanner)
            case "2":
                loadWallet(w, apiClient, scanner)
            default:
                ui.PrintError("Invalid choice.")
            }
            continue
        }
        walletAppMenu(w, apiClient, scanner)
        fmt.Printf("\n%sPress ENTER to continue...%s", ui.Yellow, ui.Reset)
        scanner.Scan()
        fmt.Print("\033[H\033[2J")
    }
}

func generateWallet(w *wallet.Wallet, scanner *bufio.Scanner) {
    wlt, err := crypto.GenerateLitecoinWallet()
    if err != nil {
        ui.PrintError("Failed to generate wallet: " + err.Error())
        return
    }
    w.PrivateKey = wlt.PrivateKey
    w.PublicKey = wlt.PublicKey
    w.Address = wlt.Address
    alias := "TEMP"
    ui.PrintInfo(fmt.Sprintf("Address: %s%s%s", ui.Cyan, w.Address, ui.Reset))
    ui.PrintInfo(fmt.Sprintf("Private: %s%s%s", ui.Yellow, w.PrivateKey, ui.Reset))
    ui.PrintPrompt("Set an alias for this wallet (default TEMP): ")
    scanner.Scan()
    userAlias := strings.TrimSpace(scanner.Text())
    if userAlias != "" {
        alias = userAlias
    }
    w.Alias = alias
    ui.PrintPrompt("Save this wallet locally for next time? (y/N): ")
    scanner.Scan()
    save := strings.TrimSpace(strings.ToLower(scanner.Text()))
    if save == "y" || save == "yes" {
        err = db.SaveWallet(w.Alias, w.PrivateKey, w.PublicKey, w.Address)
        if err == nil {
            ui.PrintSuccess("Wallet has been saved locally.")
        } else {
            ui.PrintError("Failed to save wallet!")
        }
    } else {
        ui.PrintInfo("Wallet not saved. It will not persist after logout or app exit.")
    }
}

func loadWallet(w *wallet.Wallet, apiClient *api.BlockCypherClient, scanner *bufio.Scanner) {
    aliases, err := db.ListWalletAliases()
    if err != nil || len(aliases) == 0 {
        ui.PrintError("No saved wallets found.")
        return
    }
    bal := make([]float64, len(aliases))
    for i, alias := range aliases {
        _, _, addr, _, _ := db.LoadWallet(alias)
        b, _ := apiClient.GetBalance(addr)
        bal[i] = float64(b) / 100000000
    }
    ui.PrintSection("Pick a wallet")
    for i, alias := range aliases {
        fmt.Printf("%s[%d]%s %s (%.8f LTC)\n", ui.Blue, i+1, ui.Reset, alias, bal[i])
    }
    ui.PrintPrompt("Select wallet by number: ")
    scanner.Scan()
    idx, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))
    if idx < 1 || idx > len(aliases) {
        ui.PrintError("Invalid selection.")
        return
    }
    priv, pub, addr, found, err := db.LoadWallet(aliases[idx-1])
    if !found || err != nil {
        ui.PrintError("Load error.")
        return
    }
    w.PrivateKey, w.PublicKey, w.Address, w.Alias = priv, pub, addr, aliases[idx-1]
    ui.PrintSuccess("Loaded wallet '" + w.Alias + "'")
}

func walletAppMenu(w *wallet.Wallet, apiClient *api.BlockCypherClient, scanner *bufio.Scanner) {
    ui.PrintBanner()
    shortAddr := w.Address[:6] + "..." + w.Address[len(w.Address)-6:]
    menu := []string{
        fmt.Sprintf("Alias: %s%s%s", ui.Green, w.Alias, ui.Reset),
        fmt.Sprintf("Address: %s%s%s", ui.Yellow, shortAddr, ui.Reset),
        fmt.Sprintf("Last balance: %s%.8f LTC%s", ui.Blue, lastBalance, ui.Reset),
        "",
        "1. Wallet overview",
        "2. Transaction history",
        "3. Send transaction",
        "4. Receive (show QR/address)",
        "5. Move funds to another wallet",
        "6. Change alias",
        "7. Delete this wallet",
        "8. Resync balance",
        "9. Export transactions as CSV",
        "10. Save address QR as PNG",
        "11. Vanity address generator",
        "12. Bulk wallet generator",
        "13. Logout",
        "0. Exit",
    }
    ui.PrintMenu("WALLET MENU", menu[3:])
    ui.PrintPrompt("Select option: ")
    scanner.Scan()
    choice := strings.TrimSpace(scanner.Text())
    switch choice {
    case "1":
        walletOverview(w, apiClient)
    case "2":
        showTxnHistory(w, apiClient)
    case "3":
        sendTransaction(w, apiClient, scanner)
    case "4":
        showReceive(w, scanner)
    case "5":
        moveFunds(w, apiClient, scanner)
    case "6":
        changeAlias(w, scanner)
    case "7":
        deleteCurrentWallet(w, scanner)
    case "8":
        resyncBalance(w, apiClient)
    case "9":
        exportTxCSV(w, apiClient)
    case "10":
        saveAddressQRPNG(w, scanner)
    case "11":
        vanityGenerator(scanner)
    case "12":
        bulkWalletGen(scanner)
    case "13":
        logoutWallet(w)
    case "0":
        ui.PrintInfo("Exiting...")
        os.Exit(0)
    default:
        ui.PrintError("Invalid choice.")
    }
}

func walletOverview(w *wallet.Wallet, apiClient *api.BlockCypherClient) {
    info, err := apiClient.GetAddressInfo(w.Address)
    if err != nil {
        ui.PrintError("API error: " + err.Error())
        return
    }
    lastBalance = float64(info.Balance) / 1e8
    lastSyncTime = time.Now().Format("02 Jan 2006 15:04:05")
    fmt.Printf("%sWallet alias:%s   %s\n", ui.Cyan, ui.Reset, w.Alias)
    fmt.Printf("%sAddress:%s       %s\n", ui.Cyan, ui.Reset, w.Address)
    fmt.Printf("%sBalance:%s       %.8f LTC\n", ui.Cyan, ui.Reset, lastBalance)
    fmt.Printf("%sTotal received:%s %.8f LTC\n", ui.Cyan, ui.Reset, float64(info.TotalReceived)/1e8)
    fmt.Printf("%sTotal sent:%s     %.8f LTC\n", ui.Cyan, ui.Reset, float64(info.TotalSent)/1e8)
    fmt.Printf("%sTx Count:%s       %d\n", ui.Cyan, ui.Reset, info.NTx)
}

func resyncBalance(w *wallet.Wallet, apiClient *api.BlockCypherClient) {
    info, err := apiClient.GetAddressInfo(w.Address)
    if err != nil {
        ui.PrintError("Failed to sync: " + err.Error())
        return
    }
    lastBalance = float64(info.Balance) / 1e8
    lastSyncTime = time.Now().Format("02 Jan 2006 15:04:05")
    ui.PrintSuccess(fmt.Sprintf("Synced! Balance now: %.8f LTC", lastBalance))
}

func showTxnHistory(w *wallet.Wallet, apiClient *api.BlockCypherClient) {
    info, err := apiClient.GetAddressInfo(w.Address)
    if err != nil {
        ui.PrintError("API error: " + err.Error())
        return
    }
    if len(info.Txrefs) == 0 {
        fmt.Println("(No transactions found)")
        return
    }
    fmt.Println(ui.Yellow + "Last transactions:")
    for i, t := range info.Txrefs {
        fmt.Printf(ui.Blue+" %2d. Time: %v\n     Hash: %s\n     Amount: %.8f LTC\n     Confirmations: %d\n"+ui.Reset,
            i+1, t.Received, t.Hash, float64(t.Value)/1e8, t.Confirmations)
        if i >= 9 {
            break
        }
    }
}

func sendTransaction(w *wallet.Wallet, apiClient *api.BlockCypherClient, scanner *bufio.Scanner) {
    if w.Address == "" {
        ui.PrintInfo("Generate or load a wallet first.")
        return
    }
    ui.PrintPrompt("Recipient address: ")
    scanner.Scan()
    toAddress := strings.TrimSpace(scanner.Text())
    ui.PrintPrompt("Amount (LTC) or type 'all' to send all: ")
    scanner.Scan()
    amountStr := strings.TrimSpace(scanner.Text())

    var amount float64
    var err error
    var sendAll bool
    var txHash string

    fee := int64(10000)

    if strings.ToLower(amountStr) == "all" {
        sendAll = true
        info, err := apiClient.GetAddressInfo(w.Address)
        if err != nil {
            ui.PrintError("API error: " + err.Error())
            return
        }
        totalBalance := info.Balance
        if totalBalance <= fee {
            ui.PrintError("No available balance to send after subtracting fee.")
            return
        }
        amount = float64(totalBalance-fee) / 1e8
        if amount <= 0 {
            ui.PrintError("No available balance to send after fees.")
            return
        }
    } else {
        amount, err = strconv.ParseFloat(amountStr, 64)
        if err != nil || amount <= 0 {
            ui.PrintError("Invalid amount entered.")
            return
        }
        sendAll = false
    }

    if sendAll {
        info, _ := apiClient.GetAddressInfo(w.Address)
        txHash, err = apiClient.SendTransaction(
            w.PrivateKey, w.Address, toAddress, info.Balance-fee, false,
        )
    } else {
        txHash, err = apiClient.SendTransaction(
            w.PrivateKey, w.Address, toAddress, int64(amount*1e8), false,
        )
    }

    if err != nil {
        msg := err.Error()
        switch {
        case strings.Contains(msg, "Insufficient funds"):
            ui.PrintError("Insufficient balance for this transaction.")
        case strings.Contains(msg, "zero for value"):
            ui.PrintError("Cannot send zero coins. Enter a valid amount.")
        case strings.Contains(msg, "Unable to find a transaction to spend"):
            ui.PrintError("This wallet has no LTC sent to it yet (no UTXOs to spend).")
        default:
            ui.PrintError(msg)
        }
        return
    }
    ui.PrintSuccess("Transaction sent successfully!")
    fmt.Printf("Explorer link: %shttps://live.blockcypher.com/ltc/tx/%s%s\n",
        ui.Blue, txHash, ui.Reset)
}

func showReceive(w *wallet.Wallet, scanner *bufio.Scanner) {
    ui.PrintSection("Receive Litecoin")
    ui.PrintInfo("Share your public address or QR below for payments.")
    fmt.Printf("%sAddress: %s%s%s\n\n", ui.Bold, ui.Yellow, w.Address, ui.Reset)
    qrterminal.Generate(w.Address, qrterminal.L, os.Stdout)
    ui.PrintPrompt("Copy address to clipboard (y/N)? ")
    scanner.Scan()
    inp := strings.ToLower(strings.TrimSpace(scanner.Text()))
    if inp == "y" || inp == "c" {
        copyToClipboard(w.Address)
        ui.PrintSuccess("Address copied to clipboard (if supported on this OS).")
    }
}

func copyToClipboard(addr string) {
    if _, err := exec.LookPath("pbcopy"); err == nil {
        c := exec.Command("pbcopy")
        c.Stdin = strings.NewReader(addr)
        c.Run()
        return
    }
    if _, err := exec.LookPath("xclip"); err == nil {
        c := exec.Command("xclip", "-selection", "clipboard")
        c.Stdin = strings.NewReader(addr)
        c.Run()
        return
    }
    if _, err := exec.LookPath("xsel"); err == nil {
        c := exec.Command("xsel", "--clipboard", "--input")
        c.Stdin = strings.NewReader(addr)
        c.Run()
        return
    }
    if _, err := exec.LookPath("clip"); err == nil {
        c := exec.Command("clip")
        c.Stdin = strings.NewReader(addr)
        c.Run()
        return
    }
}

func moveFunds(w *wallet.Wallet, apiClient *api.BlockCypherClient, scanner *bufio.Scanner) {
    aliases, err := db.ListWalletAliases()
    if err != nil || len(aliases) == 0 {
        ui.PrintError("No saved wallets found.")
        return
    }
    var targets []string
    for _, alias := range aliases {
        _, _, addr, found, _ := db.LoadWallet(alias)
        if found && addr != w.Address {
            targets = append(targets, alias)
        }
    }
    if len(targets) == 0 {
        ui.PrintError("No destination wallets available.")
        return
    }
    ui.PrintSection("Select destination wallet:")
    for i, a := range targets {
        _, _, addr, _, _ := db.LoadWallet(a)
        bal, _ := apiClient.GetBalance(addr)
        fmt.Printf("%s[%d]%s %s (%.8f LTC)\n", ui.Blue, i+1, ui.Reset, a, float64(bal)/1e8)
    }
    ui.PrintPrompt("Choose: ")
    scanner.Scan()
    idx, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))
    if idx < 1 || idx > len(targets) {
        ui.PrintError("Invalid selection.")
        return
    }
    _, _, destAddr, _, _ := db.LoadWallet(targets[idx-1])
    ui.PrintPrompt("Amount (LTC): ")
    scanner.Scan()
    amtStr := strings.TrimSpace(scanner.Text())
    amt, _ := strconv.ParseFloat(amtStr, 64)
    if amt <= 0 {
        ui.PrintError("Invalid amount.")
        return
    }
    txHash, err := apiClient.SendTransaction(w.PrivateKey, w.Address, destAddr, int64(amt*100000000), false)
    if err != nil {
        ui.PrintError(err.Error())
        return
    }
    ui.PrintSuccess("Funds moved. Tx hash: " + txHash)
}

func changeAlias(w *wallet.Wallet, scanner *bufio.Scanner) {
    ui.PrintPrompt("Enter new alias for this wallet (current: " + w.Alias + "): ")
    scanner.Scan()
    newAlias := strings.TrimSpace(scanner.Text())
    if newAlias == "" {
        ui.PrintInfo("No changes made.")
        return
    }
    err := db.SaveWallet(newAlias, w.PrivateKey, w.PublicKey, w.Address)
    if err == nil {
        _ = db.DeleteWallet(w.Alias)
        w.Alias = newAlias
        ui.PrintSuccess("Alias changed to: " + newAlias)
    } else {
        ui.PrintError("Failed to change alias.")
    }
}

func deleteCurrentWallet(w *wallet.Wallet, scanner *bufio.Scanner) {
    ui.PrintPrompt("Are you sure you want to delete this wallet, type its alias (" + w.Alias + ") to confirm: ")
    scanner.Scan()
    conf := strings.TrimSpace(scanner.Text())
    if conf != w.Alias {
        ui.PrintError("Wallet deletion cancelled.")
        return
    }
    _ = db.DeleteWallet(w.Alias)
    w.PrivateKey, w.PublicKey, w.Address, w.Alias = "", "", "", ""
    ui.PrintSuccess("Wallet deleted.")
}

func logoutWallet(w *wallet.Wallet) {
    w.PrivateKey, w.PublicKey, w.Address, w.Alias = "", "", "", ""
    ui.PrintInfo("Logged out of wallet session. Returning to main screen.")
}


func exportTxCSV(w *wallet.Wallet, apiClient *api.BlockCypherClient) {
    info, err := apiClient.GetAddressInfo(w.Address)
    if err != nil {
        ui.PrintError("API error: " + err.Error())
        return
    }
    fn := fmt.Sprintf("%s_%s.csv", w.Alias, time.Now().Format("20060102_150405"))
    f, err := os.Create(fn)
    if err != nil {
        ui.PrintError("Failed to create export file.")
        return
    }
    defer f.Close()
    wtr := csv.NewWriter(f)
    wtr.Write([]string{"Time", "TxHash", "Value", "Confirmations"})
    for _, t := range info.Txrefs {
        wtr.Write([]string{
            t.Received,
            t.Hash,
            fmt.Sprintf("%.8f", float64(t.Value)/1e8),
            strconv.Itoa(t.Confirmations),
        })
    }
    wtr.Flush()
    ui.PrintSuccess("Exported to: " + fn)
}

func saveAddressQRPNG(w *wallet.Wallet, scanner *bufio.Scanner) {
    ui.PrintPrompt("Enter PNG filename (default: address.png): ")
    scanner.Scan()
    fname := strings.TrimSpace(scanner.Text())
    if fname == "" {
        fname = "address.png"
    }
    err := qrcode.WriteFile(w.Address, qrcode.Medium, 256, fname)
    if err != nil {
        ui.PrintError("Couldn't save: " + err.Error())
    } else {
        ui.PrintSuccess("QR PNG saved as: " + fname)
    }
}


func vanityGenerator(scanner *bufio.Scanner) {
    ui.PrintPrompt("Enter a prefix to search for (e.g. lt, L, etc): ")
    scanner.Scan()
    prefix := strings.TrimSpace(scanner.Text())
    if prefix == "" {
        ui.PrintError("Invalid prefix")
        return
    }
    t0 := time.Now()
    for i := 1; ; i++ {
        lw, _ := crypto.GenerateLitecoinWallet()
        if strings.HasPrefix(strings.ToLower(lw.Address), strings.ToLower(prefix)) {
            fmt.Printf("Found: %s\nPrivate: %s\n", lw.Address, lw.PrivateKey)
            break
        }
        if i%10000 == 0 && time.Since(t0) > 10*time.Second {
            ui.PrintError("Stopping after 10s, not found.")
            break
        }
    }
}

func bulkWalletGen(scanner *bufio.Scanner) {
    ui.PrintPrompt("How many wallets? ")
    scanner.Scan()
    n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))
    if n <= 0 {
        ui.PrintError("Number too low.")
        return
    }
    for i := 0; i < n; i++ {
        lw, _ := crypto.GenerateLitecoinWallet()
        alias := fmt.Sprintf("Bulk%d", i+1)
        db.SaveWallet(alias, lw.PrivateKey, lw.PublicKey, lw.Address)
        fmt.Printf("[%d] %s - %s\n", i+1, lw.Address, lw.PrivateKey)
    }
    ui.PrintSuccess("Bulk wallets generated!")
}