# pd

`pd` is a library and command line utility built on top of [go-pagerduty].

It currently provides some functions for finding conflicts that one or more
users may have in a set of oncall schedules.

To use the CLI, first [get an auth token][auth_setup], and then use like:

```bash
go build -o pd.bin ./cmd/
./pd.bin --auth-token "$YOUR_AUTH_TOKEN" -s"$SCHEDULE_1,$SCHEDULE_2,..."
```

[go-pagerduty]: https://github.com/PagerDuty/go-pagerduty
[auth_setup]: https://support.pagerduty.com/docs/generating-api-keys#generating-a-personal-rest-api-key
