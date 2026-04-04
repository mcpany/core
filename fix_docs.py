import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Find all public functions
    # func Name(args) ret {
    # func (recv) Name(args) ret {

    # Needs a robust parser, since ast is available in python but this is go...
    pass
