#!/usr/bin/env python3
"""
Extract every ```terraform block from the documentation prose and classify it.

The examples under docs-examples/ are whole files and are already type-checked by
validate-examples.sh. The blocks *inside* prose are not, and they are not uniform,
so each kind has to be checked the way it can actually be checked:

  standalone  declares at least one top-level block (resource, data, locals, ...).
              Type-checked by `terraform validate` once a provider block is
              injected, exactly like a docs-examples fragment.

  mixed       a bare attribute assignment *and* a block in the same fence. The
              attribute is looked up in the schema; the block is validated.

  import      contains only `import {}` blocks. Terraform rejects these unless the
              resource named in `to` is also in the configuration, so validating one
              alone reports a failure that says nothing about the example. The `to`
              address is checked against the provider schema instead, which is the
              part that actually goes stale.

  fragment    a bare attribute assignment shown out of context, e.g.
              `iac_template_id = "my-template:1"`. Not valid at the top level of a
              .tf file, so its attribute names are looked up in the schema.

Blocks in a guide are cumulative: a later fence refers to a resource an earlier one
declared. Validating each fence alone would report "Reference to undeclared
resource" for examples that are correct as written, so every block that can be
validated is given the earlier blocks of the same page as context.tf. Where a page
redeclares an address, the latest definition wins, which is what a reader following
along would end up with.

Blocks are written to --out as per-block directories and described in a
tab-separated manifest, so the shell harness can drive `terraform validate` without
parsing JSON. Blocks that cannot be validated as configuration are adjudicated here
against the schema in --schema and carry their verdict in the manifest.

Manifest columns:
    index  kind  source  line  dir  verdict  detail
"""

from __future__ import annotations

import argparse
import glob
import json
import os
import re
import sys

# Top-level block types that make a snippet self-contained enough to validate.
CONFIG_BLOCKS = {
    "resource", "data", "module", "variable", "output",
    "locals", "provider", "terraform",
}

# Blocks addressed by type+name; a page may redeclare one, and the last wins.
ADDRESSED = {"resource", "data", "module", "variable", "output"}

FENCE_OPEN = "```terraform"
FENCE_CLOSE = "```"
IDENT = r'([A-Za-z_][A-Za-z0-9_-]*)'


def brace_delta(line: str) -> int:
    """Net brace depth change for a line, ignoring braces in strings and comments.

    Terraform examples routinely carry jsonencode({...}) and "${...}", so a naive
    count of { and } misreads nesting and every classification downstream of it.
    """
    kept: list[str] = []
    i, in_string = 0, False
    while i < len(line):
        c = line[i]
        if in_string:
            if c == "\\":
                i += 2
                continue
            if c == '"':
                in_string = False
            i += 1
            continue
        if c == '"':
            in_string = True
            i += 1
            continue
        if c == "#":
            break
        if c == "/" and i + 1 < len(line) and line[i + 1] == "/":
            break
        kept.append(c)
        i += 1
    s = "".join(kept)
    return s.count("{") - s.count("}")


def split_chunks(body: str) -> list[dict]:
    """Split a block into its top-level items, keeping each item's own comments."""
    lines = body.split("\n")
    chunks: list[dict] = []
    pending: list[str] = []
    i = 0
    while i < len(lines):
        stripped = lines[i].strip()
        if not stripped or stripped.startswith(("#", "//")):
            pending.append(lines[i])
            i += 1
            continue

        m_block = re.match(IDENT + r'\s+"', stripped) or re.match(IDENT + r'\s*\{', stripped)
        m_attr = re.match(IDENT + r'\s*=', stripped)

        start, depth = i, 0
        while i < len(lines):
            depth += brace_delta(lines[i])
            i += 1
            if depth <= 0:
                break
        text = "\n".join(pending + lines[start:i])
        pending = []

        if m_block:
            kind, name = "block", m_block.group(1)
        elif m_attr:
            kind, name = "attr", m_attr.group(1)
        else:
            kind, name = "other", stripped[:24]
        chunks.append({"kind": kind, "name": name, "text": text})
    return chunks


def address_of(chunk: dict) -> tuple | None:
    """Terraform address for a block chunk, used to dedupe cumulative context."""
    if chunk["kind"] != "block":
        return None
    btype = chunk["name"]
    if btype not in ADDRESSED:
        return (btype,)  # locals/terraform/provider: one per page is enough
    header = next(
        (l for l in chunk["text"].split("\n")
         if l.strip() and not l.strip().startswith(("#", "//"))),
        "",
    )
    return (btype, *re.findall(r'"([^"]+)"', header)[:2])


