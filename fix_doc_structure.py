import os
import re
import sys

def check_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    # Find function comments
    missing = []

    # We will do a full parse via Go parser later, but for now just to identify things
    pass
