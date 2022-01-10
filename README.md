# kickbox API client

https://docs.kickbox.com/docs/using-the-api

## API Client

Wil make real http requests against kickbox services. You can use a `test_apikey` to enable the sandbox mode.

### Simple single verification:

```golang
    client, err := kickbox.New("apikey")
    if err != nil {
        return
    }

    stats, response, err := client.Verify(context.TODO(), "example@email.com")
```

With Options:


```golang
    client, err := kickbox.New("apikey",
        kickbox.OverrideBaseURL("http://mock.server.com"),
        kickbox.MaxConcurrentConnections(100), // Default: is 25
        kickbox.CustomRateLimiter(rate.NewLimiter(rate.Every(50 / time.Second), 1)), // Default: 8000 per minute
        kickbox.CustomHTTPClient(&http.Client{}),
    )
    if err != nil {
        return 
    }

    stats, response, err := client.Verify(context.TODO(), "example@email.com")
```

### Batch Verification:

```golang
    client, _ := kickbox.New("apikey")

    emailsFile, err := os.Open(csvFilePath)
    if err != nil {
        return 
    }

    resp, _ := client.VerifyBatch(
        context.TODO(),
        emailsFile,
    )
```

With options:

```golang
    client, _ := kickbox.New("apikey")

    emailsFile, err := os.Open(csvFilePath)
    if err != nil {
        return 
    }

    resp, err := c.VerifyBatch(
        context.TODO(),
        emailsFile,
        kickbox.Filename("my_results.csv"),
        kickbox.Callback("http://notify.result.here"),
    )
```

### Batch Status Check:

```golang
    client, _ := kickbox.New("apikey")

    resp, err := client.VerifyBatchCheck(context.TODO(), "batch_ID")
```


## Local Sandbox Client

For testing and CI/CD environments, without external calls.

```golang
    client := kickbox.NewSandbox()

    stats, response, err := client.Verify(context.TODO(), "example@email.com")
```
