// Command raid-signup 跑团队报名样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/raid"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("raid-signup", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "quota", "quota / dup / unknown / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "quota":
		desk := raid.New(map[string]int{"tank": 1})
		desk.Signup("a", "tank")
		desk.Signup("b", "tank")
		fmt.Fprintf(stdout, "roster=%d waiting=%d\n", len(desk.Roster("tank")), len(desk.Waiting()))
	case "dup":
		desk := raid.New(map[string]int{"tank": 2})
		desk.Signup("a", "tank")
		desk.Signup("a", "tank")
		fmt.Fprintf(stdout, "roster=%d\n", len(desk.Roster("tank")))
	case "unknown":
		desk := raid.New(map[string]int{"tank": 1})
		fmt.Fprintf(stdout, "ok=%v\n", desk.Leave("nobody"))
	case "scan":
		desk := raid.New(map[string]int{"tank": 1000000})
		for index := 0; index < 3000; index++ {
			desk.Signup(fmt.Sprintf("p-%d", index), "tank")
		}
		for index := 0; index < 3000; index++ {
			desk.Roster("tank")
		}
		fmt.Fprintf(stdout, "scanned=%d\n", desk.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample quota|dup|unknown|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
