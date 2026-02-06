#!/usr/bin/env python3
"""Extract all symphony message sizes from log files and save to a txt file."""

import json
from pathlib import Path

def extract_symphony_sizes(logs_dir: Path, min_size: int = 8) -> list[int]:
    """Extract all symphony sizes from log files.
    
    Args:
        logs_dir: Directory containing .jsonl log files
        min_size: Minimum size to include (exclusive of sizes less than this)
    
    Returns:
        List of symphony sizes
    """
    sizes = []
    
    for log_file in sorted(logs_dir.glob('*.jsonl')):
        with open(log_file, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                
                try:
                    data = json.loads(line)
                    symphony_size = data.get('sizes', {}).get('symphony', 0)
                    
                    # Exclude messages less than min_size bytes
                    if symphony_size >= min_size:
                        sizes.append(symphony_size)
                except json.JSONDecodeError:
                    continue
    
    return sizes

def main():
    script_dir = Path(__file__).parent.parent
    logs_dir = script_dir / 'logs'
    output_file = script_dir / 'symphony_sizes.txt'
    
    print(f"Extracting symphony sizes from {logs_dir}...")
    sizes = extract_symphony_sizes(logs_dir, min_size=12)
    
    with open(output_file, 'w') as f:
        for size in sizes:
            f.write(f"{size}\n")
    
    print(f"Extracted {len(sizes)} message sizes (excluding sizes < 12 bytes)")
    print(f"Saved to {output_file}")
    
    # Print some stats
    if sizes:
        print(f"\nStats:")
        print(f"  Min: {min(sizes)}")
        print(f"  Max: {max(sizes)}")
        print(f"  Avg: {sum(sizes) / len(sizes):.1f}")

if __name__ == '__main__':
    main()
