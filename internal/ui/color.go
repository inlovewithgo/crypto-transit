package ui

import (
	"fmt"
	"strings"
)

const (
    Reset   = "\033[0m"
    Red     = "\033[31m"
    Green   = "\033[32m"
    Yellow  = "\033[33m"
    Blue    = "\033[34m"
    Magenta = "\033[35m"
    Cyan    = "\033[36m"
    White   = "\033[97m"
    Bold    = "\033[1m"
    Under   = "\033[4m"
)

func PrintBanner() {
    border := Cyan + "╔" + line("═", 46) + "╗" + Reset
    title := Bold + Under + "LITECOIN WALLET" + Reset
    fmt.Println(border)
    pad := (46-len("LITECOIN WALLET"))/2
    fmt.Printf("%s║%s%s%s%s║\n", Cyan, strings.Repeat(" ", pad), title, strings.Repeat(" ", 46-pad-len("LITECOIN WALLET")), Reset)
    fmt.Println(Cyan + "╚" + line("═", 46) + "╝" + Reset)
}


func PrintMenu(title string, items []string) {
    fmt.Println(Blue + "╔" + line("─", 44) + "╗" + Reset)
    fmt.Printf("%s║%-44s║\n", Blue, title)
    fmt.Println(Blue + "╠" + line("═", 44) + "╣" + Reset)
    for _, it := range items {
        fmt.Printf("%s║ %-43s║\n", Blue, it)
    }
    fmt.Println(Blue + "╚" + line("─", 44) + "╝" + Reset)
}


func PrintSection(title string) {
    fmt.Printf("\n%s╭─[ %s ]─╮%s\n", Magenta, title, Reset)
}

func PrintSuccess(text string)  { fmt.Printf("%s✔ %s%s\n", Green, text, Reset) }
func PrintInfo(text string)     { fmt.Printf("%s» %s%s\n", Cyan, text, Reset) }
func PrintError(text string)    { fmt.Printf("%s✖ %s%s\n", Red, text, Reset) }
func PrintPrompt(text string)   { fmt.Printf("%s%s%s", Yellow+Bold, text, Reset) }

func line(char string, n int) string { s := ""; for i := 0; i < n; i++ { s += char }; return s }
