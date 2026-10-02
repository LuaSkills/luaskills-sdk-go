"""Freeze exact current Git-listed Go module source into a fresh ZIP and member manifest.
将 Git 所列当前精确 Go 模块源码冻结为全新 ZIP 及成员清单。
"""
from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
import stat
import zipfile

from embedded_module_archive import MODULE_PATH, module_info, validate_zip
from verify_embedded_distribution import module_sources


def freeze(root: Path, output: Path) -> dict[str, object]:
    """Freeze root's actual VERSION/source in fresh absolute output; return archive/hash/version/member evidence.
    将 root 的实际 VERSION／源码冻结到全新绝对 output；返回归档／摘要／版本／成员证据。
    Existing output, missing members or invalid module contents fail without changing versions or publishing.
    已存在输出、缺失成员或非法模块内容均失败，不修改版本或发布。
    """
    if not output.is_absolute() or output.exists():
        raise ValueError("Freeze output must be a fresh absolute directory")
    # sources/version are current source facts, never inferred from a published tag or adjacent core checkout.
    # sources／version 是当前源码事实，绝不根据已发布标签或相邻核心检出推断。
    sources = module_sources(root)
    version = "v" + sources["VERSION"].decode("utf-8").strip()
    output.mkdir(parents=True)
    # archive has deterministic entry metadata; the complete actual file bytes determine its identity.
    # archive 使用确定性条目元数据；完整实际文件字节决定其身份。
    archive = output / "module.zip"
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as package:
        for name, contents in sorted(sources.items()):
            # member stores exact module/version paths and regular-file ownership without symlink targets.
            # member 保存精确模块／版本路径及普通文件归属，不包含符号链接目标。
            member = zipfile.ZipInfo(MODULE_PATH + "@" + version + "/" + name)
            member.compress_type = zipfile.ZIP_DEFLATED
            member.external_attr = (stat.S_IFREG | 0o644) << 16
            package.writestr(member, contents)
    # sha256 and manifest are derived from the finished ZIP, not estimated from its input list.
    # sha256 和 manifest 派生自已完成 ZIP，不根据输入清单估算。
    sha256 = hashlib.sha256(archive.read_bytes()).hexdigest()
    frozen_sources, manifest = validate_zip(archive, version, sha256)
    if frozen_sources != sources:
        raise ValueError("Frozen module ZIP bytes differ from selected source")
    (output / "module.mod").write_bytes(frozen_sources["go.mod"])
    (output / "module.info").write_bytes(module_info(version))
    (output / "module-members.json").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    # report contains only actual frozen module evidence; its version is not a published-library compatibility claim.
    # report 仅包含实际冻结模块证据；其版本不声称已发布库兼容性。
    report = {"mode": "frozen-module-zip", "module_zip": str(archive), "module_zip_sha256": sha256, "module_version": version, "module_members": len(sources), "member_manifest": str(output / "module-members.json")}
    (output / "freeze.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    return report


if __name__ == "__main__":
    # parser/arguments expose only a deliberate fresh destination; root is this script's current repository.
    # parser／arguments 仅暴露明确全新目标；root 是本脚本所在当前仓库。
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path)
    arguments = parser.parse_args()
    root = Path(__file__).resolve().parents[1]
    print(json.dumps(freeze(root, arguments.output)))
