import os, sys

base = r'G:\codex-AI-tools\zero'

files = {}

files['doc/philosophy.md'] = '''# Zero Language Philosophy

## Zero doesn''t add features. Zero removes problems.

Every programming language ever created solved some problems and introduced new ones. Zero is a systematic audit.'''

for path, content in files.items():
    full = os.path.join(base, path)
    os.makedirs(os.path.dirname(full), exist_ok=True)
    with open(full, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f'Written: {full}')