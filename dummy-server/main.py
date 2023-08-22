import asyncio
from collections import defaultdict
import copy
import gzip
import json
import logging
import random
from struct import unpack
import time
from typing import Any, Callable, Dict, List, Optional, Union
import zlib

from fastapi import FastAPI
from fastapi import Header
from fastapi import Request
from fastapi import Response
from fastapi import status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi.routing import APIRoute
from pydantic import BaseModel
from pyhocon import ConfigParser
from starlette.exceptions import HTTPException as StarletteHTTPException

logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)-9s %(asctime)s %(filename)s:%(lineno)d %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

requests_stats = defaultdict(
    lambda: defaultdict(int)
)  # type: Dict[str, Dict[str, int]]
events_stats = defaultdict(lambda: defaultdict(int))  # type: Dict[str, Dict[str, int]]

events_stats["total"]["total"] = 0
events_stats["total"]["success"] = 0

MSG_PAT = b"message:`s"
MSG_LEN = len(MSG_PAT)
MSG_BEG = '"message":"'.encode("utf-8")


class Parameters(object):
    http_code: Optional[int] = None
    api_message: Optional[str] = None
    api_status: Optional[str] = None
    delay_ms: Optional[int] = None
    retry_after_s: Optional[int] = None

    def __str__(self) -> str:
        return repr(self)

    def __repr__(self) -> str:
        return json.dumps(
            {
                "http_code": self.http_code,
                "api_message": self.api_message,
                "api_status": self.api_status,
                "delay_ms": self.delay_ms,
                "retry_after_s": self.retry_after_s,
            }
        )


response_params = Parameters()


def _fix_agent(inp: bytes) -> bytes:
    # fix messages
    fixing_start = time.process_time()
    buffer = b""
    pos = 0
    while True:
        n = inp.find(MSG_PAT, pos)
        if n >= 0:
            before = inp[pos:n]
            buffer += before
            buffer += MSG_BEG

            # detect how long is the message
            raw_len = inp[n + MSG_LEN : n + MSG_LEN + 4]
            bytes_len = unpack(">l", raw_len)
            buffer += (
                inp[n + MSG_LEN + 4 : n + MSG_LEN + 4 + bytes_len[0]]
                .replace(b"\\", b" ")  # remove all backslashes
                .replace(b'"', b'\\"')  # fix quotation marks
                .replace(b"\n", b" ")  # remove new lines
            )
            # close message
            buffer += b'"'
            # move position
            pos = n + MSG_LEN + 4 + bytes_len[0]
        else:
            break
    buffer += inp[pos:]
    try:
        # logging.warning("AAA - Buffer generated: %d", len(buffer))
        buffer_str = buffer.decode("utf-8")
        # logging.warning("AAA - Buffer converted: %d", len(buffer_str))
        hocon_parsed = ConfigParser.parse(buffer_str)
        # logging.warning("AAA - Hocon parsed")
        config = hocon_parsed.as_plain_ordered_dict()
        # logging.warning("AAA - Hocon To Dict")
        output_str = json.dumps(config)
        # logging.warning("AAA - JSON - dumps")
        logging.info(
            "Body fixed in %f: %d => %d",
            time.process_time() - fixing_start,
            len(inp),
            len(output_str),
        )
        return output_str.encode("utf-8")
    except Exception as e:
        logging.error(buffer)
        logging.error(e, exc_info=True)
    return inp


def fix_json(inp: bytes) -> bytes:
    """
    We just want to extract number of events. Payload send by agent is not valid json.
    It's also not valid hocon file. So let's do some string replacements and hope that
    it will end up as valid hocon string.
    """
    try:
        # https://docs.python.org/3/howto/unicode.html#the-string-type
        # there are some non-utf characters
        inp_str = inp.decode("utf-8")
        _ = json.loads(inp_str)
    except Exception as e1:
        logging.warning("It's not valid input: %s: %s", inp[:200], e1)
        return _fix_agent(inp)
    return inp


# https://fastapi.tiangolo.com/advanced/custom-request-and-route/#create-a-custom-gziprequest-class
class CompressedRequest(Request):
    async def body(self) -> bytes:
        if not hasattr(self, "_body"):
            body = await super().body()
            content_encoding = self.headers.getlist("Content-Encoding")
            logging.debug("Encoding: %s", content_encoding)
            if "deflate" in content_encoding:
                body = zlib.decompress(body)
                body = fix_json(body)
            elif "gzip" in content_encoding:
                body = gzip.decompress(body)
            self._body = body
        return self._body


class CompressedRoute(APIRoute):
    def get_route_handler(self) -> Callable:
        original_route_handler = super().get_route_handler()

        async def custom_route_handler(request: Request) -> Response:
            request = CompressedRequest(request.scope, request.receive)
            return await original_route_handler(request)

        return custom_route_handler


app = FastAPI(debug=True)
app.router.route_class = CompressedRoute


