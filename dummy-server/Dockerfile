FROM python:3.11-slim-bullseye

ENV POETRY_HOME="/opt/poetry" \
    POETRY_VERSION=1.4.0
ENV PATH="$POETRY_HOME/bin:$PATH"

RUN apt-get update
RUN apt-get install -y curl

RUN curl -sSL https://install.python-poetry.org | python3 -
RUN poetry --version

# Copy only requirements to cache them in docker layer
WORKDIR /code
COPY poetry.lock pyproject.toml /code/

# Project initialization:
RUN poetry config virtualenvs.create false
RUN poetry install --no-interaction --no-ansi --no-root

# Creating folders, and files for a project:
COPY . /code

# https://www.uvicorn.org/settings/
CMD ["poetry", "run", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--reload", "--workers", "20"]
