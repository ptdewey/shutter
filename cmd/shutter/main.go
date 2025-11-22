package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ptdewey/shutter"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: shutter [COMMAND]

Commands:
  review      Review and accept/reject new snapshots (default)
  accept-all  Accept all new snapshots
  reject-all  Reject all new snapshots
  help        Show this help message

Examples:
  shutter              # Start interactive review
  shutter review       # Same as above
  shutter accept-all   # Accept all new snapshots
  shutter reject-all   # Reject all new snapshots
`)
	}

	flag.Parse()

	var cmd string
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	var err error
	switch cmd {
	case "", "review":
		err = shutter.Review()
	case "accept-all":
		err = shutter.AcceptAll()
	case "reject-all":
		err = shutter.RejectAll()
	case "help", "-h", "--help":
		flag.Usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
