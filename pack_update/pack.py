#!/usr/bin/env python3
"""
XYGo Admin 更新包打包工具

用法：
    python pack_update/pack.py 1.2.7 --title "标题" --changelog "日志1,日志2"

自动从 git diff 获取变更文件列表，打包成更新 ZIP，并自动维护 update-index.json。
"""

import os
import sys
import json
import hashlib
import zipfile
import subprocess
import argparse

SKIP_PATTERNS = [
    'node_modules', '.git/', '__pycache__', 'pack_update/', 'pack_addon/',
    '.exe', '.exe~', 'packed.go', 'pnpm-lock.yaml', 'package-lock.json',
    'cmdtools/updater/', 'cmdtools/migrate/', 'cmdtools/checktpl/',
]

ADDON_PREFIXES = ['server/addons/', 'web/src/addons/']

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
INDEX_PATH = os.path.join(SCRIPT_DIR, 'update-index.json')

def should_skip(path):
    normalized = path.replace('\\', '/')
    for p in SKIP_PATTERNS:
        if p in normalized:
            return True
    for prefix in ADDON_PREFIXES:
        if prefix in normalized:
            after = normalized[normalized.index(prefix) + len(prefix):]
            if '/' in after:
                return True
    return False

def file_hash(path):
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(8192), b''):
            h.update(chunk)
    return h.hexdigest()

def git_diff_files(tag):
    result = subprocess.run(
        ['git', 'diff', '--name-status', tag, 'HEAD'],
        capture_output=True, text=True, encoding='utf-8', errors='ignore'
    )
    added, modified, deleted = [], [], []
    for line in result.stdout.strip().split('\n'):
        if not line.strip():
            continue
        parts = line.split('\t', 1)
        if len(parts) != 2:
            continue
        status, path = parts[0].strip(), parts[1].strip()
        if should_skip(path):
            continue
        if status == 'A':
            added.append(path)
        elif status == 'M':
            modified.append(path)
        elif status == 'D':
            deleted.append(path)

    result2 = subprocess.run(
        ['git', 'ls-files', '--others', '--exclude-standard'],
        capture_output=True, text=True, encoding='utf-8', errors='ignore'
    )
    for line in result2.stdout.strip().split('\n'):
        path = line.strip()
        if path and not should_skip(path) and path not in added:
            added.append(path)

    return added, modified, deleted

def git_show_file_hash(tag, path):
    result = subprocess.run(
        ['git', 'show', f'{tag}:{path}'],
        capture_output=True
    )
    if result.returncode != 0:
        return ''
    h = hashlib.sha256()
    h.update(result.stdout)
    return h.hexdigest()

def parse_version(tag):
    """Parse version tag to comparable tuple, stripping optional 'v' prefix."""
    ver = tag.lstrip('v')
    try:
        return tuple(int(x) for x in ver.split('.'))
    except ValueError:
        return (0,)

def get_git_tags():
    result = subprocess.run(['git', 'tag', '--sort=version:refname'],
                            capture_output=True, text=True, encoding='utf-8', errors='ignore')
    tags = [t.strip() for t in result.stdout.strip().split('\n') if t.strip()]
    tags.sort(key=parse_version)
    return tags

def load_index():
    if os.path.exists(INDEX_PATH):
        with open(INDEX_PATH, 'r', encoding='utf-8') as f:
            return json.load(f)
    return {
        "_comment": "XYGo Admin 在线更新索引 - 由 pack.py 自动维护，请勿手动编辑",
        "latest": "",
        "editions": {}
    }

def save_index(index):
    with open(INDEX_PATH, 'w', encoding='utf-8') as f:
        json.dump(index, f, indent=2, ensure_ascii=False)
        f.write('\n')

