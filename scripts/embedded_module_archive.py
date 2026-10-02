"""Own frozen Go module ZIP validation and private-proxy consumption without source replacement.
负责冻结 Go 模块 ZIP 校验及私有代理消费，不替换源码。
"""
from __future__ import annotations

import hashlib
import json
from pathlib import Path, PurePosixPath
import re
import stat
import subprocess
import zipfile


# This exact module path is shared by freezing, proxy creation and native/offline candidate consumption.
# 此精确模块路径由冻结、代理创建及原生／离线候选消费共享。
MODULE_PATH = "github.com/LuaSkills/luaskills-sdk-go"
# Required public implementation, fixtures and verification members form one distribution boundary.
# 必需公开实现、夹具及验证成员构成唯一分发边界。
REQUIRED_MODULE_MEMBERS = (
    "VERSION", "go.mod", "embedded_contract_generated.go", "embedded_wire_generated.go", "embedded_wire.go",
    "embedded_driver.go", "embedded_command.go", "embedded_client.go", "embedded_handles.go", "embedded_wait.go",
    "embedded_scope.go", "embedded_scope_control.go", "embedded_callbacks.go", "embedded_pump.go",
    "embedded_pump_native.go", "embedded_pump_service.go", "embedded_ownership.go", "embedded_persistence.go",
    "embedded_compatibility.go", "embedded_compatibility_test.go", "embedded_compatibility_native_test.go",
    "embedded_capacity.go", "embedded_capacity_native_test.go", "embedded_candidate_native_test.go",
    "embedded_ffi_cgo.go", "embedded_ffi_nocgo.go", "embedded_nocgo_test.go",
    "luaskills_ffi.h", "luaskills_json_ffi.h", "contracts/embedded/v1/contract.json", "contracts/embedded/v1/contract.sha256",
    "scripts/generate-embedded-contract/wire.go", "scripts/verify_embedded_candidate.py",
    "scripts/embedded_module_archive.py", "scripts/freeze_embedded_module.py", "scripts/verify_embedded_distribution.py",
    "examples/embedded_lifecycle/main.go", "examples/embedded_lifecycle/main_native_test.go",
)


def module_info(version: str) -> bytes:
    """Return deterministic private-proxy metadata for exact version, with a bookkeeping time rather than release time.
    返回精确 version 的确定性私有代理元数据，使用记账时间而非发布时间。
    """
    return (json.dumps({"Version": version, "Time": "1970-01-01T00:00:00Z"}) + "\n").encode("utf-8")


def require_members(sources: dict[str, bytes]) -> None:
    """Require every authoritative distribution member in sources; return nothing or raise an explicit error.
    要求 sources 包含每个权威分发成员；无返回值或抛出明确错误。
    """
    for name in REQUIRED_MODULE_MEMBERS:
        if name not in sources:
            raise ValueError(f"Missing required module member: {name}")


def validate_zip(archive: Path, version: str, expected_sha256: str) -> tuple[dict[str, bytes], dict[str, object]]:
    """Read the exact archive/version/hash, rejecting escaped or missing members; return bytes and member manifest.
    读取精确 archive／version／hash，拒绝逃逸或缺失成员；返回字节及成员清单。
    This validates frozen inputs before proxy creation, module download or any compiler execution.
    此函数在代理创建、模块下载或任何编译器执行前校验冻结输入。
    """
    if not re.fullmatch(r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+incompatible)?", version):
        raise ValueError("Invalid exact Go module version")
    if not re.fullmatch(r"[0-9a-f]{64}", expected_sha256):
        raise ValueError("Frozen module ZIP SHA-256 must be lowercase hexadecimal")
    if not archive.is_file() or archive.is_symlink():
        raise ValueError("Frozen module ZIP must be a regular file")
    if hashlib.sha256(archive.read_bytes()).hexdigest() != expected_sha256:
        raise ValueError("Frozen module ZIP SHA-256 mismatch")
    # prefix identifies one precise Go module/version root; sources preserves original member bytes.
    # prefix 标识一个精确 Go 模块／版本根；sources 保留原始成员字节。
    prefix = MODULE_PATH + "@" + version + "/"
    sources = {}
    with zipfile.ZipFile(archive) as package:
        for member in package.infolist():
            # relative and parsed must retain a canonical relative file path, not a normalized escape.
            # relative 和 parsed 必须保留规范相对文件路径，而非归一化后的逃逸路径。
            if not member.filename.startswith(prefix):
                raise ValueError(f"Module ZIP version/root mismatch: {member.filename}")
            relative = member.filename[len(prefix):]
            parsed = PurePosixPath(relative)
            if not relative or relative.endswith("/") or "\\" in relative or parsed.is_absolute() or ".." in parsed.parts or ":" in relative or parsed.as_posix() != relative:
                raise ValueError(f"Invalid module ZIP member path: {relative}")
            if relative in sources:
                raise ValueError(f"Duplicate module ZIP member: {relative}")
            if stat.S_IFMT(member.external_attr >> 16) not in (0, stat.S_IFREG):
                raise ValueError(f"Non-regular module ZIP member: {relative}")
            sources[relative] = package.read(member)
    require_members(sources)
    if sources["VERSION"].decode("utf-8").strip() != version.removeprefix("v"):
        raise ValueError("Module ZIP VERSION differs from explicit module version")
    if not re.search(r"(?m)^module\s+" + re.escape(MODULE_PATH) + r"\s*$", sources["go.mod"].decode("utf-8")):
        raise ValueError("Module ZIP go.mod path mismatch")
    # manifest records each exact file independently from the complete ZIP digest.
    # manifest 独立于完整 ZIP 摘要记录每个精确文件。
    manifest = {"format_version": 1, "module": MODULE_PATH, "version": version, "zip_sha256": expected_sha256,
                "members": {name: {"sha256": hashlib.sha256(contents).hexdigest(), "size_bytes": len(contents)} for name, contents in sorted(sources.items())}}
    return sources, manifest


