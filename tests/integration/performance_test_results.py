#!/usr/bin/env python3
"""
Compare performance test results from two different runs.

This script uses pytest-benchmark to compare JSON result files
from performance tests and generates a summary of the comparison.
"""

import argparse
import glob
import json
import os
import subprocess
import sys
from pathlib import Path


def find_result_files(base_dir):
    """Find all benchmark result JSON files in the directory structure."""
    result_files = []
    for root, _, files in os.walk(base_dir):
        for file in files:
            if file.endswith('.json'):
                result_files.append(os.path.join(root, file))
    return result_files


def get_test_name_from_path(file_path):
    """Extract test name from file path."""
    return os.path.basename(file_path).split('.')[0]


def compare_benchmarks(base_dir):
    """Compare benchmark results and generate a summary."""
    result_files = find_result_files(base_dir)
    
    if len(result_files) < 2:
        print(f"Error: Found only {len(result_files)} result files, need at least 2 to compare")
        return 1
    
    # Group files by test name (to compare the same test from different runs)
    tests_by_name = {}
    for file_path in result_files:
        # Parse directory name to get run identifier
        parent_dir = os.path.basename(os.path.dirname(file_path))
        # Attempt to get metadata from the file
        with open(file_path, 'r') as f:
            try:
                data = json.load(f)
                # Extract information about which test this is
                if 'machine_info' in data and 'env' in data['machine_info']:
                    run_id = data['machine_info']['node']
                else:
                    run_id = parent_dir
            except json.JSONDecodeError:
                run_id = parent_dir
        
        # Use the base filename to identify the test
        test_name = get_test_name_from_path(file_path)
        if test_name not in tests_by_name:
            tests_by_name[test_name] = []
        
        tests_by_name[test_name].append({
            'file_path': file_path,
            'run_id': run_id
        })
    
    # Generate comparison for each test type
    summary_file = os.path.join(base_dir, "comparison_summary.md")
    with open(summary_file, 'w') as f:
        f.write("# Performance Test Comparison Summary\n\n")
        
        for test_name, test_files in tests_by_name.items():
            if len(test_files) < 2:
                f.write(f"## {test_name}\n")
                f.write(f"Not enough results to compare (found {len(test_files)} files)\n\n")
                continue
            
            # Sort by run_id to ensure consistent comparison order
            test_files.sort(key=lambda x: x['run_id'])
            
            f.write(f"## {test_name}\n")
            f.write(f"Comparing runs: {test_files[0]['run_id']} vs {test_files[1]['run_id']}\n\n")
            
            # Use pytest-benchmark command line to compare
            cmd = [
                "pytest-benchmark", 
                "compare", 
                "--csv", f"{os.path.join(base_dir, test_name)}_comparison.csv",
                "--group-by", "name",
                "--columns", "min,max,mean,median,iqr,stddev,ops",
                "--sort", "name",
                test_files[0]['file_path'], 
                test_files[1]['file_path']
            ]
            
            try:
                # Run the comparison command and capture output
                result = subprocess.run(
                    cmd, 
                    stdout=subprocess.PIPE, 
                    stderr=subprocess.PIPE,
                    text=True,
                    check=True
                )
                
                # Write the command output to the summary
                f.write("```\n")
                f.write(result.stdout)
                f.write("```\n\n")
                
                # Check if there's a warning or error in stderr
                if result.stderr:
                    f.write("Warnings/Errors:\n")
                    f.write("```\n")
                    f.write(result.stderr)
                    f.write("```\n\n")
                    
            except subprocess.CalledProcessError as e:
                f.write(f"Error comparing benchmarks: {e}\n")
                f.write("```\n")
                if e.stdout:
                    f.write(e.stdout)
                if e.stderr:
                    f.write(e.stderr)
                f.write("```\n\n")
    
    print(f"Comparison summary written to {summary_file}")
    return 0


def main():
    parser = argparse.ArgumentParser(description='Compare performance test results.')
    parser.add_argument('--dir', default='./performance-test-results',
                        help='Directory containing performance test results')
    args = parser.parse_args()
    
    return compare_benchmarks(args.dir)


if __name__ == "__main__":
    sys.exit(main())
