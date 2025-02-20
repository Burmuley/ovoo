POST http://localhost:8808/api/v1/praddrs
{
  "email": "{{prot_addr_input}}",
  "metadata": {
    "comment": "Test protected address",
    "service_name": "shmoogle.com"
  }
}
HTTP 201
[Captures]
praddr1_id: jsonpath "$['id']"

POST http://localhost:8808/api/v1/aliases
{
  "protected_address_id": "{{praddr1_id}}",
  "metadata": {
    "comment": "test alias comment",
    "service_name": "test alias service name"
  }
}
HTTP 201
[Captures]
alias1_id: jsonpath "$['id']"
alias1_email: jsonpath "$['email']"

GET http://localhost:8808/api/v1/aliases/{{alias1_id}}
HTTP 200
[Asserts]
jsonpath "$['id']" == "{{alias1_id}}"

POST http://localhost:8808/api/v1/chains
{
    "from_email": "{{ext_sender_input}}",
    "to_email": "{{alias1_email}}"
}
HTTP 201