def consume_zip(go: str, archive: Path, version: str, sha256: str, work: Path, environment: dict[str, str]) -> tuple[Path, Path, dict[str, bytes], dict[str, object]]:
    """Download archive through a private file proxy into work's new cache, returning module/consumer/bytes/evidence.
    通过私有文件代理将 archive 下载到 work 的全新缓存，返回模块／消费方／字节／证据。
    go is installed locally; environment is updated to the exact proxy/cache with no fallback or replace directive.
    go 已在本地安装；environment 更新为精确代理／缓存，不使用回退或 replace 指令。
    """
    # sources and manifest come exclusively from the frozen, hash-verified ZIP.
    # sources 和 manifest 仅来自冻结且已校验摘要的 ZIP。
    sources, manifest = validate_zip(archive, version, sha256)
    # escaped follows Go proxy uppercase escaping for the already fixed module path.
    # escaped 对已经固定的模块路径采用 Go 代理大写转义。
    escaped = "".join("!" + character.lower() if character.isupper() else character for character in MODULE_PATH)
    proxy, cache, consumer = work / "proxy", work / "cache", work / "consumer"
    if cache.exists() or consumer.exists() or proxy.exists():
        raise ValueError("Frozen module consumption requires a fresh proxy, cache and consumer")
    # entries contains the actual ZIP and Go proxy metadata derived only from that ZIP.
    # entries 包含真实 ZIP 及仅派生自该 ZIP 的 Go 代理元数据。
    entries = proxy / escaped / "@v"
    entries.mkdir(parents=True)
    (entries / (version + ".zip")).write_bytes(archive.read_bytes())
    (entries / (version + ".mod")).write_bytes(sources["go.mod"])
    # The fixed bookkeeping time is not a claim about the release timestamp.
    # 固定记账时间不声称实际发布时间。
    (entries / (version + ".info")).write_bytes(module_info(version))
    (entries / "list").write_text(version + "\n", encoding="utf-8")
    consumer.mkdir()
    (consumer / "go.mod").write_text(f"module embedded-frozen-module-consumer\n\ngo 1.22\n\nrequire {MODULE_PATH} {version}\n", encoding="utf-8")
    environment.update(GOPROXY=proxy.as_uri(), GOMODCACHE=str(cache), GOSUMDB="off", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="")
    # download and directory prove import ownership before building any SDK or example.
    # download 和 directory 在构建任何 SDK 或示例前证明导入归属。
    download = json.loads(subprocess.check_output([go, "mod", "download", "-json", MODULE_PATH + "@" + version], cwd=consumer, env=environment))
    (work / "go-mod-download.json").write_text(json.dumps(download, indent=2) + "\n", encoding="utf-8")
    directory = Path(download["Dir"])
    if download["Path"] != MODULE_PATH or download["Version"] != version or not directory.resolve().is_relative_to(cache.resolve()):
        raise ValueError("Frozen module did not load from its exact isolated cache")
    if hashlib.sha256(Path(download["Zip"]).read_bytes()).hexdigest() != sha256:
        raise ValueError("Go cached module ZIP digest differs from frozen input")
    for name, contents in sources.items():
        if (directory / name).read_bytes() != contents:
            raise ValueError(f"Go downloaded module bytes changed: {name}")
    # imported is Go's actual package resolution, including the authoritative module version and absence of replacement.
    # imported 是 Go 实际包解析，包含权威模块版本及未使用替换的事实。
    imported = json.loads(subprocess.check_output([go, "list", "-json", MODULE_PATH], cwd=consumer, env=environment))
    (work / "go-list.json").write_text(json.dumps(imported, indent=2) + "\n", encoding="utf-8")
    if imported["ImportPath"] != MODULE_PATH or Path(imported["Dir"]).resolve() != directory.resolve() or imported["Module"]["Path"] != MODULE_PATH or imported["Module"]["Version"] != version or "Replace" in imported["Module"]:
        raise ValueError("Go package import does not use the exact frozen module ZIP")
    # evidence distinguishes real proxy/cache consumption from local source replacement.
    # evidence 区分真实代理／缓存消费与本地源码替换。
    evidence = {"mode": "frozen-module-zip", "module_zip": str(archive), "module_zip_sha256": sha256, "module_version": version, "module_members": len(sources), "module": str(directory), "gomodcache": str(cache), "module_sum": download["Sum"], "replace": False}
    (work / "module-members.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    return directory, consumer, sources, evidence
