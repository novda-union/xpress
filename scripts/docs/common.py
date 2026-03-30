from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import json
import subprocess
from typing import Iterable


ROOT = Path(__file__).resolve().parents[2]
REGISTRY_PATH = ROOT / "docs" / "registry.yml"


@dataclass(frozen=True)
class AdvisoryResult:
    classifications: list[str]
    severity: str
    documents: list[str]
    reasons: list[str]


def load_registry() -> dict:
    with REGISTRY_PATH.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def git_diff_names(staged: bool = False) -> list[str]:
    args = ["git", "diff", "--name-only"]
    if staged:
        args.append("--cached")
    proc = subprocess.run(args, cwd=ROOT, capture_output=True, text=True, check=True)
    return [line.strip() for line in proc.stdout.splitlines() if line.strip()]


def git_tracked_files() -> list[str]:
    proc = subprocess.run(["git", "ls-files"], cwd=ROOT, capture_output=True, text=True, check=True)
    return [line.strip() for line in proc.stdout.splitlines() if line.strip()]


def _matches_prefix(path: str, prefix: str) -> bool:
    return path == prefix or path.startswith(prefix)


def match_change_families(paths: Iterable[str], registry: dict) -> tuple[list[str], list[str]]:
    families: list[str] = []
    tags: list[str] = []
    for family_name, config in registry.get("change_families", {}).items():
        family_paths = config.get("paths", [])
        if any(any(_matches_prefix(path, family_path) for family_path in family_paths) for path in paths):
            families.append(family_name)
            tags.extend(config.get("tags", []))
    return families, sorted(set(tags))


def select_documents(tags: Iterable[str], registry: dict) -> list[str]:
    tag_set = set(tags)
    results: list[str] = []
    for doc in registry.get("documents", {}).values():
        owns = set(doc.get("owns", []))
        if tag_set & owns:
            results.append(doc["path"])
    return sorted(dict.fromkeys(results))


def classify_severity(paths: list[str], families: list[str]) -> str:
    if not paths or not families:
        return "no-doc-action"
    if any(family in {"runtime", "backend_architecture"} for family in families):
        return "strongly-recommend-update"
    return "suggest-review"


def doc_label(path: str) -> str:
    return path


def reflection_targets(registry: dict) -> list[str]:
    targets: list[str] = []
    for doc in registry.get("documents", {}).values():
        if doc.get("class") == "reflection":
            targets.append(doc["path"])
    return targets


def report_lines(title: str, lines: Iterable[str]) -> list[str]:
    return [title, *[f"- {line}" for line in lines]]
