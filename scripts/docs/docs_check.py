from __future__ import annotations

import argparse
import sys

from common import classify_severity, git_diff_names, load_registry, match_change_families, select_documents


def main() -> int:
    parser = argparse.ArgumentParser(description="Advisory documentation relevance check")
    parser.add_argument("--staged", action="store_true", help="Inspect staged diff instead of working tree diff")
    args = parser.parse_args()

    registry = load_registry()
    paths = git_diff_names(staged=args.staged)

    print("Documentation advisory")
    if not paths:
        print("Change classification: none")
        print("Severity: no-doc-action")
        print("Recommended docs: none")
        print("Why:")
        print("- no changed files detected")
        print("Suggested next step:")
        print("- no documentation action is likely required")
        return 0

    families, tags = match_change_families(paths, registry)
    docs = select_documents(tags, registry)
    severity = classify_severity(paths, families)

    print(f"Change classification: {', '.join(families) if families else 'local-only'}")
    print(f"Severity: {severity}")
    print("Recommended docs:")
    if docs:
        for doc in docs:
            print(f"- {doc}")
    else:
        print("- none")
    print("Why:")
    if families:
        for family in families:
            print(f"- matched change family: {family}")
    else:
        print("- no configured architecture or workflow family matched")
    print("Suggested next step:")
    if docs:
        print("- review the listed docs and decide whether they need updates")
        print("- run make docs-refresh for a broader reflection pass")
    else:
        print("- no documentation action is likely required")
    return 0


if __name__ == "__main__":
    sys.exit(main())
