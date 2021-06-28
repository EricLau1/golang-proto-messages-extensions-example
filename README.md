# Proto Messages (V1) with Field Extensions in Golang

## Build and Run

```bash
./build.sh

go run server/main.go

# open a second cmd
go run client/main.go

# open a third cmd
curl http://localhost:8889/messages/new -d '{"title":"Foo","body":"Bar","author":"anon"}'

```

- Service log output:

```json
{
        "original_name": "id",
        "protected": null,
        "editable": null,
        "custom_name": null
}
{
        "original_name": "title",
        "protected": true,
        "editable": true,
        "custom_name": "TitleC"
}
{
        "original_name": "body",
        "protected": false,
        "editable": false,
        "custom_name": "BodyC"
}
{
        "original_name": "author",
        "protected": null,
        "editable": null,
        "custom_name": null
}
```


## References:

- https://blog.golang.org/protobuf-apiv2
- https://github.com/golang/protobuf/issues/794