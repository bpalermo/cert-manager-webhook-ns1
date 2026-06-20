"""Linter aspect declarations for aspect_rules_lint.

These are surfaced as `*_test` rules (via `lint_test`) so that lint violations
fail under `bazel test //...`.
"""

load("@aspect_rules_lint//lint:lint_test.bzl", "lint_test")
load("@aspect_rules_lint//lint:shellcheck.bzl", "lint_shellcheck_aspect")

shellcheck = lint_shellcheck_aspect(
    binary = "@aspect_rules_lint//lint:shellcheck_bin",
    config = "@@//:.shellcheckrc",
)

shellcheck_test = lint_test(aspect = shellcheck)
