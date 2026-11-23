#
# Copyright 2026 Canonical, Ltd.
#
from pathlib import Path
from types import SimpleNamespace

import conftest
import pytest
from _pytest.outcomes import Failed
from test_util import tags

TESTS_DIR = Path(conftest.__file__).resolve().parent


def _fake_item(rel_path: str, tag: str = tags.PULL_REQUEST) -> SimpleNamespace:
    mark = SimpleNamespace(args=(tag,))
    return SimpleNamespace(
        path=TESTS_DIR / rel_path,
        nodeid=rel_path + "::test_example",
        iter_markers=lambda name: [mark] if name == "tags" else [],
    )


def test_pytest_itemcollected_rejects_tests_under_test_util():
    item = _fake_item("test_util/test_example.py")

    with pytest.raises(Failed, match="unit_tests"):
        conftest.pytest_itemcollected(item)


def test_pytest_itemcollected_allows_tagged_e2e_test():
    item = _fake_item("test_smoke.py")

    conftest.pytest_itemcollected(item)
