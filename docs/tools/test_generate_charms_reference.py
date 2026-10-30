#
# Copyright 2026 Canonical, Ltd.
#
import unittest

from generate_charms_reference import render_reference


class CharmGeneratorTests(unittest.TestCase):
    def setUp(self):
        self.data = {
            "name": "k8s",
            "config": {"options": {}},
            "actions": {},
        }

    def test_defaults_and_ordering(self):
        options = self.data["config"]["options"]
        for name, default in [
            ("zero", 0),
            ("false", False),
            ("empty", ""),
            ("null", None),
        ]:
            options[name] = {"type": "string", "default": default}
        options["absent"] = {"type": "string", "description": "First\n\nSecond"}
        output = render_reference(self.data)["k8s-config.md"]
        for default in ["0", "false", '""', "null"]:
            self.assertIn(f"default ``{default}``", output)
        self.assertEqual(output.count("default ``"), 4)
        self.assertIn("`string` · default ``0``", output)
        self.assertIn("First\n\nSecond", output)
        self.assertLess(output.index("### `absent`"), output.index("### `zero`"))

    def test_actions_and_nested_constraints(self):
        self.data["actions"] = {
            "run": {
                "description": "Run checks.",
                "additionalProperties": False,
                "required": ["settings"],
                "params": {
                    "optional": {"type": "boolean", "default": False},
                    "settings": {
                        "type": "object",
                        "description": "Settings to apply.",
                        "required": ["level"],
                        "properties": {
                            "level": {"type": "integer", "minimum": 0, "maximum": 10},
                        },
                    },
                },
            }
        }
        output = render_reference(self.data)["k8s-actions.md"]
        self.assertIn("(k8s-action-run)=", output)
        self.assertEqual(output.count("\nParameters:\n"), 1)
        self.assertIn("- `optional` (`boolean`, optional, default ``false``)", output)
        self.assertIn("- `settings` (`object`, required): Settings to apply.", output)
        self.assertNotIn("(k8s-action-run-settings)=", output)
        self.assertNotIn("#### ", output)
        self.assertIn("  ````yaml\n  properties:", output)
        self.assertIn("Run checks.", output)
        self.assertIn("additionalProperties: false", output)
        self.assertIn("minimum: 0", output)
        self.assertIn("maximum: 10", output)
        self.assertIn("required:\n  - level", output)


if __name__ == "__main__":
    unittest.main()
