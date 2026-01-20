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


def extract_sizes_by_lib(messages, serialization_libs):
    """Extract sizes for each serialization library from messages."""
    sizes_by_lib = {}
    skipped_count = 0
    
    for msg in messages:
        sizes = msg.get('sizes', {})
        
        # Skip this message if any serialization library has size 0
        skip_message = False
        for lib in serialization_libs:
            if lib in sizes and sizes[lib] == 0:
                skip_message = True
                break
        
        if skip_message:
            skipped_count += 1
            continue
        
        # Only add sizes if no library has 0 size
        for lib in serialization_libs:
            if lib in sizes:
                if lib not in sizes_by_lib:
                    sizes_by_lib[lib] = []
                sizes_by_lib[lib].append(sizes[lib])
    
    if skipped_count > 0:
        print(f"Messages skipped due to 0-size serialization: {skipped_count}")
    
    # Convert to numpy arrays for consistency with template
    sizes_dict = {}
    lib_labels = {'protobuf': 'Protobuf', 'flatbuffers': 'Flatbuffers', 
                  'capnproto': 'Cap\'n Proto', 'symphony': 'fRPC',
                  'symphony_hybrid': 'fRPC (B-Opt)'}
    
    for lib in serialization_libs:
        if lib in sizes_by_lib and sizes_by_lib[lib]:
            sizes_dict[lib_labels.get(lib, lib.capitalize())] = np.array(sizes_by_lib[lib])
    
    return sizes_dict


def process_logs_directory(logs_dir, application_name):
    """Process logs from a directory and return sizes by library."""
    print(f"\n{'='*60}")
    print(f"Processing {application_name} logs from: {logs_dir}")
    print(f"{'='*60}")
    
    # Parse all log files
    messages = parse_logs(str(logs_dir))
    total_messages = len(messages)
    print(f"Total messages parsed: {total_messages}")
    
    if total_messages == 0:
        print(f"No messages found for {application_name}. Using empty dataset.")
        return {}
    
    # Deduplicate messages
    unique_messages, duplicates = deduplicate_messages(messages)
    unique_count = len(unique_messages)
    print(f"Unique messages after deduplication: {unique_count}")
    print(f"Duplicate messages removed: {duplicates}")
    if total_messages > 0:
        print(f"Deduplication ratio: {duplicates/total_messages*100:.2f}%")
    
    # Extract sizes for each serialization library
    serialization_libs = ['protobuf', 'flatbuffers', 'capnproto', 'symphony', 'symphony_hybrid']
    sizes_dict = extract_sizes_by_lib(unique_messages, serialization_libs)
    
    # Print statistics
    for lib, sizes in sizes_dict.items():
        if len(sizes) > 0:
            print(f"{lib}: {len(sizes)} messages, min={np.min(sizes)}, "
                  f"max={np.max(sizes)}, median={np.median(sizes):.0f}")
    
    return sizes_dict


def plot_merged_cdfs(data_left, data_right,
                    #  x_labels=('Online Boutique\nMessage Size (bytes)', 
                    #           'Hotel Reservation\nMessage Size (bytes)'),
                    x_labels=('Online Boutique (bytes)', 
                                'Hotel Reservation (bytes)'),
                     output_filename="cdf_message_sizes.pdf",
                     system_order=None):
    """
    Plots two CDFs side-by-side with shared legend at bottom.
    Titles are removed; X-axis labels differentiate the plots.
    """
    
    # 1. Setup Figure (1 row, 2 columns)
    fig, axes = plt.subplots(1, 2, figsize=(8, 3))
    
    # Standard SIGCOMM Color Palette & Styles
    colors = ['#6acc64', '#4878d0', '#82c6e2', '#e6a04e', '#d65f5f']
    linestyles = ['-', '--', '-.', ':', (0, (3, 1, 1, 1))]
    
    # Default system order
    if system_order is None:
        system_order = ['Protobuf', 'Flatbuffers', 'Cap\'n Proto', 'fRPC', 'fRPC (B-Opt)']
    
    datasets = [data_left, data_right]
    
    # 2. Loop through both subplots
    for idx, ax in enumerate(axes):
        data_dict = datasets[idx]
        
        for i, system in enumerate(system_order):
            if system not in data_dict:
                continue
            
            sorted_data = np.sort(data_dict[system])
            yvals = np.arange(1, len(sorted_data) + 1) / len(sorted_data)
            
            ax.plot(sorted_data, yvals,
                     label=system,
                     color=colors[i % len(colors)],
                     linestyle=linestyles[i % len(linestyles)],
                     linewidth=2.5)
        
        # 3. Styling
        ax.set_yticks([0, 0.25, 0.50, 0.75, 1.0])
        ax.set_yticklabels(['0', '25', '50', '75', '100'])
        
        # Y-label only on the left plot
        ax.set_ylabel('CDF (%)' if idx == 0 else "")
        
        # X-labels customized
        ax.set_xlabel(x_labels[idx], fontsize=14)
        
        ax.set_xscale('log')
        ax.grid(True, which="major", ls="-", alpha=0.3)
        ax.set_ylim(0, 1)
    
    # 4. Shared Legend at Bottom
    handles, labels = axes[0].get_legend_handles_labels()
    
    fig.legend(handles, labels,
               loc='lower center',
               bbox_to_anchor=(0.5, -0.20),
               ncol=5,
               frameon=True,
               columnspacing=1.5,
               fontsize=12)
    
    # 5. Adjust Layout
    plt.tight_layout()
    plt.subplots_adjust(bottom=0.25)
    
    print(f"\nSaving merged plot to {output_filename}...")
    plt.savefig(output_filename, bbox_inches='tight')
    plt.close()


def main():
    # Get the directory where this script is located
    script_dir = Path(__file__).parent
    
    # Define log directories
    boutique_logs_dir = script_dir / '../logs'
    boutique_logs_dir = boutique_logs_dir.resolve()
    
    # Hotel reservation logs directory (using boutique as placeholder for now)
    # hotel_logs_dir = script_dir / '../hotel_logs'
    # hotel_logs_dir = hotel_logs_dir.resolve()
    hotel_logs_dir = Path('/users/xzhu/hotel-reservation-arpc/logs_filtered')
    
    
    # Check if hotel logs directory exists, if not use boutique as placeholder
    if not hotel_logs_dir.exists():
        print(f"Hotel reservation logs directory not found at {hotel_logs_dir}")
        print("Using online boutique logs as placeholder for hotel reservation...")
        hotel_logs_dir = boutique_logs_dir
    
    # Process both log directories
    boutique_data = process_logs_directory(boutique_logs_dir, "Online Boutique")
    hotel_data = process_logs_directory(hotel_logs_dir, "Hotel Reservation")
    
    # If hotel data is empty, use boutique data as placeholder
    if not hotel_data:
        print("\nUsing online boutique data as placeholder for hotel reservation...")
        hotel_data = boutique_data.copy()
    
    # Define system order
    system_order = ['Protobuf', 'Flatbuffers', 'Cap\'n Proto', 'fRPC', 'fRPC (B-Opt)']
    
    # Plot merged CDFs
    output_file = script_dir / 'boutique_and_hotel_serialization_size_cdf.pdf'
    plot_merged_cdfs(boutique_data, hotel_data,
                     x_labels=('Online Boutique\nMessage Size (bytes)', 
                              'Hotel Reservation\nMessage Size (bytes)'),
                     output_filename=str(output_file),
                     system_order=system_order)
    
    print(f"\nCDF plot saved to: {output_file}")


if __name__ == '__main__':
    main()

