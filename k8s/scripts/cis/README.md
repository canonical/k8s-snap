# cis-yaml-to-md

## Description

This script parses YAML files from an input directory and generates corresponding Markdown files in an output directory using a Jinja2 template.

This allows us to define a set of input files that contain CIS benchmarks and generate a complete report in Markdown format.

The input directory is expected to contain a config.yaml and kube-bench files.
For example, the current set of markdown files in docs/src/_parts/cis were generated using the following folder as the input directory: [cis-1.24-ck8s](https://github.com/canonical/kube-bench/tree/ck8s/cfg/cis-1.24-ck8s)

This script processes each input YAML file one at a time, and produces a corresponding Markdown file. Therefore, if there are 5 input files that are not named *config.yaml*, there will be 5 output files. The script uses the config.yaml file but does not render it to Markdown.

## Usage

Clone our fork of the kube-bench repository

```sh
git clone git@github.com:canonical/kube-bench.git
cd kube-bench && git checkout ck8s # TODO remove when merged
cd ..
```

Create a virtual environment

```sh
python3 -m venv venv
source /venv/bin/activate
```

Install dependencies

```sh
pip install -r requirements.txt
```

Run the script

```sh
python3 cis-yaml-to-md.py --input-dir=./kube-bench/cfg/cis-1.24-ck8s --output-dir=../../../docs/src/_parts/cis/
```

You should see the following output:

```sh
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/policies.yaml to ../../../docs/src/_parts/cis/policies.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/node.yaml to ../../../docs/src/_parts/cis/node.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/controlplane.yaml to ../../../docs/src/_parts/cis/controlplane.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/master.yaml to ../../../docs/src/_parts/cis/master.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/etcd.yaml to ../../../docs/src/_parts/cis/etcd.md
```
