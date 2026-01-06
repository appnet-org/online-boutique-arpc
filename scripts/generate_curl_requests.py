#!/usr/bin/python
#
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Generate curl requests for Online Boutique application based on locust workload pattern.

This script generates 1000 (or custom number of) curl requests that simulate user behavior
matching the weighted task distribution from the original locust script:
- index: 1x weight
- setCurrency: 2x weight
- browseProduct: 10x weight
- addToCart: 2x weight
- viewCart: 3x weight
- checkout: 1x weight

Usage Examples:
    # Generate 1000 curl commands and print examples (default behavior)
    python scripts/generate_curl_requests.py

    # Generate and save all curl commands to a shell script
    python scripts/generate_curl_requests.py --output curl_commands.sh

    # Execute all curl commands directly (sends requests to server)
    python scripts/generate_curl_requests.py --execute

    # Generate 5000 requests with custom base URL
    python scripts/generate_curl_requests.py --num-requests 5000 --base-url http://localhost:8080

    # Generate and save to file, then execute
    python scripts/generate_curl_requests.py --output requests.sh --execute

    # Generate requests for a different server
    python scripts/generate_curl_requests.py --base-url http://example.com:80

    # Generate only 100 requests for quick testing
    python scripts/generate_curl_requests.py --num-requests 100 --execute

