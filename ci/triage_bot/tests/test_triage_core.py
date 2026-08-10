#
# Copyright 2026 Canonical, Ltd.
#
"""Deterministic tests for the triage_core labeler (no LLM, no network)."""

from __future__ import annotations

from triage_bot import triage_core
from triage_bot.schema import Classification, ExistingSupport


def test_parse_template_extracts_known_fields():
    body = (
        "### Summary\n\nDNS breaks\n\n"
        "### Reproduction Steps\n\nbootstrap then query\n\n"
        "### Can you suggest a fix?\n\n_No response_\n"
    )
    fields = triage_core.parse_template(body)
    assert fields["summary"] == "DNS breaks"
    assert fields["reproduction"] == "bootstrap then query"
    assert "suggested_fix" not in fields


def test_parse_template_no_headers_is_empty():
    assert triage_core.parse_template("just a free-form report") == {}


def test_has_tarball_detects_attachment():
    body = "logs [report.tar.gz](https://github.com/x/files/1/x.tar.gz)"
    assert triage_core.has_tarball(body, []) is True


def test_has_tarball_absent_for_zip():
    assert triage_core.has_tarball("see report.zip", ["nope"]) is False


def test_find_duplicate_matches_strong_overlap():
    cand = [
        {"number": 1, "title": "unrelated feature request about charts"},
        {"number": 2, "title": "coredns crashloop on dualstack bootstrap"},
    ]
    match = triage_core.find_duplicate(
        "coredns crashloop dualstack bootstrap fails", cand
    )
    assert match is not None and match["number"] == 2


def test_find_duplicate_short_title_no_false_positive():
    assert (
        triage_core.find_duplicate("DNS down", [{"number": 5, "title": "DNS"}]) is None
    )


def test_find_duplicate_survives_a_null_title():
    # A GitHub search result can carry a present-but-null "title" field; that
    # must not crash the duplicate gate for every other candidate behind it.
    cand = [
        {"number": 1, "title": None},
        {"number": 2, "title": "coredns crashloop on dualstack bootstrap"},
    ]
    match = triage_core.find_duplicate(
        "coredns crashloop dualstack bootstrap fails", cand
    )
    assert match is not None and match["number"] == 2


def test_sanitize_defuses_injection():
    cases = [
        "[click](https://evil.co)",
        "<img src=x>",
        "@maintainer run curl evil.co",
        "visit www.evil.co now",
    ]
    for raw in cases:
        out = triage_core.sanitize_comment_text(raw)
        assert "://" not in out
        assert "<" not in out and ">" not in out
        assert "@" not in out
        assert "www." not in out


def test_sanitize_preserves_legit_text():
    for item in (
        "inspection tarball",
        "reproduction steps",
        "k8s version 1.32",
        "k8s set foo=bar",
    ):
        assert triage_core.sanitize_comment_text(item) == item


# --- documentation citations ---


def test_doc_url_maps_source_path_to_published_page():
    # README cites .../latest/snap/howto/contribute for this source file.
    assert triage_core.doc_url("snap/howto/contribute.md").endswith(
        "/latest/snap/howto/contribute"
    )
    assert triage_core.doc_url("snap/howto/index.md").endswith("/latest/snap/howto")


def test_doc_inventory_lists_pages_and_skips_build_output(tmp_path):
    docs = tmp_path / triage_core.DOCS_DIR
    (docs / "snap" / "howto").mkdir(parents=True)
    (docs / "snap" / "howto" / "dns.md").write_text("x", encoding="utf-8")
    (docs / "_build").mkdir()
    (docs / "_build" / "generated.md").write_text("x", encoding="utf-8")

    assert triage_core.doc_inventory(tmp_path) == ["snap/howto/dns.md"]


