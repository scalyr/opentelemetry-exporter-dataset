# Dummy Server

Dummy implementation of server with the API for `addEvents` API call.


## Usage

1. Start server - `make run` - and it starts listening on the port `8000`.
2. Use `make get-stats` - to get statistics
3. Use `make set-state-retry-later` - to make `addEvents` API respond with status code `429` and `Retry-After` HTTP header set to 30.
4. Observe how OTel behaves
5. Use `make set-state-ok` - resets all the previous settings.


## Support for HTTPS

* Install mkcert - `brew install mksert` - https://github.com/FiloSottile/mkcert
* Generate certificate
  * Install CA - `mkcert -install`
  * Generate keys - `mkcert localhost 127.0.0.1 ::1 host.docker.internal`
* Use `make run-https` - https://www.uvicorn.org/deployment/#running-with-https


## Other Commands
* `make set-http-code CODE=530` - API will respond with HTTP status code 530
* `make reset-http-code` - API will respond with natural HTTP status code
* `make set-api-status STATUS="hello" - JSON payload will have have attribute `status` set to `hello`
* `make reset-api-status` - JSON payload will have natural value of attribute `status`
* `make set-api-message MESSAGE="world" - JSON payload will have have attribute `message ` set to `world`
* `make reset-api-message` - JSON payload will have natural value of attribute `message `
* `make set-delay-ms DELAY=3000` - sets delay to 3000ms before response is returned
* `make reset-delay-ms` - there will be no artificial delay before response is returned
* `make set-retry-after-s RETRY_AFTER=30` - sets header `Retry-After` to 30s for all non 200 http codes
* `make reset-retry-after-s` - there will be no header `Reset-After`
* `make set-state-ok` - reset all options
* `make set-state-retry-later` - sets HTTP code to 429 and Retry-After to 30
* `make params` - returns the current configuration
