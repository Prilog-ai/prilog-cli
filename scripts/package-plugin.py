#!/usr/bin/env python3
"""Validate and package the shared Prilog plugin without bundling the CLI repo."""
import argparse
import json
import re
import struct
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CLAUDE_ROOT = ROOT / 'plugins/prilog'

def load(path):
    return json.loads(path.read_text())

def claude_files(manifest, mcp):
    interface = manifest['extensions']['com.openai']['interface']
    metadata = {key: manifest[key] for key in ('name', 'version', 'description', 'author', 'homepage', 'repository', 'license', 'keywords')}
    metadata.update({
        # Claude's listing uses the wordmark; OpenAI requires a square icon.
        'icon': './assets/logo.png',
        'documentationUrl': 'https://github.com/Prilog-ai/prilog-cli#agent-plugin-and-automation',
        'supportUrl': interface['supportURL'],
        'privacyPolicyUrl': interface['privacyPolicyURL'],
        'termsOfServiceUrl': interface['termsOfServiceURL'],
    })
    return {
        '.claude-plugin/plugin.json': metadata,
        '.mcp.json': {'mcpServers': {name: {'type': 'http', 'url': config['url']} for name, config in mcp['mcpServers'].items()}},
    }

def encode(value):
    return (json.dumps(value, indent=2) + '\n').encode()

def shared_files():
    files = {'LICENSE': (ROOT / 'LICENSE').read_bytes()}
    for directory in ('skills', 'assets'):
        for path in sorted((ROOT / directory).rglob('*')):
            if path.is_file() and path.name != '.DS_Store':
                files[path.relative_to(ROOT).as_posix()] = path.read_bytes()
    return files

def openai_files(manifest, mcp, files):
    interface = manifest['extensions']['com.openai']['interface']
    assets = set()
    for field in ('composerIcon', 'composerIconDark', 'logo', 'logoDark'):
        if field in interface:
            value = interface[field]
            assert value.startswith('./'), f'{field} must be plugin-relative'
            path = ROOT / value
            path.resolve().relative_to(ROOT)
            data = path.read_bytes()
            assert data[:8] == b'\x89PNG\r\n\x1a\n' and data[12:16] == b'IHDR', f'{field} must reference a PNG icon'
            width, height = struct.unpack('>II', data[16:24])
            assert width == height and 48 <= width <= 4096, f'{field} must be square, 48 to 4096 pixels'
            assert len(data) <= 5 * 1024 * 1024, f'{field} exceeds 5 MiB'
            assets.add(path.relative_to(ROOT).as_posix())
    for value in interface.get('screenshots', []):
        assert value.startswith('./'), 'Screenshot paths must be plugin-relative'
        path = ROOT / value
        path.resolve().relative_to(ROOT)
        assert path.is_file(), value
        assets.add(path.relative_to(ROOT).as_posix())
    package = {path: data for path, data in files.items() if not path.startswith('assets/') or path in assets}
    for path in assets:
        package[path] = (ROOT / path).read_bytes()
    review = manifest['extensions']['com.openai']['review']['test_cases']
    assert len(review['positive']) == 5 and len(review['negative']) == 3, 'OpenAI MCP review requires five positive and three negative cases'
    for case in review['positive']:
        assert all(case.get(field) for field in ('description', 'prompt', 'tools_triggered', 'expected_behavior')), 'Incomplete positive review case'
    return package | {'plugin.json': encode(manifest), 'mcp.json': encode(mcp)}

def check_claude_directory(files):
    paths = list(CLAUDE_ROOT.rglob('*'))
    assert not any(path.is_symlink() for path in paths), 'Claude submission must not contain symlinks'
    actual = {path.relative_to(CLAUDE_ROOT).as_posix() for path in paths if path.is_file()}
    assert actual == set(files), f'Regenerate plugins/prilog; missing: {set(files) - actual}, unexpected: {actual - set(files)}'
    for path, content in files.items():
        assert (CLAUDE_ROOT / path).read_bytes() == content, f'Regenerate plugins/prilog/{path}'

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
    parser.add_argument('--check', action='store_true', help='Validate sources, generated Claude files, and the submission directory without writing')
    parser.add_argument('--output', type=Path, default=ROOT / 'dist/plugin')
    parser.add_argument('--platform', choices=('all', 'openai', 'claude'), default='all', help='Build or check one platform without regenerating the other submission')
    args = parser.parse_args()
    manifest, mcp = load(ROOT / 'plugin.json'), load(ROOT / 'mcp.json')
    validate(manifest, mcp)
    generated = claude_files(manifest, mcp)
    files = shared_files()
    packages = {}
    if args.platform in ('all', 'openai'):
        packages['openai'] = openai_files(manifest, mcp, files)
    if args.platform == 'openai':
        if args.check:
            print('OpenAI plugin schemas, square icons, skill resources, and review cases validated')
        else:
            write_packages(args.output, manifest['version'], packages)
        return
    claude_package = files | {path: encode(value) for path, value in generated.items()}
    claude_package['README.md'] = (ROOT / 'scripts/plugin-readme.md').read_bytes()
    if args.check:
        for path, content in generated.items():
            assert (ROOT / path).read_bytes() == encode(content), f'Regenerate {path}'
        check_claude_directory(claude_package)
        print('Prilog plugin schemas, skill resources, Claude manifests, and submission directory validated')
        return
    for path, content in generated.items():
        target = ROOT / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(encode(content))
    for path, content in claude_package.items():
        target = CLAUDE_ROOT / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(content)
    check_claude_directory(claude_package)
    packages['claude'] = claude_package
    write_packages(args.output, manifest['version'], packages)

def write_packages(output, version, packages):
    output.mkdir(parents=True, exist_ok=True)
    for platform, package in packages.items():
        target = output / f'prilog-{platform}-{version}.zip'
        with zipfile.ZipFile(target, 'w', compression=zipfile.ZIP_DEFLATED) as archive:
            for path, data in sorted(package.items()):
                info = zipfile.ZipInfo(path, date_time=(2026, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = 0o100644 << 16
                archive.writestr(info, data)
        print(target)

if __name__ == '__main__':
    main()
