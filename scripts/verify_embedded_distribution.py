"""Verify a real module ZIP through a private file proxy and an empty module cache.
通过私有文件代理及空模块缓存验证真实模块 ZIP。
"""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

from embedded_module_archive import MODULE_PATH, consume_zip, require_members


def module_sources(root: Path) -> dict[str, bytes]:
    """Return exact Git-listed package source bytes from root, rejecting missing or escaped members.
    返回 root 中 Git 所列包源码的精确字节，拒绝缺失或逃逸成员。
    """
    # The source list excludes ignored local tools, caches and verification archives.
    # 源码清单排除忽略的本地工具、缓存及验证归档。
    listing = subprocess.check_output(["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"], cwd=root)
    names = sorted(set(name.decode("utf-8") for name in listing.split(b"\0") if name))
    # Only regular in-repository files become archive members; no external symlink targets are borrowed.
    # 仅将仓库内普通文件纳入归档；不借用外部符号链接目标。
    sources = {}
    for name in names:
        source = root / name
        if source.is_symlink() or not source.is_file() or not source.resolve().is_relative_to(root):
            raise ValueError(f"Invalid module source member: {name}")
        sources[name] = source.read_bytes()
    require_members(sources)
    return sources


def verify(go: str, module_zip: str | None = None, module_zip_sha256: str | None = None, module_version: str | None = None) -> None:
    """Build exact current source ZIP and run offline packaged tests with the installed Go executable.
    构建当前源码精确 ZIP，并使用已安装 Go 可执行文件执行离线包内测试。
    go selects the toolchain; native acceptance belongs to verify_embedded_candidate.py.
    go 选择工具链；原生验收由 verify_embedded_candidate.py 负责。
    """
    # root/work keep both default freezing and explicitly supplied ZIP consumption isolated from the current module cache.
    # root／work 使默认冻结及显式 ZIP 消费均隔离于当前模块缓存。
    root = Path(__file__).resolve().parents[1]
    if any((module_zip, module_zip_sha256, module_version)) and not all((module_zip, module_zip_sha256, module_version)):
        raise ValueError("Frozen module ZIP, SHA-256 and exact version must be supplied together")
    work = Path(tempfile.mkdtemp(prefix="module-distribution-", dir=root / ".temp"))
    if module_zip is None:
        # Local import keeps source listing independent from the freeze command's use of that same listing.
        # 本地导入使源码列举独立于冻结命令对同一列举的使用。
        from freeze_embedded_module import freeze
        frozen = freeze(root, work / "frozen")
        module_zip, module_zip_sha256, module_version = frozen["module_zip"], frozen["module_zip_sha256"], frozen["module_version"]
    # environment disables native work; consumption still proves exact ZIP ownership through Go's real file proxy.
    # environment 禁止原生工作；消费仍通过 Go 真实文件代理证明精确 ZIP 归属。
    environment = dict(os.environ, CGO_ENABLED="0", LUASKILLS_NATIVE_E2E="0")
    directory, consumer, sources, evidence = consume_zip(go, Path(module_zip), module_version, module_zip_sha256, work, environment)
    # package_path is the downloaded Go module's declared path, distinct from its opaque version.
    # package_path 是已下载 Go 模块声明路径，与其不透明版本分开。
    package_path = MODULE_PATH
    # Generation must work from the module's packaged contract with no adjacent core checkout.
    # 生成必须使用模块包内契约，不依赖相邻核心检出。
    subprocess.run([go, "run", "./scripts/generate-embedded-contract", "--check"], cwd=directory, env=environment, check=True)
    result = subprocess.run([go, "test", "-p", "4", "-parallel", "4", "-count=1", "-json", "-run", "^TestEmbedded", package_path, package_path + "/scripts/generate-embedded-contract"], cwd=consumer, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    report = result.stdout.decode("utf-8", errors="strict")
    (work / "tests.jsonl").write_text(report, encoding="utf-8")
    counts = {"pass": 0, "fail": 0, "skip": 0}
    for line in report.splitlines():
        if line.startswith("{"):
            event = json.loads(line)
            if "Test" in event and event["Action"] in counts:
                counts[event["Action"]] += 1
    if result.returncode or counts["fail"] or not counts["pass"]:
        raise RuntimeError(report[-12000:])
    evidence.update(tests=counts, evidence=str(work), native=False)
    (work / "validation.json").write_text(json.dumps(evidence, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(evidence, ensure_ascii=False))


if __name__ == "__main__":
    # The default resolves the caller's installed Go; callers may pass an exact minimum-version executable.
    # 默认解析调用方已安装 Go；调用方可传入精确最低版本可执行文件。
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default="go")
    parser.add_argument("--module-zip")
    parser.add_argument("--module-zip-sha256")
    parser.add_argument("--module-version")
    arguments = parser.parse_args()
    (Path(__file__).resolve().parents[1] / ".temp").mkdir(exist_ok=True)
    verify(arguments.go, arguments.module_zip, arguments.module_zip_sha256, arguments.module_version)
