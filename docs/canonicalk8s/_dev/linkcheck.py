import os
import sys
import subprocess
import argparse


def run_command(command, cwd):
    subprocess.run(command, check=True, shell=False, cwd=cwd)


def determine_makefile(makefile, working_dir, install_target, linkcheck_target):
    # If the Makefile has not been specified, use the starter pack Makefile (and the corresponding
    # targets) if available. Otherwise, use "Makefile".
    if makefile == "use-default":
        if os.path.exists(os.path.join(working_dir, "Makefile.sp")):
            makefile = "Makefile.sp"
            install_target = "sp-" + install_target
            linkcheck_target = "sp-" + linkcheck_target
        else:
            makefile = "Makefile"
    return makefile, install_target, linkcheck_target


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("working_dir")
    parser.add_argument("--install_target")
    parser.add_argument("--linkcheck_target")
    parser.add_argument("--makefile")
    parser.add_argument("changed_files")
    return parser.parse_args()


def main():

    args = parse_args()

    makefile, install_target, linkcheck_target = determine_makefile(
        args.makefile, args.working_dir, args.install_target, args.linkcheck_target
    )
    try:

        # Install the doc framework and run link checker
        install_cmd = ["make", "-f", makefile, install_target]
        run_command(install_cmd, args.working_dir)
        linkcheck_cmd = [
            "make",
            "-f",
            makefile,
            linkcheck_target,
        ]

        # Only add the FILES variable if changed_files is not empty
        if args.changed_files and args.changed_files.strip():
            linkcheck_cmd.append(f"FILES={args.changed_files}")

        print(f"Executing: {' '.join(linkcheck_cmd)} in {args.working_dir}")
        run_command(linkcheck_cmd, args.working_dir)

    except subprocess.CalledProcessError as e:
        cmd_str = " ".join(e.cmd) if isinstance(e.cmd, list) else e.cmd
        print(f"Command '{cmd_str}' returned non-zero exit status {e.returncode}.")
        exit(1)


if __name__ == "__main__":
    main()

