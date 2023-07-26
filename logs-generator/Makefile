.PHONY: docker-build
docker-build:
	docker build -t logs-generator .

.PHONY: docker-run
docker-run:
	docker run \
      -it \
      --rm \
      --name logs-generator \
      logs-generator

.PHONY: docker-run
docker-run-with-args:
	docker run \
      -it \
      --rm \
      --name logs-generator \
      logs-generator \
      --lines=50 \
      --line-length-mu=30 \
      --line-length-sigma=1 \
      --seed=42


.PHONY: local-build
local-build:
	poetry install --no-interaction --no-ansi --no-root

.PHONY: local-run
local-run:
	poetry run generate

.PHONY: local-run-with-args
local-run-with-args:
	poetry run generate --lines=50 --line-length-mu=30 --line-length-sigma=1 --seed=42


gen-requirements:
	poetry export -f requirements.txt -o requirements.txt
	poetry export --dev -f requirements.txt -o requirements-dev.txt
