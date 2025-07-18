#!/usr/bin/env python3
"""
ELF Patcher Script using LIEF

This tool inspects and modifies ELF binaries' RPATH and dynamic interpreter (PT_INTERP) fields.
It serves a similar purpose to the NixOS `patchelf` utility, but avoids several known bugs in
the upstream `patchelf` tool, particularly issues when modifying or truncating existing RPATH
segments that can corrupt the binary or silently fail.

To provide a more robust and reliable mechanism for ELF patching, this script uses the
Library to Instrument Executable Formats(LIEF - https://github.com/lief-project/LIEF) library,
which gives precise control over ELF structure and allows safe rewriting of RPATH and interpreter fields.

Usage:
    ./patchelf.py <elf-binary> [--set-rpath <path>] [--set-interpreter <interp>] [--output <out>]

Examples:
    ./patchelf.py ./my_binary --set-rpath '$ORIGIN/../lib' --output ./patched_binary
    ./patchelf.py ./tool --set-interpreter /lib64/ld-linux-x86-64.so.2

"""

import lief
import argparse
import sys

def print_rpath(binary):
    rpath = binary[lief.ELF.DynamicEntry.TAG.RPATH]
    runpath = binary[lief.ELF.DynamicEntry.TAG.RUNPATH]
    if rpath:
        print(f"RPATH: {rpath.name} => {rpath.value}")
    elif runpath:
        print(f"RUNPATH: {runpath.name} => {runpath.value}")
    else:
        print("No RPATH or RUNPATH found.")

def set_rpath(binary, new_rpath):
    binary.remove(lief.ELF.DynamicEntry.TAG.RPATH)
    binary.remove(lief.ELF.DynamicEntry.TAG.RUNPATH)
    binary.add(lief.ELF.DynamicEntryRpath(new_rpath))
    print(f"Set new RPATH: {new_rpath}")

def set_interpreter(binary, new_interp):
    original = binary.interpreter
    binary.interpreter = new_interp
    print(f"Set new interpreter: {new_interp} (was: {original})")

def main():
    parser = argparse.ArgumentParser(description="LIEF-based ELF RPATH and interpreter patcher")
    parser.add_argument("elf_path", help="Path to ELF binary")
    parser.add_argument("--set-rpath", help="Set a new RPATH")
    parser.add_argument("--set-interpreter", help="Set a new ELF interpreter (PT_INTERP)")
    parser.add_argument("--output", default=None, help="Output path (defaults to in-place)")
    args = parser.parse_args()

    try:
        binary = lief.parse(args.elf_path)
        if not binary or not isinstance(binary, lief.ELF.Binary):
            print("Not a valid ELF binary.")
            sys.exit(1)
    except Exception as e:
        print(f"Failed to parse ELF: {e}")
        sys.exit(1)

    did_modify = False

    if args.set_rpath:
        set_rpath(binary, args.set_rpath)
        did_modify = True

    if args.set_interpreter:
        set_interpreter(binary, args.set_interpreter)
        did_modify = True

    if did_modify:
        output_path = args.output or args.elf_path
        binary.write(output_path)
        print(f"[✓] Patched ELF written to: {output_path}")
    else:
        print_rpath(binary)

if __name__ == "__main__":
    main()
