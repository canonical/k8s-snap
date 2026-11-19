#
# Copyright 2026 Canonical, Ltd.
#

import asyncio
import importlib.util
import socket
import threading
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Generator, Tuple

import pytest
from test_util import config, tags

# Load the http_proxy module from the templates directory
spec = importlib.util.spec_from_file_location(
    "http_proxy", config.PROXY_DIR / "http_proxy.py"
)
http_proxy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(http_proxy)


class MockHTTPHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", "12")
        self.end_headers()
        self.wfile.write(b"HELLO_CLIENT")

    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        self.send_response(201)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        # Silence standard HTTP server request logs during test
        pass


@pytest.fixture
def mock_http_server() -> Generator[int, None, None]:
    server = HTTPServer(("127.0.0.1", 0), MockHTTPHandler)
    port = server.server_port
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    yield port
    server.shutdown()
    thread.join()


@pytest.fixture
def mock_echo_server() -> Generator[int, None, None]:
    async def echo_handler(reader: asyncio.StreamReader, writer: asyncio.StreamWriter):
        try:
            while True:
                data = await reader.read(4096)
                if not data:
                    break
                writer.write(data)
                await writer.drain()
        except (ConnectionResetError, BrokenPipeError, OSError):
            pass
        finally:
            writer.close()
            try:
                await writer.wait_closed()
            except Exception:
                pass

    loop = asyncio.new_event_loop()

    async def start():
        srv = await asyncio.start_server(echo_handler, "127.0.0.1", 0)
        p = srv.sockets[0].getsockname()[1]
        return srv, p

    server, port = loop.run_until_complete(start())
    thread = threading.Thread(target=loop.run_forever, daemon=True)
    thread.start()

    yield port

    async def stop():
        server.close()
        await server.wait_closed()

    asyncio.run_coroutine_threadsafe(stop(), loop).result(timeout=5)
    loop.call_soon_threadsafe(loop.stop)
    thread.join()


@pytest.fixture
def proxy_server() -> Generator[int, None, None]:
    loop = asyncio.new_event_loop()

    async def start():
        srv, p = await http_proxy.start_proxy_server(host="127.0.0.1", port=0)
        return srv, p

    server, port = loop.run_until_complete(start())
    thread = threading.Thread(target=loop.run_forever, daemon=True)
    thread.start()

    yield port

    async def stop():
        server.close()
        await server.wait_closed()

    asyncio.run_coroutine_threadsafe(stop(), loop).result(timeout=5)
    loop.call_soon_threadsafe(loop.stop)
    thread.join()


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_plain_http_get(proxy_server: int, mock_http_server: int):
    proxy_handler = urllib.request.ProxyHandler(
        {"http": f"http://127.0.0.1:{proxy_server}"}
    )
    opener = urllib.request.build_opener(proxy_handler)
    url = f"http://127.0.0.1:{mock_http_server}/test-get"
    resp = opener.open(url, timeout=5)
    assert resp.status == 200
    assert resp.read() == b"HELLO_CLIENT"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_plain_http_post(proxy_server: int, mock_http_server: int):
    proxy_handler = urllib.request.ProxyHandler(
        {"http": f"http://127.0.0.1:{proxy_server}"}
    )
    opener = urllib.request.build_opener(proxy_handler)
    url = f"http://127.0.0.1:{mock_http_server}/test-post"
    payload = b"sample_post_payload_bytes"
    req = urllib.request.Request(url, data=payload, method="POST")
    resp = opener.open(req, timeout=5)
    assert resp.status == 201
    assert resp.read() == payload


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_connect_tunnel(proxy_server: int, mock_echo_server: int):
    s = socket.create_connection(("127.0.0.1", proxy_server), timeout=5)
    connect_req = f"CONNECT 127.0.0.1:{mock_echo_server} HTTP/1.1\r\nHost: 127.0.0.1:{mock_echo_server}\r\n\r\n"
    s.sendall(connect_req.encode("iso-8859-1"))

    resp = b""
    while b"\r\n\r\n" not in resp:
        chunk = s.recv(1024)
        if not chunk:
            break
        resp += chunk

    assert b"200 Connection Established" in resp

    # Verify bidirectional payload tunneling
    test_msg = b"hello_tunnel_world"
    s.sendall(test_msg)
    echoed = s.recv(1024)
    assert echoed == test_msg

    s.close()


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_connection_refused_returns_502(proxy_server: int):
    # Bind an ephemeral port and close it immediately to guarantee it is unbound
    temp_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    temp_sock.bind(("127.0.0.1", 0))
    unused_port = temp_sock.getsockname()[1]
    temp_sock.close()

    proxy_handler = urllib.request.ProxyHandler(
        {"http": f"http://127.0.0.1:{proxy_server}"}
    )
    opener = urllib.request.build_opener(proxy_handler)
    url = f"http://127.0.0.1:{unused_port}/unreachable"

    with pytest.raises(urllib.error.HTTPError) as exc_info:
        opener.open(url, timeout=5)
    assert exc_info.value.code == 502


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_malformed_request_returns_400(proxy_server: int):
    s = socket.create_connection(("127.0.0.1", proxy_server), timeout=5)
    s.sendall(b"NOT_A_VALID_HTTP_REQUEST_LINE\r\n\r\n")
    resp = s.recv(1024)
    assert b"400 Bad Request" in resp
    s.close()


@pytest.mark.tags(tags.PULL_REQUEST)
def test_proxy_concurrent_requests(proxy_server: int, mock_http_server: int):
    def request_worker() -> Tuple[int, bytes]:
        proxy_handler = urllib.request.ProxyHandler(
            {"http": f"http://127.0.0.1:{proxy_server}"}
        )
        opener = urllib.request.build_opener(proxy_handler)
        url = f"http://127.0.0.1:{mock_http_server}/concurrent"
        resp = opener.open(url, timeout=5)
        return resp.status, resp.read()

    results = []
    threads = []
    for _ in range(8):
        t = threading.Thread(target=lambda: results.append(request_worker()))
        threads.append(t)
        t.start()

    for t in threads:
        t.join()

    assert len(results) == 8
    for status, body in results:
        assert status == 200
        assert body == b"HELLO_CLIENT"
