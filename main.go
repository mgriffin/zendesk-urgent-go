package main

import (
  "context"
  "fmt"
  "github.com/joho/godotenv"
  "log"
  "os"
  "time"
)

func main() {
	if len(os.Args) != 1 {
		usage()
	}
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
	if os.Getenv("ZENDESK_EMAIL") == "" {
		usage()
	}
	if os.Getenv("ZENDESK_TOKEN") == "" {
		usage()
	}

  ctx := context.Background()
  lastWeek := time.Now().AddDate(0, 0, -7)

  var loadingChan = make(chan string)
  tickFunc := func() ([]ticket, error) {
    return getTickets(ctx, lastWeek, loadingChan)
  }
  err = runUI(loadingChan, tickFunc)
  if err != nil {
    fmt.Fprintf(os.Stderr, "got an error:\n%s", err.Error())
    os.Exit(1)
  }
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: zd-urgent`)
	os.Exit(2)
}
