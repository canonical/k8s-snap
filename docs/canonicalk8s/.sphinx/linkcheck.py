import os
import sys
import subprocess
import argparse

def run_command(command, cwd):
    subprocess.run(command, check=True, shell=False, cwd=cwd)

parser = argparse.ArgumentParser()
parser.add_argument("working_dir")
parser.add_argument("--install_target")
parser.add_argument("--linkcheck_target")
parser.add_argument("--makefile")
parser.add_argument("changed_files")
args = parser.parse_args()

install_target = args.install_target
linkcheck_target = args.linkcheck_target
changed_files = args.changed_files
makefile = args.makefile

try:
    # If the Makefile has not been specified, use the starter pack Makefile (and the corresponding
    # targets) if available. Otherwise, use "Makefile".
    if makefile == "use-default":
        if os.path.exists(os.path.join(args.working_dir, "Makefile.sp")):
            makefile = "Makefile.sp"
            install_target = "sp-" + install_target
            linkcheck_target = "sp-" + linkcheck_target
        else:
            makefile = "Makefile"

    # Install the doc framework and run link checker
    install_cmd = ["make", "-f", makefile, install_target]
    run_command(install_cmd, args.working_dir)
    linkcheck_cmd = ["make","-f", makefile, linkcheck_target, f"FILES={changed_files}"]
    run_command(linkcheck_cmd, args.working_dir)

except subprocess.CalledProcessError as e:
    cmd_str = ' '.join(e.cmd) if isinstance(e.cmd, list) else e.cmd
    print(f"Command '{cmd_str}' returned non-zero exit status {e.returncode}.")
    exit(1)
