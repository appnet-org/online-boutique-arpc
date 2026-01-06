#!/usr/bin/env python3
"""
Parse JSONL log files and generate CDF plots for message sizes across serialization libraries.
"""

import json
import os
import glob
from collections import defaultdict
import matplotlib.pyplot as plt
import matplotlib
import numpy as np
from pathlib import Path

# --- Global Style Settings ---
matplotlib.rcParams['pdf.fonttype'] = 42
matplotlib.rcParams['ps.fonttype'] = 42
matplotlib.rcParams.update({'font.size': 14})


def parse_logs(logs_dir='../logs'):
    """Parse all JSONL files in the logs directory."""
    log_files = glob.glob(os.path.join(logs_dir, '*.jsonl'))
    
    if not log_files:
        print(f"No JSONL files found in {logs_dir}")
        return []
    
    all_messages = []
    for log_file in log_files:
        print(f"Reading {log_file}...")
        with open(log_file, 'r') as f:
            for line_num, line in enumerate(f, 1):
                line = line.strip()
                if not line:
                    continue
                try:
                    message = json.loads(line)
                    all_messages.append(message)
                except json.JSONDecodeError as e:
                    print(f"Warning: Failed to parse line {line_num} in {log_file}: {e}")
                    continue
    
    return all_messages


def deduplicate_messages(messages):
    """Deduplicate messages based on payload content."""
    seen_payloads = {}
    unique_messages = []
    duplicates_count = 0
    
    for msg in messages:
        # Create JSON string representation of payload for deduplication
        payload_str = json.dumps(msg.get('payload', {}))
        
        if payload_str not in seen_payloads:
            seen_payloads[payload_str] = True
            unique_messages.append(msg)
        else:
            duplicates_count += 1
    
    return unique_messages, duplicates_count


def calculate_cdf(sizes):
    """Calculate CDF from a list of sizes."""
    if not sizes:
        return [], []
    
    sorted_sizes = sorted(sizes)
    n = len(sorted_sizes)
    cumulative_prob = [(i + 1) / n for i in range(n)]
    
    return sorted_sizes, cumulative_prob


def main():
    # Get the directory where this script is located
    script_dir = Path(__file__).parent
    logs_dir = script_dir / '../logs'
    logs_dir = logs_dir.resolve()
    
    print(f"Parsing logs from: {logs_dir}")
    
    # Parse all log files
    messages = parse_logs(str(logs_dir))
    total_messages = len(messages)
    print(f"Total messages parsed: {total_messages}")
    
    if total_messages == 0:
        print("No messages found. Exiting.")
        return
    
    # Deduplicate messages
    unique_messages, duplicates = deduplicate_messages(messages)
    unique_count = len(unique_messages)
    print(f"Unique messages after deduplication: {unique_count}")
    print(f"Duplicate messages removed: {duplicates}")
    if total_messages > 0:
        print(f"Deduplication ratio: {duplicates/total_messages*100:.2f}%")
    
    # Extract sizes for each serialization library
    serialization_libs = ['protobuf', 'flatbuffers', 'capnproto', 'symphony']
    sizes_by_lib = defaultdict(list)
    
    for msg in unique_messages:
        sizes = msg.get('sizes', {})
        for lib in serialization_libs:
            if lib in sizes:
                sizes_by_lib[lib].append(sizes[lib])
    
    # Calculate CDFs
    cdfs = {}
    for lib in serialization_libs:
        if lib in sizes_by_lib:
            sizes, probs = calculate_cdf(sizes_by_lib[lib])
            cdfs[lib] = (sizes, probs)
            print(f"{lib}: {len(sizes)} messages, min={min(sizes) if sizes else 0}, "
                  f"max={max(sizes) if sizes else 0}, "
                  f"median={sorted(sizes)[len(sizes)//2] if sizes else 0}")
    
    # Plot CDFs
    fig, ax = plt.subplots(1, 1, figsize=(4, 3))
    
    # Standard SIGCOMM Color Palette & Styles
    colors = ['#6acc64', '#4878d0', '#82c6e2', '#e6a04e']
    linestyles = ['-', '--', '-.', ':']
    
    # Map libraries to colors/linestyles
    lib_order = ['protobuf', 'flatbuffers', 'capnproto', 'symphony']
    lib_labels = {'protobuf': 'Protobuf', 'flatbuffers': 'Flatbuffers', 
                  'capnproto': 'Cap\'n Proto', 'symphony': 'Symphony'}
    
    for i, lib in enumerate(lib_order):
        if lib in cdfs:
            sizes, probs = cdfs[lib]
            # Sizes and probs are already sorted from calculate_cdf
            ax.plot(sizes, probs,
                     label=lib_labels.get(lib, lib.capitalize()),
                     color=colors[i % len(colors)],
                     linestyle=linestyles[i % len(linestyles)],
                     linewidth=2.5)
    
    # Styling
    ax.set_yticks([0, 0.25, 0.50, 0.75, 1.0])
    ax.set_yticklabels(['0', '25', '50', '75', '100'])
    ax.set_ylabel('CDF (%)')
    ax.set_xlabel('Message Size (bytes)', fontsize=14)
    ax.set_xscale('log')
    ax.grid(True, which="major", ls="-", alpha=0.3)
    ax.set_ylim(0, 1)
    
    # Legend at bottom
    handles, labels = ax.get_legend_handles_labels()
    fig.legend(handles, labels,
               loc='lower center',
               bbox_to_anchor=(0.5, -0.15),
               ncol=4,
               frameon=True,
               columnspacing=1.5)
    
    # Save plot
    output_file = script_dir / 'cdf_message_sizes.png'
    plt.tight_layout()
    plt.subplots_adjust(bottom=0.25)
    plt.savefig(output_file, bbox_inches='tight', dpi=300)
    print(f"\nCDF plot saved to: {output_file}")
    
    plt.close()


if __name__ == '__main__':
    main()

