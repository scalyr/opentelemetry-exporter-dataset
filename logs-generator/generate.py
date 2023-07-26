import argparse
import random
import sys
import time

parser = argparse.ArgumentParser(description='Generate logs', prog='generate')
parser.add_argument('-n', '--lines', type=int, default=100, help='number of lines', dest='lines')
parser.add_argument('-m', '--line-length-mu', type=int, default=100, help="length of the line - mean", dest='length_mu')
parser.add_argument('-s', '--line-length-sigma', type=int, default=10, help="length of the line - sigma", dest='length_sigma')
parser.add_argument('-r', '--seed', type=int, default=int(time.time()), help="seed for random generator", dest='seed')
args = parser.parse_args()

def g(i, n):
    print(f"Generated {i:7d} / {n:7d}", file=sys.stderr)

def main():
    print(
        f"Generate logs - "
        f"lines: {args.lines}, "
        f"lines-mu: {args.length_mu}, "
        f"lines-sigma: {args.length_sigma}, "
        f"seed: {args.seed}",
        file=sys.stderr
    )
    for i in range(args.lines):
        if i % 100000 == 0:
            g(i, args.lines)
        line_length = int(random.gauss(mu=args.length_mu, sigma=args.length_sigma))
        prefix = f"{line_length} "
        print(f"{prefix}" + "A" * (line_length - len(prefix)))

    g(args.lines, args.lines)
