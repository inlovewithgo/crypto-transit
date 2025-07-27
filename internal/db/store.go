package db

import (
    "database/sql"
    "fmt"
    "os"
	"strings"
    "path/filepath"
    _ "modernc.org/sqlite"
)

const (
    walletDBFile    = "litecoin_wallet.db"
    TempWalletAlias = "TEMP"

    cReset  = "\033[0m"
    cGreen  = "\033[32m"
    cYellow = "\033[33m"
    cCyan   = "\033[36m"
    cRed    = "\033[31m"
)

func getWalletDBPath() string {
    dir, err := os.Getwd()
    if err != nil {
        dir = "."
    }
    path := filepath.Join(dir, walletDBFile)
    fmt.Printf("%s[INFO]%s Path to wallet database: %s\n", cCyan, cReset, path)
    return path
}

func nice(err error) string {
    if err == nil {
        return fmt.Sprintf("%s[SUCCESS]%s Done.\n", cGreen, cReset)
    }
    return fmt.Sprintf("%s[ERROR]%s %v\n", cRed, cReset, err)
}

func InitDB() (*sql.DB, error) {
    fmt.Printf("%s[INFO]%s Opening wallet database...\n", cCyan, cReset)
    db, err := sql.Open("sqlite", getWalletDBPath())
    if err != nil {
        fmt.Print(nice(err))
        return nil, err
    }
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS wallet (
            alias TEXT PRIMARY KEY,
            private TEXT NOT NULL,
            public TEXT NOT NULL,
            address TEXT NOT NULL
        );
    `)
    if err != nil {
        fmt.Print(nice(err))
        return nil, err
    }
    fmt.Printf("%s[SUCCESS]%s Opened or created wallet database.\n", cGreen, cReset)
    return db, nil
}

func SaveWallet(alias, priv, pub, addr string) error {
    fmt.Printf("%s[INFO]%s Saving wallet: alias=%s... ", cCyan, cReset, alias)
    db, err := InitDB()
    if err != nil {
        fmt.Print(nice(err))
        return err
    }
    defer db.Close()
    _, err = db.Exec(`
        INSERT OR REPLACE INTO wallet(alias, private, public, address) VALUES(?, ?, ?, ?)`, alias, priv, pub, addr)
    fmt.Print(nice(err))
    return err
}

func LoadWallet(alias string) (priv, pub, addr string, found bool, err error) {
    fmt.Printf("%s[INFO]%s Loading wallet: alias=%s...\n", cCyan, cReset, alias)
    db, err := InitDB()
    if err != nil {
        fmt.Print(nice(err))
        return "", "", "", false, err
    }
    defer db.Close()
    row := db.QueryRow(`SELECT private, public, address FROM wallet WHERE alias=?`, alias)
    err = row.Scan(&priv, &pub, &addr)
    if err == sql.ErrNoRows {
        fmt.Printf("%s[WARN]%s No record for alias %s\n", cYellow, cReset, alias)
        return "", "", "", false, nil
    }
    if err != nil {
        fmt.Print(nice(err))
    } else {
        fmt.Printf("%s[SUCCESS]%s Wallet loaded: %s\n", cGreen, cReset, alias)
    }
    return priv, pub, addr, err == nil, err
}

func DeleteWallet(alias string) error {
    fmt.Printf("%s[INFO]%s Deleting wallet: alias=%s... ", cCyan, cReset, alias)
    db, err := InitDB()
    if err != nil {
        fmt.Print(nice(err))
        return err
    }
    defer db.Close()
    _, err = db.Exec(`DELETE FROM wallet WHERE alias=?`, alias)
    fmt.Print(nice(err))
    return err
}

func ListWalletAliases() ([]string, error) {
    fmt.Printf("%s[INFO]%s Getting list of all wallet aliases...\n", cCyan, cReset)
    db, err := InitDB()
    if err != nil {
        fmt.Print(nice(err))
        return nil, err
    }
    defer db.Close()
    rows, err := db.Query(`SELECT alias FROM wallet`)
    if err != nil {
        fmt.Print(nice(err))
        return nil, err
    }
    defer rows.Close()
    var aliases []string
    for rows.Next() {
        var alias string
        if err := rows.Scan(&alias); err == nil {
            aliases = append(aliases, alias)
        }
    }
    if len(aliases) > 0 {
        fmt.Printf("%s[SUCCESS]%s Found aliases: %s\n", cGreen, cReset, strings.Join(aliases, ", "))
    } else {
        fmt.Printf("%s[WARN]%s No wallets stored yet.\n", cYellow, cReset)
    }
    return aliases, nil
}
