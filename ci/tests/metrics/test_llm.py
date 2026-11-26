#
# Copyright 2026 Canonical, Ltd.
#
"""Chat-model client tests."""

import sys
import unittest
from pathlib import Path
from typing import Any, Dict, List
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from metrics.llm import (  # noqa: E402
    PROVIDERS,
    ChatClient,
    LLMError,
    parse_json_reply,
    resolve_provider,
)


class FakeResponse:
    def __init__(self, status: int, body: Any, content_type: str = "application/json"):
        self.status_code = status
        self._body = body
        self.headers = {"content-type": content_type}
        self.text = body if isinstance(body, str) else str(body)

    def json(self) -> Any:
        if isinstance(self._body, str):
            raise ValueError("not JSON")
        return self._body


class FakeSession:
    def __init__(self, responses: List[FakeResponse]) -> None:
        self.responses = list(responses)
        self.calls: List[Dict[str, Any]] = []

    def post(self, url: str, **kwargs: Any) -> FakeResponse:
        self.calls.append({"url": url, **kwargs})
        return self.responses.pop(0)


def _reply(content: str) -> FakeResponse:
    return FakeResponse(200, {"choices": [{"message": {"content": content}}]})


class ResolveProviderTest(unittest.TestCase):
    def test_auto_picks_the_first_provider_with_a_token(self) -> None:
        self.assertEqual(resolve_provider("auto", {"GITHUB_TOKEN": "t"}).name, "github")
        self.assertEqual(
            resolve_provider(
                "auto", {"GITHUB_TOKEN": "t", "OPENROUTER_API_KEY": "k"}
            ).name,
            "openrouter",
        )
        self.assertEqual(resolve_provider("github", {}).name, "github")

    def test_missing_credentials_and_unknown_providers_fail(self) -> None:
        with self.assertRaisesRegex(LLMError, "no model credentials"):
            resolve_provider("auto", {"GITHUB_TOKEN": " "})
        with self.assertRaisesRegex(LLMError, "unknown provider"):
            resolve_provider("nope", {})
        with self.assertRaisesRegex(LLMError, "GITHUB_TOKEN"):
            ChatClient(PROVIDERS["github"], session=FakeSession([]), environ={})


class ChatClientTest(unittest.TestCase):
    def test_complete_posts_a_deterministic_json_request(self) -> None:
        session = FakeSession([_reply('{"decision": "unknown"}')])
        client = ChatClient(
            PROVIDERS["github"], session=session, environ={"GITHUB_TOKEN": "tok"}
        )

        reply = client.complete([{"role": "user", "content": "hi"}])

        self.assertEqual(reply, '{"decision": "unknown"}')
        call = session.calls[0]
        self.assertEqual(
            call["url"], "https://models.github.ai/inference/chat/completions"
        )
        self.assertEqual(call["headers"]["Authorization"], "Bearer tok")
        self.assertEqual(call["json"]["model"], "openai/gpt-4.1")
        self.assertEqual(call["json"]["temperature"], 0)
        self.assertEqual(call["json"]["response_format"], {"type": "json_object"})

    def test_rate_limits_and_server_errors_are_retried(self) -> None:
        limited = FakeResponse(429, "slow down")
        limited.headers["retry-after"] = "3"
        session = FakeSession([limited, FakeResponse(502, "bad gateway"), _reply("{}")])
        client = ChatClient(
            PROVIDERS["openrouter"],
            model="some/model",
            session=session,
            environ={"OPENROUTER_API_KEY": "k"},
        )

        with mock.patch("metrics.llm.time.sleep") as sleep:
            self.assertEqual(client.complete([]), "{}")

        self.assertEqual([c.args[0] for c in sleep.call_args_list], [3, 4])
        self.assertEqual(session.calls[0]["json"]["model"], "some/model")

    def test_client_errors_and_non_completions_raise(self) -> None:
        env = {"GITHUB_TOKEN": "t"}
        bad_request = ChatClient(
            PROVIDERS["github"],
            session=FakeSession([FakeResponse(400, "unknown model")]),
            environ=env,
        )
        plain_text = ChatClient(
            PROVIDERS["github"],
            session=FakeSession([FakeResponse(200, "OK", "text/plain")]),
            environ=env,
        )

        with self.assertRaisesRegex(LLMError, "HTTP 400.*unknown model"):
            bad_request.complete([])
        with self.assertRaisesRegex(LLMError, "no chat completion.*text/plain"):
            plain_text.complete([])


class ParseJsonReplyTest(unittest.TestCase):
    def test_extracts_the_object_from_fences_and_prose(self) -> None:
        self.assertEqual(parse_json_reply('{"a": 1}'), {"a": 1})
        self.assertEqual(parse_json_reply('```json\n{"a": 1}\n```'), {"a": 1})
        self.assertEqual(
            parse_json_reply('Sure:\n{"a": {"b": 2}} trailing'), {"a": {"b": 2}}
        )

    def test_rejects_replies_without_an_object(self) -> None:
        for text in ("no json here", '{"a": ', "[1, 2]"):
            with self.subTest(text=text), self.assertRaises(LLMError):
                parse_json_reply(text)


if __name__ == "__main__":
    unittest.main()