def is_elided(body: str) -> bool:
    """True if the block says, in the usual way, that fields are left out.

    A `# ...` line means the example deliberately shows one part of a resource, so
    the required arguments it omits are not defects and `terraform validate` cannot
    judge it. Its attribute names are still worth checking.
    """
    return any(
        re.fullmatch(r'(#|//)\s*\.{3}\.*', line.strip())
        for line in body.split("\n")
    )


def classify(chunks: list[dict], body: str) -> str:
    block_types = {c["name"] for c in chunks if c["kind"] == "block"}
    has_attr = any(c["kind"] == "attr" for c in chunks)
    if block_types & CONFIG_BLOCKS:
        if is_elided(body):
            return "partial"
        return "mixed" if has_attr else "standalone"
    if "import" in block_types:
        return "import"
    if has_attr:
        return "fragment"
    return "unknown"


def block_surface(text: str) -> tuple[str | None, list[str]]:
    """A block's type and its own immediate attribute names.

    Only the outermost level is returned. Going deeper would descend into
    jsonencode({...}) payloads, whose keys are JSON fields rather than schema
    attributes, and report every one of them as unknown.
    """
    lines = text.split("\n")
    header_idx = next(
        (i for i, l in enumerate(lines)
         if l.strip() and not l.strip().startswith(("#", "//"))),
        None,
    )
    if header_idx is None:
        return None, []
    header = lines[header_idx]
    labels = re.findall(r'"([^"]+)"', header)
    btype = labels[0] if labels else None

    names: list[str] = []
    depth = brace_delta(header)
    for line in lines[header_idx + 1:]:
        stripped = line.strip()
        if depth == 1 and stripped and not stripped.startswith(("#", "//")):
            if m := re.match(IDENT + r'\s*(=|\{)', stripped):
                names.append(m.group(1))
        depth += brace_delta(line)
    return btype, names


def extract(docs_dirs: list[str]) -> list[dict]:
    """Pull every terraform-fenced block out of the markdown sources."""
    blocks: list[dict] = []
    files: list[str] = []
    for d in docs_dirs:
        for pat in ("**/*.md", "**/*.md.tmpl"):
            files.extend(glob.glob(os.path.join(d, pat), recursive=True))
    for path in sorted(set(files)):
        with open(path, encoding="utf-8") as fh:
            lines = fh.read().split("\n")
        in_block, buf, start = False, [], 0
        for lineno, line in enumerate(lines, 1):
            if not in_block and line.strip() == FENCE_OPEN:
                in_block, buf, start = True, [], lineno
            elif in_block and line.strip() == FENCE_CLOSE:
                in_block = False
                blocks.append({"source": path, "line": start, "body": "\n".join(buf)})
            elif in_block:
                buf.append(line)
        if in_block:
            print(f"WARN unclosed ```terraform fence in {path} at line {start}",
                  file=sys.stderr)
    return blocks


def load_schema(path: str) -> tuple[set[str], set[str], set[str]]:
    """Return (resource types, data source types, every attribute name)."""
    with open(path, encoding="utf-8") as fh:
        schema = json.load(fh)

    resources: set[str] = set()
    datasources: set[str] = set()
    attributes: set[str] = set()

    def walk(block: dict) -> None:
        for name, attr in block.get("attributes", {}).items():
            attributes.add(name)
            # SingleNested/ListNested attributes hide their children under
            # nested_type rather than block_types; missing this reports real
            # attributes such as iac_template_id as unknown.
            if nested := attr.get("nested_type"):
                walk(nested)
        for name, nested in block.get("block_types", {}).items():
            attributes.add(name)
            walk(nested.get("block", {}))

    for provider in schema.get("provider_schemas", {}).values():
        for name, body in provider.get("resource_schemas", {}).items():
            resources.add(name)
            walk(body.get("block", {}))
        for name, body in provider.get("data_source_schemas", {}).items():
            datasources.add(name)
            walk(body.get("block", {}))
    return resources, datasources, attributes


def check_import(body: str, resources: set[str], datasources: set[str]):
    """An import block's `to` must name a resource type the provider registers."""
    targets = re.findall(r'^\s*to\s*=\s*([A-Za-z_][A-Za-z0-9_]*)\.', body, re.M)
    if not targets:
        return "FAIL", "import block has no `to` address"
    bad = [t for t in targets if t not in resources]
    if bad:
        hint = " (that is a data source, not a resource)" \
            if any(t in datasources for t in bad) else ""
        return "FAIL", f"unknown resource type: {', '.join(sorted(set(bad)))}{hint}"
    return "PASS", ",".join(sorted(set(targets)))