def main():
    parser = argparse.ArgumentParser(description='XYGo Admin 更新包打包工具')
    parser.add_argument('version', help='目标版本号（如 1.2.7）')
    parser.add_argument('--tag', default=None, help='旧版本 git tag（默认取最新 tag）')
    parser.add_argument('--title', default='版本更新', help='更新标题')
    parser.add_argument('--changelog', default='', help='更新日志（逗号分隔）')
    parser.add_argument('--edition', default='open', help='版本类型（open/commercial）')
    parser.add_argument('--has-migration', action='store_true', help='是否包含数据库迁移')
    parser.add_argument('--output', default=SCRIPT_DIR, help='更新包 ZIP 输出目录（默认 pack_update/）')
    args = parser.parse_args()

    version = args.version

    old_tag = args.tag
    if not old_tag:
        tags = get_git_tags()
        if not tags:
            print('  错误：仓库中没有任何 tag')
            sys.exit(1)
        old_tag = tags[-1]

    from_version = old_tag.lstrip('v')
    changelog = [c.strip() for c in args.changelog.split(',') if c.strip()] if args.changelog else ['版本更新']

    print()
    print('  ╔══════════════════════════════════════╗')
    print('  ║  XYGo Admin 更新包打包工具           ║')
    print('  ╚══════════════════════════════════════╝')
    print()
    print(f'  旧版本 tag: {old_tag}')
    print(f'  新版本号:   {version}')
    print()

    # 1. 获取差异
    print('  [1/4] 获取文件差异 ... ', end='', flush=True)
    added, modified, deleted = git_diff_files(old_tag)
    changed = added + modified
    print(f'OK (新增 {len(added)}, 修改 {len(modified)}, 删除 {len(deleted)})')

    if not changed and not deleted:
        print('  无变更，不生成更新包')
        return

    for f in added[:10]:
        print(f'    [新增] {f}')
    for f in modified[:10]:
        print(f'    [修改] {f}')
    for f in deleted[:5]:
        print(f'    [删除] {f}')
    total = len(added) + len(modified) + len(deleted)
    shown = min(len(added), 10) + min(len(modified), 10) + min(len(deleted), 5)
    if total > shown:
        print(f'    ... 共 {total} 个变更')

    # checksums
    checksums = {}
    for f in modified:
        old_hash = git_show_file_hash(old_tag, f)
        if old_hash:
            checksums[f] = old_hash

    # 2. 打包 ZIP
    output_dir = os.path.abspath(args.output)
    zip_path = os.path.join(output_dir, f'{version}.zip')

    print(f'  [2/4] 打包 {version}.zip ... ', end='', flush=True)

    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zf:
        update_yaml = f'version: "{version}"\n'
        update_yaml += f'from: "{from_version}"\n'
        update_yaml += f'title: "{args.title}"\n'
        update_yaml += 'changelog:\n'
        for c in changelog:
            update_yaml += f'  - "{c}"\n'
        update_yaml += f'has_migration: {"true" if args.has_migration else "false"}\n'
        update_yaml += 'deleted_files:\n'
        if deleted:
            for d in deleted:
                update_yaml += f'  - "{d}"\n'
        else:
            update_yaml += '  []\n'

        zf.writestr(f'{version}/update.yaml', update_yaml)
        zf.writestr(f'{version}/checksums.json', json.dumps(checksums, indent=2, ensure_ascii=False))

        for rel in changed:
            abs_path = os.path.abspath(rel)
            if os.path.exists(abs_path):
                zf.write(abs_path, f'{version}/files/{rel}')

    size = os.path.getsize(zip_path)
    print(f'OK ({size:,} bytes, {len(changed)} 个文件)')

    # 3. 自动维护 update-index.json
    print('  [3/4] 更新索引 ... ', end='', flush=True)

    index = load_index()
    index['latest'] = version

    if args.edition not in index['editions']:
        index['editions'][args.edition] = {"_comment": f"{args.edition} 版更新列表", "updates": []}

    entry = {
        "version": version,
        "from": from_version,
        "title": args.title,
        "changelog": changelog,
        "url": f"https://xygoupload.xingyunwangluo.com/updates/{args.edition}/{version}.zip",
        "mirrors": [
            f"https://gitee.com/a751300685a/xygo-admin/releases/download/v{version}/update-{version}.zip"
        ],
        "has_migration": args.has_migration
    }

    updates = index['editions'][args.edition]['updates']
    updates = [u for u in updates if u.get('version') != version]
    updates.append(entry)
    index['editions'][args.edition]['updates'] = updates

    save_index(index)
    print(f'OK ({INDEX_PATH})')

    # 4. 完成
    print('  [4/4] 完成')
    print()
    print('  ════════════════════════════════════════')
    print('  打包完成！')
    print(f'  更新包:  {zip_path}')
    print(f'  索引:    {INDEX_PATH}')
    print(f'  文件数:  {len(changed)} 个变更, {len(deleted)} 个删除')
    print('  ════════════════════════════════════════')
    print()
    print('  下一步:')
    print(f'    1. 上传 pack_update/{version}.zip 到 CDN: updates/{args.edition}/{version}.zip')
    print(f'    2. 上传 pack_update/update-index.json 到 CDN: updates/update-index.json')
    print(f'    3. 推送代码和标签到远程:')
    print(f'       git push origin master')
    print(f'       git tag v{version}')
    print(f'       git push origin v{version}')

if __name__ == '__main__':
    main()
