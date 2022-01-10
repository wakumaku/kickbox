# kickbox API client

https://docs.kickbox.com/docs/using-the-api

API Client

```
    client, err := kickbox.New("apikey")
    iferr ...
    client.Verify(context.TODO(), "example@email.com")
```

Sandbox Client

```
    client := kickbox.NewSandbox()
    iferr ...
    client.Verify(context.TODO(), "example@email.com")
```

In your application

```
    var emailVerifier   kickbox.Verifier

    if env == dev {
        emailVerifier = kickbox.NewSandbox()
    } else {
        emailVerifier = kickbox.New("apikey")
    }

```






git remote add origin git@github.com:wakumaku/go-kickbox.git
git branch -M main
git push -u origin main