# better error message when parsing payload
@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    exc_str = f"{exc}".replace("\n", " ").replace("   ", " ")
    logging.error(f"{request.__dict__}: {exc_str}")
    logging.error(f"Invalid request body: {exc.body}")
    content = {"status": "10422", "message": exc_str, "data": None}
    return JSONResponse(
        content=content, status_code=status.HTTP_422_UNPROCESSABLE_ENTITY
    )


@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(request, exc):
    global requests_stats
    requests_stats[request.headers.get("User-Agent")][
        f"R:{exc.status_code}:{exc.detail}"
    ] += 1
    exc_str = f"{exc.detail}: {exc.status_code}"
    logging.error(f"{request}: {exc_str}")
    content = {"status": "error", "message": exc_str, "detail": str(exc.detail)}
    return JSONResponse(content=content, status_code=exc.status_code)


class AddEventsResponse(BaseModel):
    status: str
    message: str
    bytesCharged: int


class Event(BaseModel):
    sev: Optional[int]
    ts: Optional[str]
    log: Optional[str]
    thread: Optional[str]
    attrs: Dict[str, Any]


class Thread(BaseModel):
    id: str
    name: str


class Log(BaseModel):
    id: str
    attrs: Dict[str, Any]


class SessionInfo(BaseModel):
    serverType: Optional[str]
    serverId: Optional[str]
    region: Optional[str]


class AddEventsRequestParams(BaseModel):
    token: str
    session: str
    sessionInfo: Optional[SessionInfo]
    events: Optional[List[Event]]
    threads: Optional[List[Thread]]
    logs: Optional[List[Log]]


@app.get("/")
async def root():
    return {"message": "Hello World"}


@app.get("/stats")
async def get_stats():
    logging.info("Stats - events: %r", dict(events_stats))
    logging.info("Stats - requests: %r", dict(requests_stats))
    return {
        "events": dict(events_stats),
        "requests": dict(requests_stats),
    }


@app.get("/params/")
async def params(
    http_code: Optional[int] = None,
    api_message: Optional[str] = None,
    api_status: Optional[str] = None,
    delay_ms: Optional[int] = None,
    retry_after_s: Optional[int] = None,
):
    """
    Set properties that should API return.

    :param http_code:   http code
    :param api_message: message returned by API
    :param api_status:  status returned by API
    :param delay_ms:    delay in MS
    :param retry_after_s: header value for "Retry-After"
    :return:
    """
    global response_params

    before = copy.deepcopy(response_params)

    if http_code is not None:
        response_params.http_code = http_code
    if api_message is not None:
        response_params.api_message = api_message
    if api_status is not None:
        response_params.api_status = api_status
    if delay_ms is not None:
        response_params.delay_ms = delay_ms
    if retry_after_s is not None:
        response_params.retry_after_s = retry_after_s

    logging.info("Setting: %s => %s", before, response_params)
    return {
        "old": before,
        "new": response_params,
    }


@app.post("/api/addEvents", response_model=AddEventsResponse)
@app.post("/addEvents", response_model=AddEventsResponse)
async def add_events(
    payload: AddEventsRequestParams,
    user_agent: Union[str, None] = Header(default=None),  # noqa: B008
):
    global response_params
    global events_stats
    global requests_stats

    num_events = len(payload.events) if payload.events else 0
    user_agent = user_agent or ""

    logging.info(
        "agent: %s; events: %d; total: %d; success: %d; params: %s; message: %s",
        user_agent,
        num_events,
        events_stats[user_agent]["total"],
        events_stats[user_agent]["success"],
        response_params,
        payload.events[0].attrs.get("message") if payload.events else "",
    )

    if response_params.delay_ms:
        logging.info("Sleeping for %d ms", response_params.delay_ms)
        await asyncio.sleep(response_params.delay_ms / 1000.0)

    events_stats[user_agent]["total"] += num_events
    events_stats[user_agent]["success"] += 0
    if not events_stats[user_agent]["first"]:
        events_stats[user_agent]["first"] = time.time_ns()
    events_stats[user_agent]["last"] = time.time_ns()

    events_stats["total"]["total"] += num_events
    events_stats["total"]["success"] += 0
    if not events_stats["total"]["first"]:
        events_stats["total"]["first"] = time.time_ns()
    events_stats["total"]["last"] = time.time_ns()

    if response_params.http_code is None or response_params.http_code <= 201:
        requests_stats[user_agent]["R:200:Ok"] += 1
        requests_stats["total"]["R:200:Ok"] += 1
        events_stats[user_agent]["success"] += num_events
        events_stats["total"]["success"] += num_events
        return AddEventsResponse(
            status=response_params.api_status or "success",
            message=response_params.api_message or "success",
            bytesCharged=random.randint(10, 1000),
        )
    else:
        requests_stats[user_agent][f"F:{response_params.http_code}:Set"] += 1
        requests_stats["total"][f"F:{response_params.http_code}:Set"] += 1
        response = JSONResponse(
            content={
                "status": response_params.api_status or "error",
                "message": response_params.api_message or "error",
            },
            status_code=response_params.http_code,
        )
        if response_params.retry_after_s:
            response.headers.append("Retry-After", str(response_params.retry_after_s))

        return response
