package main

import (
	"flag"
	"fmt"
	"os"
	"tatlin_core/search"
)

// для сортировки имен в нужном порядке использовать флаги: -order=ascending, -order=descending, -order=random
// по дефолту: -order=descending

func main() {
	var order string
	flag.StringVar(&order, "order", "random", "flag for order of names in file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Waiting for file_name")
		os.Exit(1)
	}

	fileName := flag.Arg(0)

	fileStat, err := search.New(fileName)

	if err != nil {
		fmt.Printf("failed: %s\n", err.Error())
		os.Exit(1)
	}

	switch order {
	case "ascending":
		fileStat.DoAscendingOrder()
	case "descending":
		fileStat.DoDescendingOrder()
	case "random":
	default:
		fmt.Printf("wrong flag: %s\n", order)
	}

	fileStat.ShowStats()
}
