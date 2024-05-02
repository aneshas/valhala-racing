Host: "http://localhost:4000"

if #Meta.Environment.Type == "development" && #Meta.Environment.Cloud != "local" {
    Host: "http://xxx:4000"
}