package main

import (
  "context"
  "encoding/base64"
  "encoding/json"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "sort"
  "strconv"
  "time"
)

type ticket struct {
  Id          int `json:"id"`
  Subject     string `json:"subject"`
  Description string `json:"description"`
  Priority    string `json:"priority"`
  Status      string `json:"status"`
  RequesterId int `json:"requester_id"`
  Requester   string
  OrgId       int `json:"organization_id"`
  Org         string
  CreatedAt   time.Time `json:"created_at"`
}

type urgentTicket struct {
  Results     []struct {
    Id        int `json:"id"`
  } `json:"results"`
}

type zendeskTicketAPIResponse struct {
  TicketList  ticket `json:"ticket"`
  UserList    []user `json:"users"`
  OrgList     []organization `json:"organizations"`
}

type organization struct {
  Id          int `json:"id"`
  Name        string `json:"name"`
}

type user struct {
  Id          int `json:"id"`
  Name        string `json:"name"`
  Email       string `json:"email"`
}

func getTickets(ctx context.Context, targetTime time.Time, loadingChan chan string) ([]ticket, error) {
  var tickets []ticket
  numbers := getUrgentTicketNumbers(ctx, targetTime)
  for _,number := range numbers {
    loadingChan <- "loading"
    ticket, users, orgs := getTicketInfo(ctx, number)

    for _,v := range users {
      if v.Id == ticket.RequesterId {
        ticket.Requester = v.Name
        break
      } else {
        ticket.Requester = "Blank"
      }
    }

    for _,v := range orgs {
      if v.Id == ticket.OrgId {
        ticket.Org = v.Name
        break
      } else {
        ticket.Org = "Blank"
      }
    }
    tickets = append(tickets, ticket)
  }
  close(loadingChan)
  sort.Slice(tickets, func(i, j int) bool {
    return tickets[i].CreatedAt.After(tickets[j].CreatedAt)
  })
  return tickets, nil
}

func makeRequest(ctx context.Context, url string) ([]byte, error) {
  zendeskURL := os.Getenv("ZENDESK_URL")
	email := os.Getenv("ZENDESK_EMAIL")
	token := os.Getenv("ZENDESK_TOKEN")
  authToken := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + token))

  req, err := http.NewRequest(http.MethodGet, zendeskURL + url, nil)
  req.Header.Set("Authorization", "Basic " + authToken)
  resp, err := http.DefaultClient.Do(req.WithContext(ctx))
  if err != nil{
    log.Fatal(err)
  }
  body, err := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  return []byte(body), err
}

func getUrgentTicketNumbers(ctx context.Context, targetTime time.Time) ([]string) {
  v := url.Values{}
  v.Set("query", "tags:urgent_ticket created>=" + targetTime.Format("2006-01-02"))
  url := "/search.json?" + v.Encode()
  search, err := makeRequest(ctx, url)
  if err != nil {
    log.Fatal(err)
  }
  var z = new(urgentTicket)
  err = json.Unmarshal(search, &z)
  if err != nil {
    log.Fatal(err)
  }
  var results []string 
  for _,v := range z.Results {
    results = append(results, strconv.Itoa(v.Id))
  }
  return results
}

func getTicketInfo(ctx context.Context, number string) (ticket, []user, []organization) {
  // sideload the user and org information with the call to the ticket endpoint
  url := "/tickets/" + number + ".json?include=users,organizations"
  ticketbody, err := makeRequest(ctx, url)
  if err != nil {
    log.Fatal(err)
  }
  var ticket = new(zendeskTicketAPIResponse)
  err = json.Unmarshal(ticketbody, &ticket)
  if err != nil {
    log.Fatal(err)
  }

  return ticket.TicketList, ticket.UserList, ticket.OrgList
}
