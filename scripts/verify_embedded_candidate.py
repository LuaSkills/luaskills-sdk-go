"""Run real cgo acceptance against one explicitly frozen local candidate, without downloads.
针对一个显式冻结的本地候选库执行真实 cgo 验收，不下载任何资产。
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

from verify_embedded_distribution import module_sources
from embedded_module_archive import MODULE_PATH, consume_zip, validate_zip


def digest(path: Path) -> str:
    """Return SHA-256 of the exact regular file selected by path.
    返回 path 所选精确普通文件的 SHA-256。
    """
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(arguments: argparse.Namespace) -> None:
    """Validate explicit candidate inputs, run cgo tests and the lifecycle example, and retain evidence.
    校验显式候选输入，执行 cgo 测试及生命周期示例，并保留证据。
    arguments supplies the installed toolchain, frozen library/hash/description, and optional race mode.
    arguments 提供已安装工具链、冻结库／摘要／描述，以及可选竞争检测模式。
    """
    # Validate before invoking Go: a missing library must never become a skipped native test.
    # 调用 Go 前先校验：缺少库绝不能转为跳过原生测试。
    root = Path(__file__).resolve().parents[1]
    library = Path(arguments.library).resolve(strict=True)
    description = Path(arguments.description).resolve(strict=True)
    if not library.is_file() or not description.is_file():
        raise ValueError("Candidate library and description must be regular files")
    if not re.fullmatch(r"[0-9a-f]{64}", arguments.library_sha256):
        raise ValueError("--library-sha256 must be an explicit lowercase SHA-256")
    if digest(library) != arguments.library_sha256:
        raise ValueError("Candidate library SHA-256 mismatch")
    # Go's existing cgo directives select exactly these platform library names.
    # Go 现有 cgo 指令精确选择这些平台库名称。
    names = {"win32": "luaskills.dll", "linux": "libluaskills.so", "darwin": "libluaskills.dylib"}
    if sys.platform not in names or library.name != names[sys.platform]:
        raise ValueError("Candidate library name does not match the current cgo platform")
    expected = json.loads(description.read_text(encoding="utf-8"))
    if not isinstance(expected, dict):
        raise ValueError("--description must contain the raw frozen CoreDescription JSON object")
    # ZIP mode requires all exact frozen inputs; development source replacement requires its own explicit switch.
    # ZIP 模式要求全部精确冻结输入；开发源码替换要求独立显式开关。
    zip_inputs = (arguments.module_zip, arguments.module_zip_sha256, arguments.module_version)
    if any(zip_inputs) and (not all(zip_inputs) or arguments.source_development):
        raise ValueError("Provide all frozen module ZIP inputs, or choose --source-development exclusively")
    if not any(zip_inputs) and not arguments.source_development:
        raise ValueError("Choose exact frozen module ZIP inputs or explicit --source-development")
    if all(zip_inputs):
        if not Path(arguments.module_zip).is_absolute():
            raise ValueError("--module-zip must be an absolute frozen archive path")
        validate_zip(Path(arguments.module_zip), arguments.module_version, arguments.module_zip_sha256)
    # The local cache and offline proxy forbid implicit toolchain or dependency acquisition.
    # 本地缓存及离线代理禁止隐式获取工具链或依赖。
    environment = dict(os.environ, CGO_ENABLED="1", LUASKILLS_NATIVE_E2E="1", GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off", GOWORK="off", GOFLAGS="")
    for name in ("CGO_LDFLAGS", "CGO_LDFLAGS_ALLOW", "CGO_LDFLAGS_DISALLOW", "LIBRARY_PATH", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH", "DYLD_FALLBACK_LIBRARY_PATH"):
        environment.pop(name, None)
    (root / ".temp").mkdir(exist_ok=True)
    work = Path(tempfile.mkdtemp(prefix="embedded-candidate-", dir=root / ".temp"))
    # module identity has one authority shared with ZIP freezing and proxy import verification.
    # module 身份与 ZIP 冻结及代理导入校验共享唯一权威。
    module = MODULE_PATH
    if all(zip_inputs):
        candidate, consumer, sources, module_evidence = consume_zip(arguments.go, Path(arguments.module_zip), arguments.module_version, arguments.module_zip_sha256, work, environment)
    else:
        # Source mode is explicitly development-only; it never produces frozen-module ZIP success evidence.
        # 源码模式明确仅用于开发，绝不产生冻结模块 ZIP 成功证据。
        sources = module_sources(root)
        candidate = work / "module"
        for name, contents in sources.items():
            member = candidate / name
            member.parent.mkdir(parents=True, exist_ok=True)
            member.write_bytes(contents)
            if member.read_bytes() != contents:
                raise ValueError(f"Candidate module bytes changed: {name}")
        consumer = work / "consumer"
        consumer.mkdir()
        (consumer / "go.mod").write_text(f'module embedded-candidate-consumer\n\ngo 1.22\n\nrequire {module} v0.0.0\n\nreplace {module} => "{candidate.as_posix()}"\n', encoding="utf-8")
        environment["GOMODCACHE"] = str(work / "cache")
        module_evidence = {"mode": "source-development", "module": str(candidate), "module_members": len(sources), "replace": True}
    # Private staging leaves a single SDK library in the link directory and beside Windows executables.
    # 私有暂存使链接目录中仅有一个 SDK 库，并将 Windows 可执行文件置于该库旁。
    staged = work / library.name
    shutil.copyfile(library, staged)
    frozen = work / "core-description.json"
    frozen.write_bytes(description.read_bytes())
    if digest(staged) != arguments.library_sha256:
        raise ValueError("Staged candidate library SHA-256 mismatch")
    environment["CGO_LDFLAGS"] = '"-L' + work.as_posix() + '"'
    environment["PATH"] = str(work) + os.pathsep + environment["PATH"]
    if sys.platform == "linux":
        environment["LD_LIBRARY_PATH"] = str(work)
        environment["CGO_LDFLAGS"] += ' "-Wl,-rpath,' + work.as_posix() + '"'
    elif sys.platform == "darwin":
        environment["DYLD_LIBRARY_PATH"] = str(work)
    environment["LUASKILLS_CANDIDATE_LIBRARY"] = str(staged)
    environment["LUASKILLS_CANDIDATE_SHA256"] = arguments.library_sha256
    environment["LUASKILLS_CANDIDATE_DESCRIPTION"] = str(frozen)
    # Compile once into the staging directory, then execute that exact binary through test2json.
    # 仅编译一次到暂存目录，再通过 test2json 执行该精确二进制。
    suffix = ".exe" if sys.platform == "win32" else ""
    binary = work / ("embedded.test" + suffix)
    race = ["-race"] if arguments.race else []
    subprocess.run([arguments.go, "run", "./scripts/generate-embedded-contract", "--check"], cwd=candidate, env=environment, check=True)
    subprocess.run([arguments.go, "test", "-p", "4", *race, "-c", "-o", str(binary), module], cwd=consumer, env=environment, check=True)
    # Native tests read packaged fixtures from the candidate copy, matching ordinary go test's package directory.
    # 原生测试从候选副本读取包内夹具，与普通 go test 的包目录一致。
    result = subprocess.run([arguments.go, "tool", "test2json", "-t", "-p", module, str(binary), "-test.v=test2json", "-test.run=^TestEmbedded", "-test.parallel=1", "-test.timeout=" + arguments.timeout], cwd=candidate, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    report = result.stdout.decode("utf-8")
    (work / "tests.jsonl").write_text(report, encoding="utf-8")
    # Every selected native test must run; the special evidence test must prove the linked description.
    # 每个选定原生测试都必须执行；专用证据测试必须证明已链接库的描述。
    counts = {"pass": 0, "fail": 0, "skip": 0}
    evidence_passed = False
    for line in report.splitlines():
        if not line.startswith("{"):
            continue
        event = json.loads(line)
        if "Test" in event and event["Action"] in counts:
            counts[event["Action"]] += 1
        if event.get("Test") == "TestEmbeddedCandidateEvidence" and event["Action"] == "pass":
            evidence_passed = True
    if result.returncode or counts["fail"] or counts["skip"] or not evidence_passed:
        # Keep every event: the failing test can precede many later passing tests.
        # 保留全部事件：失败测试可能位于大量后续通过测试之前。
        raise RuntimeError(f"Candidate tests failed; evidence: {work}\n{report}")
    # The lifecycle regression exercises cancelled post-admission ownership against this same candidate module/library.
    # 生命周期回归针对同一个候选模块／库验证入场后取消的所有权。
    example_tests = work / ("embedded-lifecycle.test" + suffix)
    subprocess.run([arguments.go, "test", "-p", "4", *race, "-c", "-o", str(example_tests), module + "/examples/embedded_lifecycle"], cwd=consumer, env=environment, check=True)
    example_test_result = subprocess.run([arguments.go, "tool", "test2json", "-t", "-p", module + "/examples/embedded_lifecycle", str(example_tests), "-test.v=test2json", "-test.parallel=1", "-test.timeout=" + arguments.timeout], cwd=candidate, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    (work / "example-tests.jsonl").write_bytes(example_test_result.stdout)
    # lifecycle_counts keeps this independent example regression visible without conflating root package results.
    # lifecycle_counts 单独展示此示例回归，不混淆根包结果。
    lifecycle_counts = {"pass": 0, "fail": 0, "skip": 0}
    for line in example_test_result.stdout.decode("utf-8").splitlines():
        if line.startswith("{"):
            event = json.loads(line)
            if "Test" in event and event["Action"] in lifecycle_counts:
                lifecycle_counts[event["Action"]] += 1
    if example_test_result.returncode or lifecycle_counts["fail"] or lifecycle_counts["skip"] or not lifecycle_counts["pass"]:
        # Preserve complete lifecycle diagnostics through the same failed-child transport.
        # 经同一失败子进程运输路径保留完整生命周期诊断。
        raise RuntimeError(f"Candidate lifecycle tests failed; evidence: {work}\n{example_test_result.stdout.decode('utf-8')}")
    example = work / ("embedded-lifecycle" + suffix)
    subprocess.run([arguments.go, "build", "-p", "4", *race, "-o", str(example), module + "/examples/embedded_lifecycle"], cwd=consumer, env=environment, check=True)
    # Preserve actual example output beside test events; a timeout remains a failed ownership observation.
    # 在测试事件旁保存真实示例输出；超时仍属于失败的所有权观察。
    example_result = subprocess.run([str(example)], cwd=work, env=environment, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=120)
    (work / "example.log").write_bytes(example_result.stdout)
    print(example_result.stdout.decode("utf-8"), end="")
    example_result.check_returncode()
    if digest(staged) != arguments.library_sha256 or digest(library) != arguments.library_sha256:
        raise ValueError("Candidate library changed during verification")
    for name, contents in sources.items():
        if (candidate / name).read_bytes() != contents:
            raise ValueError(f"Candidate module changed during verification: {name}")
    # The successful report records only proven outcomes and remains durable with the exact tested module.
    # 成功报告仅记录已证明结果，并与精确已测试模块一起持久保留。
    summary = {"library": str(library), "sha256": arguments.library_sha256, "description_sha256": digest(frozen), "module_members": len(sources), "module": str(candidate), "race": arguments.race, "tests": counts, "lifecycle_tests": lifecycle_counts, "evidence": str(work)}
    summary.update(module_evidence)
    (work / "validation.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, ensure_ascii=False))


if __name__ == "__main__":
    # Required frozen inputs keep this gate independent of remote artifacts and release defaults.
    # 必填冻结输入使此门禁独立于远端资产及发布默认值。
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default="go", help="Installed Go executable; no toolchain download")
    parser.add_argument("--library", required=True, help="Exact local candidate shared-library file")
    parser.add_argument("--library-sha256", required=True, help="Frozen lowercase SHA-256 of the shared library")
    parser.add_argument("--description", required=True, help="Frozen raw CoreDescription JSON file")
    parser.add_argument("--module-zip", help="Absolute frozen Go module ZIP path")
    parser.add_argument("--module-zip-sha256", help="Frozen module ZIP lowercase SHA-256")
    parser.add_argument("--module-version", help="Exact frozen Go module version")
    parser.add_argument("--source-development", action="store_true", help="Explicit development source-copy/replace mode, not ZIP acceptance")
    parser.add_argument("--race", action="store_true", help="Enable Go race detection for tests and example")
    parser.add_argument("--timeout", default="10m", help="Go native test timeout")
    arguments = parser.parse_args()
    verify(arguments)
