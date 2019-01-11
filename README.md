### Analyse urgent Zendesk tickets

Get a list of tickets that have paged engineers for the last week.
From this list, you can see a view of the initial message and from there might be able to figure out if the ticket really was urgent.

Massive inspiration taken from https://github.com/WillAbides/discuss

### How do I run it?

Make sure you have a `.env` file with the following fields:

```
ZENDESK_EMAIL="your login email"
ZENDESK_TOKEN="your api token"
ZENDESK_URL="your zendesk URL"
```

Then run it with:

```
go run *.go
```

Exit it by hitting `CTRL+c`.

### TODO

- Add a way to flag tickets that are not considered urgent and create a report
- Add a custom timespan
- Run the Zendesk API calls in parallel to speed up retrieval
- Simplify it all 😀
