#!/usr/bin/env python3
"""Validate and package the shared Prilog plugin without bundling the CLI repo."""
import argparse
import json
import re
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def load(path):
    return json.loads(path.read_text())

def claude_files(manifest, mcp):
    return {
        '.claude-plugin/plugin.json': {key: manifest[key] for key in ('name', 'version', 'description', 'author', 'homepage', 'repository', 'license', 'keywords')},
        '.mcp.json': {'mcpServers': {name: {'type': 'http', 'url': config['url']} for name, config in mcp['mcpServers'].items()}},
    }

def encode(value):
    return (json.dumps(value, indent=2) + '\n').encode()

def validate(manifest, mcp):
    from jsonschema import Draft202012Validator
    for name, value in [('plugin', manifest), ('mcp', mcp)]:
        schema = load(ROOT / 'plugin-schemas' / f'{name}.schema.json')
        Draft202012Validator(schema).validate(value)
    assert manifest['name'] == 'prilog' and re.fullmatch(r'\d+\.\d+\.\d+', manifest['version'])
    assert mcp['mcpServers'] == {'prilog': {'type': 'streamable-http', 'url': 'https://prilog.ai/api/mcp'}}
    assert 'apps' not in manifest['extensions']['com.openai'] and 'hooks' not in manifest['extensions']['com.openai']
    interface = manifest['extensions']['com.openai']['interface']
    for field, limit in [('displayName', 30), ('shortDescription', 30), ('longDescription', 4000), ('developerName', 80)]:
        assert isinstance(interface.get(field), str) and 0 < len(interface[field]) <= limit, f'{field} must contain 1 to {limit} characters for public submission'
    skill = (ROOT / 'skills/prilog/SKILL.md').read_text()
    assert skill.startswith('---\nname: prilog\ndescription: ') and '\n---\n' in skill[4:]
    for path in ['skills/prilog/references/mcp.md', 'skills/prilog/references/cli.md', 'assets/icon.png', 'assets/logo.png', 'LICENSE']:
        assert (ROOT / path).is_file(), path
    # Every packaged path is explicitly allowed, never a working-tree archive.
    for directory in ('skills', 'assets'):
        for path in (ROOT / directory).rglob('*'):
            if path.is_file():
                assert not path.is_symlink(), path
                path.resolve().relative_to(ROOT)

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='Validate sources and generated Claude files without writing')
    parser.add_argument('--output', type=Path, default=ROOT / 'dist/plugin')
    args = parser.parse_args()
    manifest, mcp = load(ROOT / 'plugin.json'), load(ROOT / 'mcp.json')
    validate(manifest, mcp)
    generated = claude_files(manifest, mcp)
    if args.check:
        for path, content in generated.items():
            assert (ROOT / path).read_bytes() == encode(content), f'Regenerate {path}'
        print('Prilog plugin schemas, skill resources, and Claude manifests validated')
        return
    for path, content in generated.items():
        target = ROOT / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(encode(content))
    files = {'LICENSE': (ROOT / 'LICENSE').read_bytes()}
    for directory in ('skills', 'assets'):
        for path in sorted((ROOT / directory).rglob('*')):
            if path.is_file(): files[path.relative_to(ROOT).as_posix()] = path.read_bytes()
    args.output.mkdir(parents=True, exist_ok=True)
    for platform, extras in [('openai', {'plugin.json': encode(manifest), 'mcp.json': encode(mcp)}), ('claude', {path: encode(value) for path, value in generated.items()})]:
        target = args.output / f'prilog-{platform}-{manifest["version"]}.zip'
        with zipfile.ZipFile(target, 'w', compression=zipfile.ZIP_DEFLATED) as archive:
            for path, data in sorted((files | extras).items()):
                info = zipfile.ZipInfo(path, date_time=(2026, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = 0o100644 << 16
                archive.writestr(info, data)
        print(target)

if __name__ == '__main__':
    main()