Arguments:
    --num-requests N    Number of requests to generate (default: 1000)
    --base-url URL      Base URL for the application (default: http://10.96.88.88)
    --execute           Execute the curl commands (default: just print them)
    --output FILE       Output file to save curl commands (optional)

Dependencies:
    - faker: pip install faker
"""

import random
from faker import Faker
import datetime
import subprocess
import sys

fake = Faker()

# Base URL - adjust if needed
BASE_URL = "http://10.96.88.88"

products = [
    '0PUK6V6EV0',
    '1YMWWN1N4O',
    '2ZYFJ3GM2N',
    '66VCHSJNUP',
    '6E92ZMYYFZ',
    '9SIQT8TOJO',
    'L9ECAV7KIM',
    'LS4PSXUNUM',
    'OLJCESPC7Z']

currencies = ['EUR', 'USD', 'JPY', 'CAD', 'GBP', 'TRY']

def generate_index_curl():
    """Generate curl command for index page"""
    return f'curl -s "{BASE_URL}/"'

def generate_set_currency_curl():
    """Generate curl command for setting currency"""
    currency = random.choice(currencies)
    return f'curl -s -X POST "{BASE_URL}/setCurrency" -d "currency_code={currency}"'

def generate_browse_product_curl():
    """Generate curl command for browsing a product"""
    product = random.choice(products)
    return f'curl -s "{BASE_URL}/product/{product}"'

def generate_view_cart_curl():
    """Generate curl command for viewing cart"""
    return f'curl -s "{BASE_URL}/cart"'

def generate_add_to_cart_curl():
    """Generate curl command for adding item to cart"""
    product = random.choice(products)
    quantity = random.randint(1, 10)
    # First view the product, then add to cart
    curl1 = f'curl -s "{BASE_URL}/product/{product}"'
    curl2 = f'curl -s -X POST "{BASE_URL}/cart" -d "product_id={product}" -d "quantity={quantity}"'
    return [curl1, curl2]

def generate_empty_cart_curl():
    """Generate curl command for emptying cart"""
    return f'curl -s -X POST "{BASE_URL}/cart/empty"'

def generate_checkout_curl():
    """Generate curl command for checkout"""
    product = random.choice(products)
    quantity = random.randint(1, 10)
    current_year = datetime.datetime.now().year + 1
    
    # First add to cart
    curl1 = f'curl -s "{BASE_URL}/product/{product}"'
    curl2 = f'curl -s -X POST "{BASE_URL}/cart" -d "product_id={product}" -d "quantity={quantity}"'
    
    # Then checkout
    email = fake.email()
    street_address = fake.street_address().replace('"', '\\"')
    zip_code = fake.zipcode()
    city = fake.city().replace('"', '\\"')
    state = fake.state_abbr()
    country = fake.country().replace('"', '\\"')
    credit_card = fake.credit_card_number(card_type="visa")
    exp_month = random.randint(1, 12)
    exp_year = random.randint(current_year, current_year + 70)
    cvv = f"{random.randint(100, 999)}"
    
    curl3 = (f'curl -s -X POST "{BASE_URL}/cart/checkout" '
             f'-d "email={email}" '
             f'-d "street_address={street_address}" '
             f'-d "zip_code={zip_code}" '
             f'-d "city={city}" '
             f'-d "state={state}" '
             f'-d "country={country}" '
             f'-d "credit_card_number={credit_card}" '
             f'-d "credit_card_expiration_month={exp_month}" '
             f'-d "credit_card_expiration_year={exp_year}" '
             f'-d "credit_card_cvv={cvv}"')
    
    return [curl1, curl2, curl3]

def generate_logout_curl():
    """Generate curl command for logout"""
    return f'curl -s "{BASE_URL}/logout"'

def generate_requests(num_requests=1000):
    """Generate a list of curl commands based on weighted task distribution"""
    # Task weights from the original locust script
    task_weights = {
        'index': 1,
        'setCurrency': 2,
        'browseProduct': 10,
        'addToCart': 2,
        'viewCart': 3,
        'checkout': 1
    }
    
    total_weight = sum(task_weights.values())
    
    # Calculate number of requests for each task type
    tasks = []
    for task_name, weight in task_weights.items():
        count = int((weight / total_weight) * num_requests)
        tasks.extend([task_name] * count)
    
    # Add remaining requests to reach exactly num_requests
    remaining = num_requests - len(tasks)
    if remaining > 0:
        tasks.extend(['browseProduct'] * remaining)  # Add to most common task
    
    # Shuffle tasks to simulate random user behavior
    random.shuffle(tasks)
    
    # Generate curl commands
    curl_commands = []
    for task in tasks:
        if task == 'index':
            curl_commands.append(generate_index_curl())
        elif task == 'setCurrency':
            curl_commands.append(generate_set_currency_curl())
        elif task == 'browseProduct':
            curl_commands.append(generate_browse_product_curl())
        elif task == 'addToCart':
            cmds = generate_add_to_cart_curl()
            curl_commands.extend(cmds)
        elif task == 'viewCart':
            curl_commands.append(generate_view_cart_curl())
        elif task == 'checkout':
            cmds = generate_checkout_curl()
            curl_commands.extend(cmds)
    
    return curl_commands

def main():
    """Main function to generate and optionally execute curl requests"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Generate 1000 curl requests for Online Boutique')
    parser.add_argument('--num-requests', type=int, default=1000,
                       help='Number of requests to generate (default: 1000)')
    parser.add_argument('--base-url', type=str, default='http://10.96.88.88',
                       help='Base URL for the application (default: http://10.96.88.88)')
    parser.add_argument('--execute', action='store_true',
                       help='Execute the curl commands (default: just print them)')
    parser.add_argument('--output', type=str,
                       help='Output file to save curl commands (optional)')
    
    args = parser.parse_args()
    
    global BASE_URL
    BASE_URL = args.base_url
    
    print(f"Generating {args.num_requests} curl requests...")
    curl_commands = generate_requests(args.num_requests)
    
    print(f"Generated {len(curl_commands)} curl commands")
    
    if args.output:
        with open(args.output, 'w') as f:
            for cmd in curl_commands:
                f.write(cmd + '\n')
        print(f"Curl commands saved to {args.output}")
    
    if args.execute:
        print("Executing curl commands...")
        for i, cmd in enumerate(curl_commands, 1):
            if i % 100 == 0:
                print(f"Executed {i}/{len(curl_commands)} requests...")
            try:
                subprocess.run(cmd, shell=True, capture_output=True, text=True)
            except Exception as e:
                print(f"Error executing command {i}: {e}", file=sys.stderr)
        print("All requests executed!")
    else:
        # Print first few and last few commands as examples
        print("\nFirst 5 curl commands:")
        for i, cmd in enumerate(curl_commands[:5], 1):
            print(f"{i}. {cmd}")
        
        if len(curl_commands) > 10:
            print("\n...")
            print(f"\nLast 5 curl commands:")
            for i, cmd in enumerate(curl_commands[-5:], len(curl_commands)-4):
                print(f"{i}. {cmd}")
        
        print(f"\nUse --execute to run all {len(curl_commands)} commands")
        print(f"Use --output <file> to save all commands to a file")

if __name__ == '__main__':
    main()
