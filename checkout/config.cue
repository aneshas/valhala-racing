Host: "http://localhost:4000"

if #Meta.Environment.Name == "staging" {
    Host: "http://xxx:4000"
}
