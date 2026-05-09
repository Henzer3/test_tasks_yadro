package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"test.task.dnsmanager/internal/adapters/dns"
	"test.task.dnsmanager/internal/entity"
)

const serverAddress = "localhost:28080"

func main() {
	var ip string
	flag.StringVar(&ip, "i", "", "flag for ip")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  dns-client [flags] <command>

Commands:
  AddDNS       Add DNS server
  DeleteDNS    Delete DNS server
  ListDNS      Show DNS servers list

Flags:
`)
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, `
Examples:
  dns-client -i 8.8.8.8 AddDNS
  dns-client -i 8.8.8.8 DeleteDNS
  dns-client ListDNS
  dns-client --help
`)
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	dns, err := dns.NewClient(serverAddress)
	if err != nil {
		fmt.Println("cant connect to server")
		os.Exit(1)
	}

	defer dns.Close()

	command := flag.Arg(0)

	ctx := context.Background()

	switch command {
	case "AddDNS":
		addDNS(ctx, dns, ip)
	case "DeleteDNS":
		deleteDNS(ctx, dns, ip)
	case "ListDNS":
		listDNS(ctx, dns)
	default:
		fmt.Println("Unknown command, try again")
	}
}

func addDNS(ctx context.Context, dns *dns.Client, ip string) {
	if err := dns.AddDNS(ctx, ip); err != nil {
		switch err {
		case entity.ErrInvalidArgument:
			fmt.Println("invalid ip")
		case entity.ErrAlreadyexist:
			fmt.Println("ip already exist")
		case entity.ErrUnAvailableServer:
			fmt.Println("server Unavailable")
		case entity.ErrInternalError:
			fmt.Println("internal error")
		default:
			fmt.Println("unknown error")
		}
		return
	}

	fmt.Println("added")
}

func deleteDNS(ctx context.Context, dns *dns.Client, ip string) {
	if err := dns.DeleteDNS(ctx, ip); err != nil {
		switch err {
		case entity.ErrInvalidArgument:
			fmt.Println("invalid ip")
		case entity.ErrNotFoundDNS:
			fmt.Println("not found")
		case entity.ErrUnAvailableServer:
			fmt.Println("server Unavailable")
		case entity.ErrInternalError:
			fmt.Println("internal error")
		default:
			fmt.Println("unknown error")
		}
		return
	}

	fmt.Println("deleted")
}

func listDNS(ctx context.Context, dns *dns.Client) {
	res, err := dns.GetList(ctx)
	if err != nil {
		switch err {
		case entity.ErrUnAvailableServer:
			fmt.Println("server Unavailable")
		case entity.ErrInternalError:
			fmt.Println("internal error")
		default:
			fmt.Println("unknown error")
		}
		return
	}

	fmt.Printf("Result:\n\n")

	for _, v := range res {
		fmt.Println(v)
	}
}
