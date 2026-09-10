#
# Copyright 2026 Canonical, Ltd.
#
"""Tests for the Stage B rule pack.

The emphasis here is on the rule pack *failing loudly*. A silently broken rule
under-classifies without telling anyone, which is precisely the blind spot this
system is meant to remove, so most of these tests assert that bad input raises
rather than that good input works.
"""

import pytest

from metrics.models import ClassifiedBy, FailureClass, JobFailure
from metrics.rules import RulePackError, classify_with_rules, load_rules

SIG_A = "c55cefd2ca3f28e7"
SIG_B = "38c23bd3a1bab957"


def write_pack(tmp_path, body):
    """Write a rule pack, substituting %(a)s / %(b)s with sample signatures.

    Plain strings rather than f-strings on purpose: pycodestyle parses
    f-string bodies on Python 3.12 and reports style errors for the YAML
    inside them.
    """
    path = tmp_path / "failure_rules.yaml"
    path.write_text(body % {"a": SIG_A, "b": SIG_B})
    return path


def make_job(**kwargs):
    defaults = dict(
        job_id=1,
        run_id=2,
        attempt=1,
        job_name="Integration (ubuntu:24.04, amd64, 1.35-classic/edge) / t.py::x",
        html_url="https://example.invalid/1",
    )
    defaults.update(kwargs)
    return JobFailure(**defaults)


class TestLoading:
    def test_missing_file_is_not_an_error(self, tmp_path):
        # Before the pack is seeded, classifying nothing is the honest answer.
        pack = load_rules(tmp_path / "nope.yaml")
        assert pack.rules == []
        assert pack.version == 0

    def test_loads_a_valid_pack(self, tmp_path):
        path = write_pack(
            tmp_path,
            """
version: 3
rules:
  - id: kube-proxy-inactive
    signature_id: %(a)s
    class: product.bug
    subclass: service_not_active
    owner: k8s-team
    issue: 1234
""",
        )
        pack = load_rules(path)
        assert pack.version == 3
        assert len(pack.rules) == 1
        rule = pack.rules[0]
        assert rule.signature_ids == [SIG_A]
        assert rule.failure_class == FailureClass.PRODUCT_BUG.value
        assert rule.owner == "k8s-team"
        assert rule.issue == 1234

    def test_rejects_unknown_class(self, tmp_path):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    signature_id: %(a)s\n"
            "    class: product.oops\n",
        )
        with pytest.raises(RulePackError, match="unknown class"):
            load_rules(path)

    def test_rejects_subclass_that_does_not_belong_to_its_class(self, tmp_path):
        # 'lxd_setup' is real, but it belongs to infra.provisioning.
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    signature_id: %(a)s\n"
            "    class: product.bug\n    subclass: lxd_setup\n",
        )
        with pytest.raises(RulePackError, match="not valid for"):
            load_rules(path)

    def test_rejects_a_rule_with_no_criteria(self, tmp_path):
        # Such a rule matches every failure and would poison the whole dataset.
        path = write_pack(
            tmp_path, "version: 1\nrules:\n  - id: r\n    class: product.bug\n"
        )
        with pytest.raises(RulePackError, match="no criteria"):
            load_rules(path)

    def test_rejects_a_malformed_signature_id(self, tmp_path):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    signature_id: NOTAHASH\n"
            "    class: product.bug\n",
        )
        with pytest.raises(RulePackError, match="hex digest"):
            load_rules(path)

    def test_rejects_duplicate_rule_ids(self, tmp_path):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n"
            "  - id: dupe\n    signature_id: %(a)s\n    class: product.bug\n"
            "  - id: dupe\n    signature_id: %(b)s\n    class: test.bug\n",
        )
        with pytest.raises(RulePackError, match="duplicate rule id"):
            load_rules(path)

    def test_rejects_an_unmatchable_job_field(self, tmp_path):
        # Matching on a classification field would make results order-dependent.
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    class: product.bug\n"
            "    job:\n      failure_class: product.bug\n",
        )
        with pytest.raises(RulePackError, match="not matchable"):
            load_rules(path)

    def test_rejects_an_invalid_regex(self, tmp_path):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    class: product.bug\n"
            "    excerpt_pattern: '([unclosed'\n",
        )
        with pytest.raises(RulePackError, match="invalid regex"):
            load_rules(path)

    def test_rejects_confidence_out_of_range(self, tmp_path):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n  - id: r\n    signature_id: %(a)s\n"
            "    class: product.bug\n    confidence: 1.5\n",
        )
        with pytest.raises(RulePackError, match="confidence"):
            load_rules(path)

    def test_rejects_a_missing_version(self, tmp_path):
        path = write_pack(
            tmp_path,
            "rules:\n  - id: r\n    signature_id: %(a)s\n    class: product.bug\n",
        )
        with pytest.raises(RulePackError, match="version"):
            load_rules(path)

    def test_warns_when_a_pinned_rule_is_unreachable(self, tmp_path, caplog):
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n"
            "  - id: broad\n    signature_id: [%(a)s, %(b)s]\n"
            "    class: infra.runner\n"
            "  - id: narrow\n    signature_id: %(a)s\n    class: product.bug\n",
        )
        with caplog.at_level("WARNING"):
            load_rules(path)
        assert "unreachable" in caplog.text

    def test_does_not_warn_when_the_earlier_rule_adds_criteria(self, tmp_path, caplog):
        # Splitting one signature by context is deliberate: the same error text
        # means provisioning on Multipass and a product bug on LXD.
        path = write_pack(
            tmp_path,
            "version: 1\nrules:\n"
            "  - id: on-multipass\n    signature_id: %(a)s\n"
            "    class: infra.provisioning\n    subclass: multipass_setup\n"
            "    job:\n      substrate: '^multipass$'\n"
            "  - id: elsewhere\n    signature_id: %(a)s\n    class: product.bug\n",
        )
        with caplog.at_level("WARNING"):
            load_rules(path)
        assert "unreachable" not in caplog.text


