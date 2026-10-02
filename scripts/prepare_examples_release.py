"""Prepare and inspect the actual published examples ZIP without development-only APIs.
准备并检查真实已发布示例 ZIP，排除仅开发版提供的 API。
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import posixpath
import re
import tempfile
from urllib.parse import unquote, urlsplit
import zipfile

from verify_embedded_distribution import module_sources


# This single published-example inventory expands only alongside a matching smoke-tested SDK module version.
# 此唯一已发布示例清单仅随匹配且已冒烟验证的 SDK 模块版本扩展。
PUBLISHED_EXAMPLES = ("basic", "query", "call", "lifecycle", "runtime_lease", "provider_callback")
# These shipped documents preserve every relative Markdown file link in the standalone package.
# 这些随包文档保留独立包中的每个相对 Markdown 文件链接。
PACKAGE_DOCUMENTS = ("LICENSE", "README.md", "README_cn.md", "examples/README.md", "examples/README_cn.md", "docs/embedded-validation.md")


def prepare(root: Path, output: Path, version: str, module_tag: str, npm_version: str) -> None:
    """Copy exact source to fresh output for matching Go version/module_tag and independently explicit npm_version; return nothing.
    将精确源码复制到匹配 Go version／module_tag 及独立显式 npm_version 的全新 output，无返回值。
    Existing output and inconsistent version are errors, preventing stale development examples from surviving.
    已存在输出及不一致版本均报错，防止陈旧开发示例残留。
    """
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?", version) or module_tag != "v" + version:
        raise ValueError("Package version and explicit module tag must match")
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", npm_version):
        raise ValueError("Explicit final NPM installer version is required")
    if output.exists():
        raise ValueError("Package output already exists; use a fresh directory")
    # Git-listed bytes exclude local installed fixture libraries and caches before runtime installation.
    # Git 所列字节在运行时安装前排除本地已安装夹具库及缓存。
    sources = module_sources(root)
    selected = set(PACKAGE_DOCUMENTS)
    for example in PUBLISHED_EXAMPLES:
        if f"examples/{example}/main.go" not in sources:
            raise ValueError(f"Missing published example: {example}")
    for name in sources:
        if any(name.startswith(f"examples/{example}/") for example in PUBLISHED_EXAMPLES) or name.startswith("examples/fixture-runtime/"):
            selected.add(name)
    for name in selected:
        if name not in sources:
            raise ValueError(f"Missing standalone package document: {name}")
    output.mkdir(parents=True)
    for name in sorted(selected):
        # member belongs to an already validated repository-relative source path.
        # member 属于已经校验的仓库相对源码路径。
        member = output / name
        member.parent.mkdir(parents=True, exist_ok=True)
        member.write_bytes(sources[name])
    (output / "go.mod").write_text(f"module luaskills-sdk-go-examples\n\ngo 1.22\n\nrequire github.com/LuaSkills/luaskills-sdk-go {module_tag}\n", encoding="utf-8")
    (output / "EXAMPLES_RELEASE.md").write_text(f"""# LuaSkills Go SDK Examples {version}

These six examples use `github.com/LuaSkills/luaskills-sdk-go {module_tag}`.
这六个示例使用 `github.com/LuaSkills/luaskills-sdk-go {module_tag}`。

See [English examples](examples/README.md), [中文示例](examples/README_cn.md), and [development validation boundaries](docs/embedded-validation.md).
The development-only `embedded_lifecycle` example is excluded; obtain it from a matching development source checkout.
仅开发版提供的 `embedded_lifecycle` 示例已排除；须从匹配的开发源码检出中取得。

