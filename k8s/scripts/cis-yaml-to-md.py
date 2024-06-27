import argparse
import logging
from pathlib import Path

import yaml
from jinja2 import Environment, FileSystemLoader

from typing import Any

USAGE = """
python3 cis-yaml-to-md.py --input-directory=INPUT_DIR --output-directory=OUTPUT_DIR
"""

DESCRIPTION = """
Parse the YAML files in the input directory and generate
Markdown files in the output directory. The Markdown files are generated using the
Jinja2 template file.

It is expected that the input directory contains a config.yaml file.
(See https://raw.githubusercontent.com/canonical/kube-bench/ck8s/cfg/cis-1.24-ck8s/config.yaml)

It is also expected that the input directory contains the control files.
(See controlplane.md, etcd.yaml, master.yaml, node.yaml and policies.yaml from the same repo.)

The config.yaml file will not be rendered to Markdown, but it will be used to extract variables
that will be replaced when generating the markdown files.
"""

logging.basicConfig(level=logging.INFO)

LOG = logging.getLogger(__name__)

DIR = Path(__file__).absolute().parent

JINJA_TEMPLATE = "cis-template.jinja2"
KUBE_BENCH_CONFIG_FILE = "config.yaml"
KUBE_BENCH_CONTROL_OUTPUTS = "expected-outputs.yaml"


def get_variable_substitutions(data: Any) -> dict[str, str]:
    return {
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
        "$kubeletcafile": data["node"]["kubelet"]["cafile"][0],
        "$kubeletconf": data["node"]["kubelet"]["confs"][0],
        "$kubeletkubeconfig": data["node"]["kubelet"]["kubeconfig"][0],
        "$kubeletsvc": data["node"]["kubelet"]["svc"][0],
        "$proxykubeconfig": data["node"]["proxy"]["kubeconfig"][0],
        "$schedulerbin": data["master"]["scheduler"]["bins"][0],
        "$schedulerconf": data["master"]["scheduler"]["confs"][0],
        "$schedulerkubeconfig": data["master"]["scheduler"]["kubeconfig"][0],
    }


def render_jinja_template(**kwargs) -> str:
    def to_yaml_filter(value):
        return yaml.dump(value, default_flow_style=False).strip()

    env = Environment(
        loader=FileSystemLoader(DIR),
        trim_blocks=True,
        lstrip_blocks=True,
    )
    env.filters["to_yaml"] = to_yaml_filter

    return env.get_template(JINJA_TEMPLATE).render(**kwargs)


def generate_markdown_from_yaml_file(
    original_yaml_file: Path,
    substitutions: dict[str, str] | None = None,
    custom_outputs_by_control_id: dict[str, str] | None = None,
):
    markdown_content = render_jinja_template(
        kube_bench_control_file=yaml.safe_load(original_yaml_file.read_text()),
        custom_outputs_by_control_id=custom_outputs_by_control_id,
    )

    if substitutions:
        for old, new in substitutions.items():
            markdown_content = markdown_content.replace(old, new)

    return markdown_content


def setup_directories(input_dir: str, output_dir: str) -> tuple[Path, Path]:
    input_path = Path(input_dir).expanduser()
    output_path = Path(output_dir).expanduser()
    output_path.mkdir(exist_ok=True)
    return input_path, output_path


def get_kube_bench_input_files(input_dir: Path) -> list[Path]:
    return [
        file
        for file in input_dir.iterdir()
        if file.is_file()
        and file.suffix == ".yaml"
        and file.name != KUBE_BENCH_CONFIG_FILE
    ]


def process_files(input_dir, output_dir, substitutions, custom_outputs_by_control_id):
    for file in get_kube_bench_input_files(input_dir):
        output_file = output_dir / f"{file.stem}.md"
        markdown_content = generate_markdown_from_yaml_file(
            file, substitutions, custom_outputs_by_control_id
        )
        output_file.write_text(markdown_content)

        LOG.info(f"Rendered {file} to {output_file}.")


def parse_arguments():
    parser = argparse.ArgumentParser(usage=USAGE, description=DESCRIPTION)
    parser.add_argument(
        "--input-dir", type=str, required=True, help="Input directory path"
    )
    parser.add_argument(
        "--output-dir", type=str, required=True, help="Output directory path"
    )
    return parser.parse_args()


def main():
    args = parse_arguments()

    input_dir, output_dir = setup_directories(args.input_dir, args.output_dir)

    substitutions = get_variable_substitutions(
        yaml.safe_load((input_dir / KUBE_BENCH_CONFIG_FILE).read_text())
    )

    custom_outputs_by_control_id: dict[str, str] = yaml.safe_load(
        (DIR / KUBE_BENCH_CONTROL_OUTPUTS).read_text()
    )

    process_files(input_dir, output_dir, substitutions, custom_outputs_by_control_id)


if __name__ == "__main__":
    main()
