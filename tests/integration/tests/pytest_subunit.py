#
# Copyright 2026 Canonical, Ltd.
#

# Based on:
# * https://pypi.org/project/pytest-subunit/
# * https://github.com/jelmer/pytest-subunit

import datetime
import io
import re

from _pytest._io import TerminalWriter
from _pytest.terminal import TerminalReporter
from subunit import StreamResultToBytes


# We're extending TerminalReporter to also write a subunit stream to a file.
# This approach is much easier than having a standalone plugin for subunit
# reports. By doing so, we avoid maintaining a lot of pytest reporter boilerplate.
class SubunitTerminalReporter(TerminalReporter):
    ansi_escape_re = re.compile(
        r"(?:\x1B[@-Z\\-_]|[\x80-\x9A\x9C-\x9F]|(?:\x1B\[|\x9B)[0-?]*[ -/]*[@-~])"
    )

    def __init__(self, reporter, subunit_path):
        super().__init__(reporter.config)
        self.writer = self._tw
        self.skipped = []
        self.failed = []

        self.subunit_file = open(subunit_path, "ab")

    def _status(self, report, status):
        test_id = report.nodeid

        summary = io.StringIO()
        writer = TerminalWriter(summary)
        report.toterminal(writer)
        writer.flush()

        out_report = f"""
----------------------------  summary ({status}) -----------------------------------
{summary.getvalue()}
------------------------------- captured log ---------------------------------
{report.caplog}
------------------------------ captured stdout -------------------------------
{report.capstdout}
------------------------------ captured stderr -------------------------------
{report.capstderr}
"""  # noqa

        # Remove ANSI color codes for now.
        # TODO: consider handling color codes in subunit2html.py.
        out_report = self.ansi_escape_re.sub("", out_report)

        result = StreamResultToBytes(self.subunit_file)
        result.startTestRun()
        result.status(
            test_id=test_id,
            timestamp=datetime.datetime.fromtimestamp(
                report.start, datetime.timezone.utc
            ),
        )
        result.status(
            test_id=test_id,
            test_status=status,
            timestamp=datetime.datetime.fromtimestamp(
                report.stop, datetime.timezone.utc
            ),
            file_name="summary",
            file_bytes=out_report.encode("utf8"),
            mime_type="text/plain; charset=utf8",
        )
        result.stopTestRun()

    def pytest_runtest_logreport(self, report):
        super().pytest_runtest_logreport(report)

        test_id = report.nodeid
        if report.when in ["setup", "session"]:
            self._status(report, "exists")
            if report.outcome == "passed":
                # Avoid reporting successful initialization twice.
                # self._status(report, 'inprogress')
                pass
            elif report.outcome == "failed":
                self._status(report, "fail")
            elif report.outcome == "skipped":
                self._status(report, "skip")
        elif report.when in ["call"]:
            if hasattr(report, "wasxfail"):
                if report.skipped:
                    self._status(report, "xfail")
                elif report.failed:
                    self._status(report, "uxsuccess")
            elif report.outcome == "failed":
                self._status(report, "fail")
                self.failed.append(test_id)
            elif report.outcome == "skipped":
                self._status(report, "skip")
                self.skipped.append(test_id)
        elif report.when in ["teardown"]:
            if test_id not in self.skipped and test_id not in self.failed:
                if report.outcome == "passed":
                    self._status(report, "success")
                elif report.outcome == "failed":
                    self._status(report, "fail")
        else:
            raise Exception("Unknown pytest phase: %s" % report)
