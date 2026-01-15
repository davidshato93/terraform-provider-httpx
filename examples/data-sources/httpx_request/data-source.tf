data "httpx_request" "example" {
  url    = "https://jsonplaceholder.typicode.com/posts/1"
  method = "GET"

  expect = [200]
}

output "post" {
  value = jsondecode(data.httpx_request.example.response_body)
}

