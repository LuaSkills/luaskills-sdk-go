"""Verify a real module ZIP through a private file proxy and an empty module cache.
通过私有文件代理及空模块缓存验证真实模块 ZIP。
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import zipfile


def verify(go: str) -> None:
    """Build exact current Git-listed source bytes and run packaged tests with the selected Go executable.
    构建 Git 所列当前源码的精确字节，并使用指定 Go 可执行文件运行包内测试。
    go selects an installed toolchain; compiler and native-library environment are inherited explicitly.
    go 选择已安装工具链；编译器及原生库环境显式继承。
    """
    # The source list excludes ignored local tools, caches and verification archives.
    # 源码清单排除忽略的本地工具、缓存及验证归档。
    root = Path(__file__).resolve().parents[1]
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
    for required in ("go.mod", "embedded_contract_generated.go", "embedded_ffi_cgo.go", "luaskills_ffi.h", "luaskills_json_ffi.h", "contracts/embedded/v1/contract.json"):
        if required not in sources:
            raise ValueError(f"Missing distribution member: {required}")
    # This development-only version exists solely in the private file proxy, never in a public registry.
    # 此开发专用版本仅存在于私有文件代理中，绝不进入公共注册表。
    module = "github.com/LuaSkills/luaskills-sdk-go"
    version = "v0.0.0-local.1"
    work = Path(tempfile.mkdtemp(prefix="module-distribution-", dir=root / ".temp"))
    escaped = "".join("!" + character.lower() if character.isupper() else character for character in module)
    proxy = work / "proxy"
    entries = proxy / escaped / "@v"
    entries.mkdir(parents=True)
    archive = entries / f"{version}.zip"
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as output:
        for name, contents in sources.items():
            output.writestr(f"{module}@{version}/{name}", contents)
    with zipfile.ZipFile(archive) as package:
        if len(package.namelist()) != len(set(package.namelist())):
            raise ValueError("Duplicate module ZIP member")
        for name, contents in sources.items():
            if package.read(f"{module}@{version}/{name}") != contents:
                raise ValueError(f"Module ZIP bytes changed: {name}")
    (entries / f"{version}.mod").write_bytes(sources["go.mod"])
    (entries / f"{version}.info").write_text(json.dumps({"Version": version, "Time": "2026-09-28T00:00:00Z"}), encoding="utf-8")
    (entries / "list").write_text(version + "\n", encoding="utf-8")
    consumer = work / "consumer"
    consumer.mkdir()
    (consumer / "go.mod").write_text(f"module embedded-distribution-consumer\n\ngo 1.22\n\nrequire {module} {version}\n", encoding="utf-8")
    # An empty cache and a private proxy force every SDK implementation import to come from the constructed ZIP.
    # 空缓存及私有代理强制全部 SDK 实现导入来自所构建 ZIP。
    environment = dict(os.environ, GOPROXY=proxy.as_uri(), GOMODCACHE=str(work / "cache"), GOSUMDB="off", GOTOOLCHAIN="local", GOWORK="off")
    downloaded = subprocess.check_output([go, "mod", "download", "-json", module + "@" + version], cwd=consumer, env=environment)
    manifest = json.loads(downloaded)
    directory = Path(manifest["Dir"])
    if not directory.resolve().is_relative_to(work / "cache"):
        raise ValueError("Module did not load from the isolated archive cache")
    for name, contents in sources.items():
        if (directory / name).read_bytes() != contents:
            raise ValueError(f"Downloaded module bytes changed: {name}")
    # Generation must work from the module's packaged contract with no adjacent core checkout.
    # 生成必须使用模块包内契约，不依赖相邻核心检出。
    subprocess.run([go, "run", "./scripts/generate-embedded-contract", "--check"], cwd=directory, env=environment, check=True)
    result = subprocess.run([go, "test", "-p", "4", "-count=1", "-json", "-run", "^TestEmbedded", module], cwd=consumer, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
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
    print(json.dumps({"archive": str(archive), "sha256": hashlib.sha256(archive.read_bytes()).hexdigest(), "members": len(sources), "tests": counts, "evidence": str(work)}, ensure_ascii=False))


if __name__ == "__main__":
    # The default resolves the caller's installed Go; callers may pass an exact minimum-version executable.
    # 默认解析调用方已安装 Go；调用方可传入精确最低版本可执行文件。
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default="go")
    arguments = parser.parse_args()
    (Path(__file__).resolve().parents[1] / ".temp").mkdir(exist_ok=True)
    verify(arguments.go)
