#
# Copyright 2026 Canonical, Ltd.
#

import argparse
import asyncio
import logging
import signal
import sys
import urllib.parse
from typing import Optional, Tuple

LOG = logging.getLogger("http-proxy")

HOP_BY_HOP_HEADERS = {
    "proxy-connection",
    "proxy-authorization",
    "connection",
    "keep-alive",
    "upgrade",
}


def parse_host_port(target: str, default_port: int = 80) -> Tuple[str, int]:
    """Parses host and port from authority or target string, supporting IPv6 literals."""
    if target.startswith("["):
        host_part, _, rest = target[1:].partition("]:")
        port = int(rest) if rest.isdigit() else default_port
        return host_part, port
    if ":" in target:
        host, _, port_str = target.partition(":")
        port = int(port_str) if port_str.isdigit() else default_port
        return host, port
    return target, default_port


async def pipe(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    """Streams data from reader to writer until EOF or connection error."""
    try:
        while True:
            data = await reader.read(65536)
            if not data:
                break
            writer.write(data)
            await writer.drain()
    except (
        ConnectionResetError,
        BrokenPipeError,
        asyncio.IncompleteReadError,
        OSError,
    ):
        pass
    finally:
        try:
            if writer.can_write_eof():
                writer.write_eof()
            else:
                writer.close()
        except Exception:
            pass


async def handle_connect(
    target: str,
    client_reader: asyncio.StreamReader,
    client_writer: asyncio.StreamWriter,
) -> None:
    """Handles an HTTP CONNECT tunnel request."""
    # Consume remaining client headers up to empty line
    while True:
        line = await client_reader.readline()
        if not line or line in (b"\r\n", b"\n"):
            break

    host, port = parse_host_port(target, default_port=443)

    try:
        upstream_reader, upstream_writer = await asyncio.wait_for(
            asyncio.open_connection(host, port),
            timeout=15.0,
        )
    except Exception as e:
        LOG.debug("Failed to connect to upstream %s:%d: %s", host, port, e)
        try:
            client_writer.write(
                b"HTTP/1.1 502 Bad Gateway\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"
            )
            await client_writer.drain()
        except Exception:
            pass
        return

    try:
        client_writer.write(b"HTTP/1.1 200 Connection Established\r\n\r\n")
        await client_writer.drain()
    except Exception:
        upstream_writer.close()
        return

    pipe1 = asyncio.create_task(pipe(client_reader, upstream_writer))
    pipe2 = asyncio.create_task(pipe(upstream_reader, client_writer))
    try:
        done, pending = await asyncio.wait(
            [pipe1, pipe2],
            return_when=asyncio.FIRST_COMPLETED,
        )
        for task in pending:
            task.cancel()
            try:
                await task
            except (asyncio.CancelledError, Exception):
                pass
    finally:
        upstream_writer.close()
        try:
            await upstream_writer.wait_closed()
        except Exception:
            pass


async def handle_forward_http(
    method: str,
    target: str,
    client_reader: asyncio.StreamReader,
    client_writer: asyncio.StreamWriter,
) -> None:
    """Handles a plain HTTP forward request."""
    headers = []
    content_length = 0
    is_chunked = False
    host_header: Optional[str] = None

    while True:
        line = await client_reader.readline()
        if not line or line in (b"\r\n", b"\n"):
            break
        if len(headers) > 100:
            client_writer.write(
                b"HTTP/1.1 431 Request Header Fields Too Large\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"
            )
            await client_writer.drain()
            return
        decoded = line.decode("iso-8859-1", errors="replace").strip()
        if ":" in decoded:
            name, _, val = decoded.partition(":")
            name_strip = name.strip()
            val_strip = val.strip()
            name_lower = name_strip.lower()
            if name_lower == "host":
                host_header = val_strip
            elif name_lower == "content-length":
                try:
                    content_length = int(val_strip)
                except ValueError:
                    content_length = 0
            elif name_lower == "transfer-encoding" and "chunked" in val_strip.lower():
                is_chunked = True

            if name_lower not in HOP_BY_HOP_HEADERS:
                headers.append((name_strip, val_strip))

    split = urllib.parse.urlsplit(target)
    if split.scheme and split.hostname:
        upstream_host = split.hostname
        upstream_port = split.port or (443 if split.scheme == "https" else 80)
        path = split.path or "/"
        if split.query:
            path = f"{path}?{split.query}"
    elif host_header:
        upstream_host, upstream_port = parse_host_port(host_header, default_port=80)
        path = target
    else:
        client_writer.write(
            b"HTTP/1.1 400 Bad Request\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"
        )
        await client_writer.drain()
        return

    try:
        upstream_reader, upstream_writer = await asyncio.wait_for(
            asyncio.open_connection(upstream_host, upstream_port),
            timeout=15.0,
        )
    except Exception as e:
        LOG.debug("Failed to connect to %s:%d: %s", upstream_host, upstream_port, e)
        try:
            client_writer.write(
                b"HTTP/1.1 502 Bad Gateway\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"
            )
            await client_writer.drain()
        except Exception:
            pass
        return

    try:
        req_lines = [f"{method} {path} HTTP/1.1".encode("iso-8859-1")]
        has_host = False
        for h_name, h_val in headers:
            if h_name.lower() == "host":
                has_host = True
            req_lines.append(f"{h_name}: {h_val}".encode("iso-8859-1"))

        if not has_host:
            host_val = (
                upstream_host
                if upstream_port in (80, 443)
                else f"{upstream_host}:{upstream_port}"
            )
            req_lines.append(f"Host: {host_val}".encode("iso-8859-1"))

        req_lines.append(b"Connection: close")
        req_lines.append(b"")
        req_lines.append(b"")
        upstream_writer.write(b"\r\n".join(req_lines))
        await upstream_writer.drain()

        # Forward request body
        if content_length > 0:
            remaining = content_length
            while remaining > 0:
                chunk = await client_reader.read(min(remaining, 65536))
                if not chunk:
                    break
                upstream_writer.write(chunk)
                await upstream_writer.drain()
                remaining -= len(chunk)
        elif is_chunked:
            while True:
                chunk_header = await client_reader.readline()
                if not chunk_header:
                    break
                upstream_writer.write(chunk_header)
                await upstream_writer.drain()
                chunk_size_str = chunk_header.strip().split(b";")[0]
                try:
                    chunk_size = int(chunk_size_str, 16)
                except ValueError:
                    break
                if chunk_size == 0:
                    while True:
                        trailer = await client_reader.readline()
                        if not trailer:
                            break
                        upstream_writer.write(trailer)
                        await upstream_writer.drain()
                        if trailer in (b"\r\n", b"\n"):
                            break
                    break
                chunk_data = await client_reader.readexactly(chunk_size + 2)
                upstream_writer.write(chunk_data)
                await upstream_writer.drain()

        # Stream response back to client until EOF
        while True:
            chunk = await upstream_reader.read(65536)
            if not chunk:
                break
            client_writer.write(chunk)
            await client_writer.drain()
    except (
        ConnectionResetError,
        BrokenPipeError,
        asyncio.IncompleteReadError,
        OSError,
    ):
        pass
    finally:
        upstream_writer.close()
        try:
            await upstream_writer.wait_closed()
        except Exception:
            pass


async def handle_client(
    client_reader: asyncio.StreamReader,
    client_writer: asyncio.StreamWriter,
) -> None:
    """Dispatches incoming client connection to CONNECT or plain HTTP handler."""
    try:
        req_line = await client_reader.readline()
        if not req_line:
            return
        parts = req_line.decode("iso-8859-1", errors="replace").strip().split()
        if len(parts) != 3:
            client_writer.write(
                b"HTTP/1.1 400 Bad Request\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"
            )
            await client_writer.drain()
            return

        method, target, _ = parts[0].upper(), parts[1], parts[2]
        if method == "CONNECT":
            await handle_connect(target, client_reader, client_writer)
        else:
            await handle_forward_http(method, target, client_reader, client_writer)
    except (
        ConnectionResetError,
        BrokenPipeError,
        asyncio.IncompleteReadError,
        OSError,
    ):
        pass
    finally:
        client_writer.close()
        try:
            await client_writer.wait_closed()
        except Exception:
            pass


async def start_proxy_server(
    host: str = "0.0.0.0", port: int = 3128
) -> Tuple[asyncio.Server, int]:
    """Starts the proxy server and returns (server, bound_port)."""
    server = await asyncio.start_server(
        handle_client,
        host=host,
        port=port,
        reuse_address=True,
    )
    actual_port = server.sockets[0].getsockname()[1]
    return server, actual_port


async def main() -> None:
    """CLI entrypoint for running the proxy daemon."""
    parser = argparse.ArgumentParser(description="Local HTTP/HTTPS Forward Proxy")
    parser.add_argument("--bind", default="0.0.0.0", help="Address to bind on")
    parser.add_argument("--port", type=int, default=3128, help="Port to listen on")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s [%(levelname)s] %(message)s",
    )

    server, port = await start_proxy_server(host=args.bind, port=args.port)
    LOG.info("Proxy server listening on %s:%d", args.bind, port)

    loop = asyncio.get_running_loop()
    stop_event = asyncio.Event()

    for sig in (signal.SIGTERM, signal.SIGINT):
        try:
            loop.add_signal_handler(sig, stop_event.set)
        except (NotImplementedError, RuntimeError):
            pass

    async with server:
        stop_task = asyncio.create_task(stop_event.wait())
        server_task = asyncio.create_task(server.serve_forever())
        done, pending = await asyncio.wait(
            [stop_task, server_task], return_when=asyncio.FIRST_COMPLETED
        )
        for t in pending:
            t.cancel()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except (KeyboardInterrupt, SystemExit):
        sys.exit(0)