def check_attrs(names: list[str], attributes: set[str]):
    """Every attribute shown out of context should still exist in the schema."""
    if not names:
        return "FAIL", "no attributes found"
    bad = [n for n in names if n not in attributes]
    if bad:
        return "FAIL", f"unknown attribute: {', '.join(sorted(set(bad)))}"
    return "PASS", ",".join(sorted(set(names)))


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--docs-dir", action="append", required=True,
                    help="directory of markdown sources (repeatable)")
    ap.add_argument("--out", required=True, help="directory to write block dirs into")
    ap.add_argument("--schema", required=True, help="terraform providers schema -json")
    ap.add_argument("--manifest", required=True, help="TSV manifest to write")
    args = ap.parse_args()

    resources, datasources, attributes = load_schema(args.schema)
    if not resources:
        print("ERROR: provider schema contained no resource types", file=sys.stderr)
        return 2

    os.makedirs(args.out, exist_ok=True)
    blocks = extract(args.docs_dir)

    # Earlier blocks of the same page, keyed by address so a redeclaration wins.
    context: dict[str, dict[tuple, str]] = {}
    rows = []

    for idx, block in enumerate(blocks, 1):
        chunks = split_chunks(block["body"])
        kind = classify(chunks, block["body"])
        source = block["source"]
        seen = context.setdefault(source, {})

        block_chunks = [c for c in chunks if c["kind"] == "block"]
        attr_names = [c["name"] for c in chunks if c["kind"] == "attr"]

        if kind == "partial":
            # Deliberately incomplete, so it is neither validated nor added to the
            # page context, where its missing arguments would fail later blocks.
            problems, shown = [], []
            for c in block_chunks:
                btype, names = block_surface(c["text"])
                if btype and btype not in resources and btype not in datasources:
                    problems.append(f"unknown resource type: {btype}")
                bad = [n for n in names if n not in attributes]
                if bad:
                    problems.append(f"unknown attribute: {', '.join(sorted(set(bad)))}")
                shown.extend(names)
            verdict, detail = ("FAIL", "; ".join(problems)) if problems \
                else ("PASS", ",".join(sorted(set(shown))) or "elided")

        if kind in ("standalone", "mixed"):
            verdict, detail = "DEFER", "terraform validate"
            if kind == "mixed":
                # The bare attribute cannot go in a .tf file; check it separately
                # and validate only the block half.
                verdict_attr, detail_attr = check_attrs(attr_names, attributes)
                if verdict_attr == "FAIL":
                    verdict, detail = "FAIL", detail_attr

            own = {a for c in block_chunks if (a := address_of(c))}
            bdir = os.path.join(args.out, f"{idx:03d}")
            os.makedirs(bdir, exist_ok=True)
            with open(os.path.join(bdir, "main.tf"), "w", encoding="utf-8") as fh:
                fh.write("\n\n".join(c["text"] for c in block_chunks) + "\n")

            # Context excludes anything this block declares itself, so a page that
            # shows the same resource twice does not become a duplicate.
            prior = [text for addr, text in seen.items() if addr not in own]
            ctx_path = os.path.join(bdir, "context.tf")
            if prior:
                with open(ctx_path, "w", encoding="utf-8") as fh:
                    fh.write("\n\n".join(prior) + "\n")
            elif os.path.exists(ctx_path):
                os.remove(ctx_path)

            for c in block_chunks:
                if addr := address_of(c):
                    seen[addr] = c["text"]

            rows.append((str(idx), kind, source, str(block["line"]),
                         bdir, verdict, detail))
            continue

        if kind == "import":
            verdict, detail = check_import(block["body"], resources, datasources)
        elif kind == "fragment":
            verdict, detail = check_attrs(attr_names, attributes)
        elif kind == "unknown":
            verdict, detail = "FAIL", "no recognisable terraform content"
        # a partial block was already adjudicated above

        bdir = os.path.join(args.out, f"{idx:03d}")
        os.makedirs(bdir, exist_ok=True)
        with open(os.path.join(bdir, "main.tf"), "w", encoding="utf-8") as fh:
            fh.write(block["body"] + "\n")
        rows.append((str(idx), kind, source, str(block["line"]),
                     bdir, verdict, detail))

    with open(args.manifest, "w", encoding="utf-8") as fh:
        for row in rows:
            fh.write("\t".join(row) + "\n")

    counts: dict[str, int] = {}
    for _, kind, *_ in rows:
        counts[kind] = counts.get(kind, 0) + 1
    summary = ", ".join(f"{k} {v}" for k, v in sorted(counts.items()))
    print(f"==> extracted {len(rows)} terraform blocks ({summary})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
