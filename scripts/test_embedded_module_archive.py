"""Verify actual frozen module ZIP integrity and pre-execution rejection without downloads.
验证真实冻结模块 ZIP 完整性及执行前拒绝，不下载任何内容。
"""
import hashlib
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

from embedded_module_archive import MODULE_PATH, consume_zip, validate_zip
from freeze_embedded_module import freeze


class FrozenModuleArchiveTests(unittest.TestCase):
    """Test real ZIP byte tampering, member escapes, missing files and exact module identity mismatches.
    测试真实 ZIP 字节篡改、成员逃逸、缺失文件及精确模块身份不匹配。
    """

    def test_frozen_zip_boundaries(self) -> None:
        """Freeze current source, validate its member manifest, then reject six altered inputs before Go execution.
        冻结当前源码、验证成员清单，并在 Go 执行前拒绝六种已变更输入。
        Return nothing; all alterations are isolated ZIP fixtures, never repository source edits.
        无返回值；所有变更均为隔离 ZIP 夹具，绝不编辑仓库源码。
        """
        # root selects current source facts; work contains only fresh local freeze/negative artifacts.
        # root 选择当前源码事实；work 仅包含全新本地冻结／负例产物。
        root = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as temporary:
            work = Path(temporary)
            frozen = freeze(root, work / "frozen")
            archive = Path(frozen["module_zip"])
            version = frozen["module_version"]
            # sources/manifest retain the exact validated finished ZIP contents.
            # sources／manifest 保留已完成 ZIP 的精确已校验内容。
            sources, manifest = validate_zip(archive, version, frozen["module_zip_sha256"])
            self.assertEqual(version, "v" + (root / "VERSION").read_text(encoding="utf-8").strip())
            self.assertEqual(manifest["members"]["go.mod"]["sha256"], hashlib.sha256(sources["go.mod"]).hexdigest())
            self.assertEqual(len(manifest["members"]), frozen["module_members"])
            for mutation in ("zip bytes", "wrong hash", "wrong version", "member escape", "missing driver", "wrong embedded VERSION"):
                with self.subTest(mutation):
                    # modified uses actual ZIP serialization; sha/version inputs remain explicit for each failure.
                    # modified 使用真实 ZIP 序列化；每个失败的 sha／version 输入均保持显式。
                    modified = work / (mutation + ".zip")
                    sha, selected_version = frozen["module_zip_sha256"], version
                    if mutation in ("zip bytes", "wrong hash", "wrong version"):
                        modified.write_bytes(archive.read_bytes())
                        if mutation == "zip bytes":
                            modified.write_bytes(modified.read_bytes() + b"tampered")
                        elif mutation == "wrong hash":
                            sha = "0" * 64
                        else:
                            selected_version = "v999.0.0"
                    else:
                        with zipfile.ZipFile(modified, "w") as package:
                            for name, contents in sources.items():
                                if mutation == "missing driver" and name == "embedded_driver.go":
                                    continue
                                if mutation == "wrong embedded VERSION" and name == "VERSION":
                                    contents = b"999.0.0\n"
                                package.writestr(MODULE_PATH + "@" + version + "/" + name, contents)
                            if mutation == "member escape":
                                package.writestr(MODULE_PATH + "@" + version + "/../escape.go", b"package escaped")
                        sha = hashlib.sha256(modified.read_bytes()).hexdigest()
                    with patch("embedded_module_archive.subprocess.check_output") as go:
                        with self.assertRaises(ValueError):
                            consume_zip("must-not-run", modified, selected_version, sha, work / "consume", {})
                        go.assert_not_called()
            with self.assertRaisesRegex(ValueError, "fresh absolute"):
                freeze(root, work / "frozen")


if __name__ == "__main__":
    unittest.main()