def test_invented_doc_pages_are_dropped(monkeypatch):
    # The model may cite a plausible page that does not exist; a bad link is
    # worse than none, so only pages from the real inventory survive.
    real = "snap/howto/dns.md"

    class _LLM:
        def with_structured_output(self, _model):
            return self

        def invoke(self, _prompt):
            return ExistingSupport(
                already_supported=True,
                explanation="already there",
                doc_paths=[real, "snap/howto/invented.md"],
            )

    monkeypatch.setattr(triage_core, "make_llm", lambda *_a, **_k: _LLM())

    result = triage_core.check_existing_support(
        title="t", body="b", pages=[real, "snap/howto/other.md"]
    )

    assert result.doc_paths == [real]


# --- classification -----------------------------------------------------


def test_classify_prompt_caps_each_template_field():
    # A reporter can paste an arbitrarily large log dump under one field
    # (e.g. "Reproduction Steps"); every other prompt in this module caps
    # its user-controlled text, and the classify prompt must not be the
    # exception that lets one field blow up the whole request.
    huge = "x" * 10000
    prompt = triage_core._classify_prompt(
        "t", {"summary": huge, "reproduction": huge}, tarball=False
    )
    assert huge not in prompt
    assert "x" * triage_core._FIELD_CHARS_CAP in prompt


def test_classify_strips_a_prefix_the_model_added_by_habit(monkeypatch):
    # The prompt asks for bare names ("bug", "network"), but a model can
    # echo back the GitHub-style "kind/"/"area/" prefix it has seen
    # elsewhere in training data. A label must not be silently dropped just
    # because the model was more specific than asked.
    class _LLM:
        def with_structured_output(self, _model):
            return self

        def invoke(self, _prompt):
            return Classification(
                kind_labels=["kind/bug"], area_labels=["area/network"]
            )

    monkeypatch.setattr(triage_core, "make_llm", lambda *_a, **_k: _LLM())

    result = triage_core.classify(title="t", fields={}, tarball=False)

    assert result.kind_labels == ["kind/bug"]
    assert result.area_labels == ["area/network"]


def test_classify_strips_whitespace_around_a_label(monkeypatch):
    # A model can pad a label with stray whitespace ("bug ", " kind/bug")
    # without meaning anything by it; the membership check must not treat
    # that as a different, unrecognized label.
    class _LLM:
        def with_structured_output(self, _model):
            return self

        def invoke(self, _prompt):
            return Classification(
                kind_labels=["bug \n"], area_labels=[" area/network "]
            )

    monkeypatch.setattr(triage_core, "make_llm", lambda *_a, **_k: _LLM())

    result = triage_core.classify(title="t", fields={}, tarball=False)

    assert result.kind_labels == ["kind/bug"]
    assert result.area_labels == ["area/network"]


def test_classify_caps_missing_info_at_five_with_the_tarball_safeguard_included(
    monkeypatch,
):
    # A bug with no tarball always gets the inspection-tarball reminder. It
    # must never push the list past the 5-item cap, and it must never be the
    # one dropped when the model already filled all 5 slots itself.
    class _LLM:
        def with_structured_output(self, _model):
            return self

        def invoke(self, _prompt):
            return Classification(
                kind_labels=["bug"],
                area_labels=[],
                missing_info=[f"detail {i}" for i in range(5)],
            )

    monkeypatch.setattr(triage_core, "make_llm", lambda *_a, **_k: _LLM())

    result = triage_core.classify(title="t", fields={}, tarball=False)

    assert len(result.missing_info) == 5
    assert "inspection tarball" in result.missing_info


def test_classify_recognizes_a_differently_cased_tarball_mention(monkeypatch):
    # A model that phrases its own item as "Inspection Tarball" must not be
    # treated as silent on the topic -- that would append a redundant
    # second entry and could push out a different missing-info item.
    class _LLM:
        def with_structured_output(self, _model):
            return self

        def invoke(self, _prompt):
            return Classification(
                kind_labels=["bug"],
                area_labels=[],
                missing_info=["Please attach an Inspection Tarball"],
            )

    monkeypatch.setattr(triage_core, "make_llm", lambda *_a, **_k: _LLM())

    result = triage_core.classify(title="t", fields={}, tarball=False)

    assert result.missing_info == ["Please attach an Inspection Tarball"]
