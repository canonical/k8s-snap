from typing import Dict
import yaml
import argparse
from pathlib import Path

def parse_yaml_content(yaml_file):
    with open(yaml_file, "r") as f:
        try:
            data: Dict = yaml.safe_load(f)
            config = data.get("config", {})
            return config.get("options")
        except yaml.YAMLError as e:
            print(f"Error parsing YAML file {yaml_file}: {e}")
            return {}

def generate_markdown(config_data, output_file):
    with open(output_file, "w") as f:
        for key, values in sorted(config_data.items()):
            f.write(f"### {key}\n")
            if "type" in values:
                f.write(f"**Type:** `{values["type"]}`\n")
            if "default" in values and values['default']:
                f.write(f"**Default Value:** `{values["default"]}`\n")
            f.write("\n")
            if "description" in values and values['description']:
                description = values["description"].strip()
                f.write(f"{description}\n")
            f.write("\n")

def parse_arguments():
    parser = argparse.ArgumentParser(
        description="Generate markdown documentation from charmcraft YAML files."
    )
    parser.add_argument(
        "input_files",
        nargs="+",
        type=str,
        help="One or more charmcraft YAML files to process"
    )
    parser.add_argument(
        "--output-dir",
        "-o",
        type=str,
        default=".",
        help="Directory where markdown files will be generated (default: current directory)"
    )
    return parser.parse_args()

def main():
    args = parse_arguments()

    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    for yaml_file in args.input_files:
        yaml_path = Path(yaml_file)
        if not yaml_path.exists():
            print(f"Error: File {yaml_file} not found")
            continue

        output_file = output_dir / f"{yaml_path.stem}.md"
        config_data = parse_yaml_content(yaml_file)
        if config_data:
            generate_markdown(config_data, output_file)
            print(f"Generated documentation for {yaml_file} charm in {output_file}")
        else:
            print(f"No config section found in {yaml_file}")

if __name__ == "__main__":
    main()