class TestMatching:
    def test_signature_pin_matches_only_its_signature(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n  - id: r\n    signature_id: %(a)s\n"
                "    class: product.bug\n    subclass: service_not_active\n",
            )
        )
        assert pack.apply(make_job(signature_id=SIG_A)).id == "r"
        assert pack.apply(make_job(signature_id=SIG_B)) is None

    def test_classification_fields_are_written_through(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 7\nrules:\n  - id: r\n    signature_id: %(a)s\n"
                "    class: product.bug\n    subclass: service_not_active\n"
                "    confidence: 0.8\n",
            )
        )
        job = make_job(signature_id=SIG_A)
        pack.apply(job)
        assert job.failure_class == "product.bug"
        assert job.subclass == "service_not_active"
        assert job.confidence == 0.8
        assert job.classified_by == ClassifiedBy.RULES.value
        assert job.ruleset_version == 7
        assert job.rule_id == "r"

    def test_first_matching_rule_wins(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n"
                "  - id: first\n    signature_id: %(a)s\n    class: infra.runner\n"
                "  - id: second\n    signature_id: %(a)s\n    class: product.bug\n",
            )
        )
        job = make_job(signature_id=SIG_A)
        assert pack.apply(job).id == "first"
        assert job.failure_class == "infra.runner"

    def test_criteria_are_anded(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n  - id: r\n    class: infra.provisioning\n"
                "    subclass: lxd_setup\n"
                "    excerpt_pattern: 'no space left'\n"
                "    job:\n      arch: arm64\n",
            )
        )
        assert pack.apply(make_job(evidence_excerpt="no space left", arch="arm64"))
        # Right text, wrong arch.
        assert (
            pack.apply(make_job(evidence_excerpt="no space left", arch="amd64")) is None
        )
        # Right arch, wrong text.
        assert pack.apply(make_job(evidence_excerpt="all good", arch="arm64")) is None

    def test_excerpt_pattern_does_not_match_a_job_without_an_excerpt(self, tmp_path):
        # F5 jobs have no log at all; they must not fall into a text rule.
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n  - id: r\n    class: product.bug\n"
                "    excerpt_pattern: '.*'\n",
            )
        )
        assert pack.apply(make_job(evidence_excerpt=None)) is None

    def test_router_verdicts_are_never_overwritten(self, tmp_path):
        # Stage A decided from job metadata, which outranks log text: a job
        # whose runner vanished will print product-looking noise on the way out.
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n  - id: r\n    signature_id: %(a)s\n"
                "    class: product.bug\n",
            )
        )
        job = make_job(
            signature_id=SIG_A,
            failure_class="infra.runner",
            subclass="runner_lost",
            classified_by=ClassifiedBy.ROUTER.value,
        )
        assert pack.apply(job) is None
        assert job.failure_class == "infra.runner"

    def test_patterns_are_case_insensitive(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 1\nrules:\n  - id: r\n    class: external.dependency\n"
                "    subclass: snap_store\n"
                "    excerpt_pattern: 'cannot install'\n",
            )
        )
        assert pack.apply(make_job(evidence_excerpt="Cannot Install snap"))


class TestSummary:
    def test_summary_counts_each_outcome(self, tmp_path):
        pack = load_rules(
            write_pack(
                tmp_path,
                "version: 2\nrules:\n  - id: r\n    signature_id: %(a)s\n"
                "    class: product.bug\n",
            )
        )
        jobs = [
            make_job(signature_id=SIG_A),
            make_job(signature_id=SIG_A),
            make_job(signature_id=SIG_B),
            make_job(classified_by=ClassifiedBy.ROUTER.value),
        ]
        summary = classify_with_rules(jobs, pack)
        assert summary == {
            "ruleset_version": 2,
            "matched": 2,
            "unmatched": 1,
            "skipped_router": 1,
            "by_rule": {"r": 2},
        }


class TestShippedRulePack:
    def test_the_repository_rule_pack_loads_and_validates(self):
        # Guards against a bad rule reaching main: every seeded signature id,
        # class, subclass and regex is checked here.
        pack = load_rules()
        assert pack.version >= 1
        assert pack.rules, "the shipped rule pack should not be empty"
        for rule in pack.rules:
            assert rule.notes, "rule {rule.id} should explain itself"