```bash
go mod download
npx @luaskills/sdk@{npm_version} install-runtime --database none --runtime-root examples/fixture-runtime
CGO_ENABLED=1 go run ./examples/basic
CGO_ENABLED=1 go run ./examples/query
CGO_ENABLED=1 go run ./examples/runtime_lease
```
""", encoding="utf-8")


def verify_archive(archive: Path, package_name: str) -> dict[str, object]:
    """Inspect real archive under package_name, returning six-example/hash/link evidence or explicit failure.
    检查 package_name 下的真实 archive，返回六示例／摘要／链接证据或明确失败。
    Reject unpublished directories, duplicate/escaped paths, missing examples and broken relative document targets.
    拒绝未发布目录、重复／逃逸路径、缺失示例及文档相对目标断链。
    """
    # Runtime installation can add fixture assets and go.sum, but cannot expand the example source inventory.
    # 运行时安装可添加夹具资产及 go.sum，但不能扩展示例源码清单。
    root_files = set(PACKAGE_DOCUMENTS) | {"go.mod", "go.sum", "EXAMPLES_RELEASE.md"}
    with zipfile.ZipFile(archive) as package:
        # members and files preserve exact ZIP paths before any normalization.
        # members 和 files 在归一化前保留精确 ZIP 路径。
        members = package.namelist()
        files = {}
        if len(members) != len(set(members)):
            raise ValueError("Duplicate examples ZIP member")
        for member in members:
            # path is parsed without resolving against the host filesystem.
            # path 的解析不依赖宿主文件系统解析。
            path = PurePosixPath(member)
            if member == package_name + "/":
                continue
            if path.is_absolute() or ".." in path.parts or "\\" in member or not member.startswith(package_name + "/"):
                raise ValueError(f"Invalid examples ZIP member: {member}")
            # relative strips only the exact package prefix, never guesses artifact roots.
            # relative 仅移除精确包前缀，绝不猜测资产根。
            relative = member[len(package_name) + 1:]
            if member.endswith("/"):
                if relative.rstrip("/") not in ("examples", "docs", "examples/fixture-runtime", *(f"examples/{example}" for example in PUBLISHED_EXAMPLES)) and not relative.startswith("examples/fixture-runtime/") and not any(relative.startswith(f"examples/{example}/") for example in PUBLISHED_EXAMPLES):
                    raise ValueError(f"Unpublished examples ZIP directory: {relative}")
                continue
            if relative not in root_files and not relative.startswith("examples/fixture-runtime/") and not any(relative.startswith(f"examples/{example}/") for example in PUBLISHED_EXAMPLES):
                raise ValueError(f"Unpublished examples ZIP member: {relative}")
            files[relative] = package.read(member)
        for required in (*PACKAGE_DOCUMENTS, "go.mod", "EXAMPLES_RELEASE.md", *(f"examples/{example}/main.go" for example in PUBLISHED_EXAMPLES)):
            if required not in files:
                raise ValueError(f"Missing examples ZIP member: {required}")
        # links counts relative file targets in shipped host documents; runtime asset documents have separate ownership.
        # links 统计随包宿主文档的相对文件目标；运行时资产文档归独立所有者维护。
        links = 0
        for name in (*PACKAGE_DOCUMENTS, "EXAMPLES_RELEASE.md"):
            if not name.endswith(".md"):
                continue
            # prose excludes fenced code, where Go generic signatures are not Markdown links.
            # prose 排除围栏代码块，其中 Go 泛型签名不是 Markdown 链接。
            prose = re.sub(r"(?ms)^(`{3,}|~{3,})[^\n]*\n.*?^\1[ \t]*$", "", files[name].decode("utf-8").replace("\r\n", "\n"))
            for href in re.findall(r"\[[^\]]*\]\(([^)]+)\)", prose):
                # parsed distinguishes external links and anchors from relative file references.
                # parsed 区分外部链接及锚点与相对文件引用。
                parsed = urlsplit(href)
                if parsed.scheme or parsed.netloc or not parsed.path:
                    continue
                # target is the exact normalized package-relative Markdown file reference.
                # target 是精确归一化包相对 Markdown 文件引用。
                target = posixpath.normpath(posixpath.join(posixpath.dirname(name), unquote(parsed.path)))
                if target not in files:
                    raise ValueError(f"Broken standalone documentation link: {name} -> {href}")
                links += 1
    return {"archive": str(archive), "sha256": hashlib.sha256(archive.read_bytes()).hexdigest(), "examples": list(PUBLISHED_EXAMPLES), "development_examples": [], "members": len(files), "relative_document_links": links}


def write_archive(output: Path, archive: Path, package_name: str) -> None:
    """Write exact sorted package bytes with fixed timestamps/modes and no compressor dependence; return nothing.
    使用固定时间戳／模式及无压缩器依赖写入精确排序包字节；无返回值。
    """
    if archive.exists() or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*", package_name):
        raise ValueError("Examples archive requires a fresh path and explicit safe package root")
    archive.parent.mkdir(parents=True, exist_ok=True)
    # Stored members keep byte identity independent of filesystem mtime/mode and zlib implementation versions.
    # 存储成员使字节身份独立于文件系统 mtime／模式及 zlib 实现版本。
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_STORED) as package:
        for source in sorted(output.rglob("*")):
            if source.is_symlink():
                raise ValueError("Examples package cannot contain symbolic links")
            if source.is_file():
                member = zipfile.ZipInfo(package_name + "/" + source.relative_to(output).as_posix(), (1980, 1, 1, 0, 0, 0))
                member.create_system = 3
                member.external_attr = 0o100644 << 16
                member.compress_type = zipfile.ZIP_STORED
                package.writestr(member, source.read_bytes())


def check(root: Path, npm_version: str) -> None:
    """Create a real package/ZIP from root with explicit offline npm_version and retain packaging evidence; return nothing.
    从 root 及显式离线 npm_version 创建真实包／ZIP，并保留打包证据，无返回值。
    This does not download or verify published-tag compilation or runtime installation.
    此入口不下载内容，也不验证已发布标签编译或运行时安装。
    """
    # work is isolated; version and package_name derive from the unchanged authoritative VERSION file.
    # work 为隔离目录；version 和 package_name 派生自未修改的权威 VERSION 文件。
    (root / ".temp").mkdir(exist_ok=True)
    work = Path(tempfile.mkdtemp(prefix="examples-release-check-", dir=root / ".temp"))
    version = (root / "VERSION").read_text(encoding="utf-8").strip()
    package_name = "luaskills-sdk-go-examples-" + version
    # output and archive are fresh reviewable artifacts retained under the exact temporary directory.
    # output 和 archive 是保留于精确临时目录的全新可审核产物。
    output = work / package_name
    prepare(root, output, version, "v" + version, npm_version)
    archive = work / (package_name + ".zip")
    write_archive(output, archive, package_name)
    # report records actual ZIP evidence and the exact generated dependency tag.
    # report 记录真实 ZIP 证据及精确生成的依赖标签。
    report = verify_archive(archive, package_name)
    report["module_tag"] = "v" + version
    report["npm_version"] = npm_version
    report["formal_registry_consumption"] = False
    report["evidence"] = str(work)
    (work / "validation.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(report))


if __name__ == "__main__":
    # parser and mode expose the same implementation to workflow preparation, final ZIP verification and offline CI.
    # parser 和 mode 向工作流准备、最终 ZIP 校验及离线 CI 暴露同一实现。
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--output", type=Path)
    mode.add_argument("--verify-archive", type=Path)
    mode.add_argument("--check", action="store_true")
    parser.add_argument("--version")
    parser.add_argument("--module-tag")
    parser.add_argument("--npm-version", help="Independent exact installer version; offline --check does not authenticate publication")
    parser.add_argument("--package-name")
    # arguments retain explicit selected inputs; root selects only this script's repository.
    # arguments 保留显式所选输入；root 仅选择本脚本所在仓库。
    arguments = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    if arguments.check:
        if not arguments.npm_version:
            parser.error("--check requires explicit --npm-version")
        check(root, arguments.npm_version)
    elif arguments.output:
        if not arguments.version or not arguments.module_tag or not arguments.npm_version:
            parser.error("--output requires --version, --module-tag and --npm-version")
        prepare(root, arguments.output.resolve(), arguments.version, arguments.module_tag, arguments.npm_version)
    else:
        if not arguments.package_name:
            parser.error("--verify-archive requires --package-name")
        print(json.dumps(verify_archive(arguments.verify_archive, arguments.package_name)))
