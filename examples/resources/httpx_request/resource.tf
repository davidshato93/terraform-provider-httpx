resource "httpx_request" "example" {
  url    = "https://jsonplaceholder.typicode.com/posts/1"
  method = "GET"

  expect = [200]

  store_response_body = true
}

output "response" {
  value = httpx_request.example.response_body
}

