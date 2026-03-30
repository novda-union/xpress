from __future__ import annotations

import sys

from common import git_tracked_files, load_registry, match_change_families, reflection_targets


REFLECTION_FOCUS = [
    "Makefile",
    "docker-compose.yml",
    "server/internal/handler/router.go",
    "server/internal/middleware/auth.go",
    "server/internal/service/permission_service.go",
    "web/src/App.tsx",
    "web/src/store/cart.ts",
    "web/vite.config.ts",
    "admin/layouts/default.vue",
    "admin/composables/useBranchContext.ts",
]


def main() -> int:
    registry = load_registry()
    tracked = git_tracked_files()
    focus_paths = [path for path in REFLECTION_FOCUS if path in tracked]
    families, _ = match_change_families(focus_paths, registry)
    reflection_docs = reflection_targets(registry)
    docs = [doc["path"] for doc in registry.get("documents", {}).values() if doc.get("class") == "spec"]
    if "runtime" in families:
        docs.insert(0, "README.md")

    print("Documentation reflection")
    print("Reflection scope:")
    for path in REFLECTION_FOCUS:
        state = "present" if path in tracked else "missing"
        print(f"- {path}: {state}")
    print("Tracked surface summary:")
    print(f"- tracked files scanned: {len(tracked)}")
    print(f"- focus files matched: {', '.join(focus_paths) if focus_paths else 'none'}")
    print(f"- change families matched: {', '.join(families) if families else 'none'}")
    print("Recommended docs:")
    for path in dict.fromkeys([*reflection_docs, *docs]):
        print(f"- {path}")
    print("Why:")
    if families:
        for family in families:
            print(f"- repo surface matches change family: {family}")
    else:
        print("- no specific change family matched the current tracked surface")
    print("Suggested next step:")
    print("- review AGENTS.md and any listed docs for stale reflection or workflow guidance")
    return 0


if __name__ == "__main__":
    sys.exit(main())
