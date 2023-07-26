.PHONY: docker-build
docker-build:
	docker build -t dataset-dummy-server .

.PHONY: docker-run
docker-run:
	docker run \
      -p 8000:8000 \
      -it \
      --rm \
      --name dataset-dummy-server \
      dataset-dummy-server

local-build:
	poetry install --no-interaction --no-ansi --no-root
local-run: run

run: run-http
run-http:
	poetry run uvicorn main:app --port 8000 --host 0.0.0.0 --reload

run-https:
	poetry run uvicorn main:app --port 8443 --host 0.0.0.0 --ssl-keyfile=./localhost+3-key.pem --ssl-certfile=./localhost+3.pem --reload

set-http-code:
	curl -X GET 'http://127.0.0.1:8000/params/?http_code='$(CODE)

reset-http-code:
	curl -X GET 'http://127.0.0.1:8000/params/?http_code=0'

set-api-status:
	curl -X GET 'http://127.0.0.1:8000/params/?api_status='$(STATUS)

reset-api-status:
	curl -X GET 'http://127.0.0.1:8000/params/?api_status='

set-api-message:
	curl -X GET 'http://127.0.0.1:8000/params/?api_message='$(MESSAGE)

reset-api-message:
	curl -X GET 'http://127.0.0.1:8000/params/?api_message='

set-delay-ms:
	curl -X GET 'http://127.0.0.1:8000/params/?delay_ms='$(DELAY)

reset-delay-ms:
	curl -X GET 'http://127.0.0.1:8000/params/?delay_ms=0'

set-retry-after-s:
	curl -X GET 'http://127.0.0.1:8000/params/?retry_after_s='$(RETRY_AFTER)

reset-retry-after-s:
	curl -X GET 'http://127.0.0.1:8000/params/?retry_after_s=0'

params:
	curl -X GET 'http://127.0.0.1:8000/params/'

get-stats: stats
stats:
	curl -X GET 'http://127.0.0.1:8000/stats'

coverage: tests
tests:
	poetry run pytest test.py

set-state-ok: \
	reset-api-message \
	reset-api-status \
	reset-delay-ms \
	reset-http-code \
	reset-retry-after-s

set-state-retry-later: set-state-ok
	$(MAKE) set-http-code CODE=429
	$(MAKE) set-retry-after-s RETRY_AFTER=30

gen-requirements:
	poetry export -f requirements.txt -o requirements.txt
	poetry export --dev -f requirements.txt -o requirements-dev.txt
