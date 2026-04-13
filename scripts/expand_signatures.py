#!/usr/bin/env python3
"""Expand signatures seed files into a larger JSON DB.

Usage: python scripts/expand_signatures.py --target 250

This script reads JSON files from test/signatures/ (by default
`popular_signatures.json`) and writes `test/signatures/expanded_signatures.json`
containing at least --target entries by duplicating and adding suffixes when
necessary. It is safe to run multiple times.
"""
import argparse
import json
import os
import sys

DEFAULT_INPUT = 'test/signatures/popular_signatures.json'
OUTPUT = 'test/signatures/expanded_signatures.json'


def load_input(path):
    if not os.path.exists(path):
        raise SystemExit(f"input file not found: {path}")
    with open(path, 'r', encoding='utf-8') as f:
        return json.load(f)


def save_output(entries, path):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(entries, f, indent=2, ensure_ascii=False)


def expand(entries, target):
    out = list(entries)
    i = 0
    base_len = len(entries)
    if base_len == 0:
        raise SystemExit('no base signatures to expand')
    # generate duplicated variants with numeric suffixes
    while len(out) < target:
        src = entries[i % base_len]
        new = dict(src)
        # ensure name uniqueness
        suffix = f"-dup-{len(out)-base_len+1}" if len(out) >= base_len else f"-dup-{len(out)+1}"
        new_name = new.get('name', 'sig') + suffix
        new['name'] = new_name
        # optionally tweak header_contains values for variety
        if 'header_contains' in new and isinstance(new['header_contains'], dict):
            hd = {}
            for k, v in new['header_contains'].items():
                # append suffix to header value as harmless variant
                hd[k] = v
            new['header_contains'] = hd
        out.append(new)
        i += 1
    return out


if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('--input', '-i', default=DEFAULT_INPUT)
    p.add_argument('--output', '-o', default=OUTPUT)
    p.add_argument('--target', '-t', type=int, default=250)
    args = p.parse_args()

    entries = load_input(args.input)
    expanded = expand(entries, args.target)
    save_output(expanded, args.output)
    print(f"Wrote {len(expanded)} signatures to {args.output}")
