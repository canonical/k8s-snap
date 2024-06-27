# cis-yaml-to-md

## Description

This script parses YAML files from an input directory and generates corresponding Markdown files in an output directory using a Jinja2 template.

This allows us to define a set of input files that contain CIS benchmarks and generate a complete report in Markdown format.

## Usage

1. Clone our fork of the kube-bench repository

```
git clone git@github.com:canonical/kube-bench.git
cd kube-bench && git checkout ck8s # TODO remove when merged
cd ..
```

2. Create a virtual environment

```
python3 -m venv venv
source /venv/bin/activate
```

3. Install requirements

```
pip install -r requirements.txt
```

4. Run the script

```
python3 cis-yaml-to-md.py --input-dir=./kube-bench/cfg/cis-1.24-ck8s --output-dir=../../../docs/src/_parts/cis/
```

You should see the following output:

```
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/policies.yaml to ../../../docs/src/_parts/cis/policies.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/node.yaml to ../../../docs/src/_parts/cis/node.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/controlplane.yaml to ../../../docs/src/_parts/cis/controlplane.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/master.yaml to ../../../docs/src/_parts/cis/master.md.
INFO:__main__:Rendered kube-bench/cfg/cis-1.24-ck8s/etcd.yaml to ../../../docs/src/_parts/cis/etcd.md
```
