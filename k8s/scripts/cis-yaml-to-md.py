import argparse
import logging
from pathlib import Path

import yaml
from jinja2 import Environment, FileSystemLoader

USAGE = """
python3 cis-yaml-to-md.py --input-directory=INPUT_DIR --output-directory=OUTPUT_DIR
"""

DESCRIPTION = """
Parse the YAML files in the input directory and generate
Markdown files in the output directory. The Markdown files are generated using the
Jinja2 template file.

It is expected that the input directory contains a config.yaml file. This file
will not be rendered to Markdown, but it will be used to extract variables that
will be replaced in the Jinja2 template.
"""

logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

JINJA_TEMPLATE = "cis-template.jinja2"
CONFIG_FILE = "config.yaml"


def get_variable_substitutions(data: str):
    return {
        "$DATA_DIR": data["master"]["etcd"]["confs"][0],
        "$apiserverbin": data["master"]["apiserver"]["bins"][0],
        "$apiserverconf": data["master"]["apiserver"]["confs"][0],
        "$controllermanagerbin": data["master"]["controllermanager"]["bins"][0],
        "$controllermanagerconf": data["master"]["controllermanager"]["confs"][0],
        "$controllermanagerkubeconfig": data["master"]["controllermanager"][
            "kubeconfig"
        ][0],
        "$etcdbin": data["master"]["etcd"]["bins"][0],
        "$etcdconf": data["master"]["etcd"]["confs"][0],
        "$kubeletbin": data["node"]["kubelet"]["bins"][0],
        "$kubeletbin": data["node"]["kubelet"]["bins"][0],
        "$kubeletcafile": data["node"]["kubelet"]["cafile"][0],
        "$kubeletconf": data["node"]["kubelet"]["confs"][0],
        "$kubeletkubeconfig": data["node"]["kubelet"]["kubeconfig"][0],
        "$kubeletsvc": data["node"]["kubelet"]["svc"][0],
        "$proxykubeconfig": data["node"]["proxy"]["kubeconfig"][0],
        "$schedulerbin": data["master"]["scheduler"]["bins"][0],
        "$schedulerconf": data["master"]["scheduler"]["confs"][0],
        "$schedulerkubeconfig": data["master"]["scheduler"]["kubeconfig"][0],
    }


def make_template(data: str) -> str:
    def to_yaml_filter(value):
        return yaml.dump(value, default_flow_style=False).strip()

    env = Environment(
        loader=FileSystemLoader("."),
        trim_blocks=True,
        lstrip_blocks=True,
    )
    env.filters["to_yaml"] = to_yaml_filter

    return env.get_template(JINJA_TEMPLATE).render(
        title=data["text"],
        groups=data["groups"],
    )


def generate_markdown(input_dir: Path, output_dir: Path):
    # All files in the input directory, but not the config.yaml file.
    input_files = [
        file
        for file in input_dir.iterdir()
        if file.is_file() and file.suffix == ".yaml" and file.name != "config.yaml"
    ]

    substs = get_variable_substitutions(
        yaml.safe_load((input_dir / "config.yaml").read_text())
    )

    for file in input_files:
        control_data = yaml.safe_load(file.read_text())

        markdown_content = make_template(control_data)

        for old, new in substs.items():
            markdown_content = markdown_content.replace(old, new)

        output_dir.mkdir(exist_ok=True)

        output_file = output_dir / f"{file.stem}.md"
        output_file.touch()
        output_file.write_text(markdown_content)

        LOG.info(f"Rendered {file} to {output_file}.")


def main():
    parser = argparse.ArgumentParser(usage=USAGE, description=DESCRIPTION)
    parser.add_argument(
        "--input-dir", type=str, required=True, help="Input directory path"
    )
    parser.add_argument(
        "--output-dir", type=str, required=True, help="Output directory path"
    )
    args = parser.parse_args()

    generate_markdown(Path(args.input_dir), Path(args.output_dir))


if __name__ == "__main__":
    main()
