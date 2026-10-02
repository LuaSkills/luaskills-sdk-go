"""Exercise publication guards against real altered example ZIPs rather than workflow source text.
针对真实已变更示例 ZIP 验证发布护栏，不检查工作流源码字符串。
"""
from pathlib import Path
import os
import tempfile
import unittest
import zipfile

from prepare_examples_release import prepare, verify_archive, write_archive


class PublishedArchiveTests(unittest.TestCase):
    """Reject actual development entries and missing linked documents in a prepared standalone archive.
    拒绝已准备独立归档中的实际开发条目及缺失链接文档。
    """

    def test_deterministic_actual_zip_across_metadata_changes(self) -> None:
        """Recreate actual package bytes after changed mtimes and require identical ZIP/sidecar and fixed member metadata.
        改变 mtime 后重建实际包字节，要求 ZIP／sidecar 相同及固定成员元数据。
        """
        root = Path(__file__).resolve().parents[1]
        version = (root / "VERSION").read_text(encoding="utf-8").strip()
        name = "luaskills-sdk-go-examples-" + version
        with tempfile.TemporaryDirectory() as temporary:
            work = Path(temporary)
            output = work / "prepared"
            prepare(root, output, version, "v" + version, "1.2.3")
            first, second = work / "first" / (name + ".zip"), work / "second" / (name + ".zip")
            write_archive(output, first, name)
            for source in output.rglob("*"):
                if source.is_file():
                    os.utime(source, (1800000000, 1800000000))
            write_archive(output, second, name)
            self.assertEqual(first.read_bytes(), second.read_bytes())
            first_report, second_report = verify_archive(first, name), verify_archive(second, name)
            self.assertEqual(first_report["sha256"], second_report["sha256"])
            self.assertEqual(first_report["sha256"] + "  " + first.name, second_report["sha256"] + "  " + second.name)
            with zipfile.ZipFile(first) as package:
                for member in package.infolist():
                    self.assertEqual(member.date_time, (1980, 1, 1, 0, 0, 0))
                    self.assertEqual(member.external_attr >> 16, 0o100644)
                    self.assertEqual(member.compress_type, zipfile.ZIP_STORED)
                    self.assertEqual(member.extra, b"")

    def test_actual_archive_boundaries(self) -> None:
        """Prepare repository example bytes, inspect a real ZIP, then prove three material archive regressions fail.
        准备仓库示例字节、检查真实 ZIP，并证明三个实质归档回归均失败。
        No download or native execution occurs; return nothing and retain no synthetic source changes.
        不下载或执行原生代码；无返回值，不保留合成源码变更。
        """
        # root/version select current repository facts; work/output are isolated actual package artifacts.
        # root／version 选择当前仓库事实；work／output 是隔离实际包产物。
        root = Path(__file__).resolve().parents[1]
        version = (root / "VERSION").read_text(encoding="utf-8").strip()
        with tempfile.TemporaryDirectory() as temporary:
            work = Path(temporary)
            output = work / "examples-package"
            # npm_version is deliberately independent from Go to detect installer documentation version drift.
            # npm_version 特意独立于 Go，以检测安装器文档版本漂移。
            npm_version = "1.2.3"
            prepare(root, output, version, "v" + version, npm_version)
            # files preserve every prepared package byte before controlled ZIP-only mutations.
            # files 在受控仅 ZIP 变更前保留每个已准备包字节。
            files = {source.relative_to(output).as_posix(): source.read_bytes() for source in output.rglob("*") if source.is_file()}
            for mutation in ("none", "development file", "development directory", "missing linked document"):
                with self.subTest(mutation):
                    # archive is a real file consumed by the same final workflow verifier.
                    # archive 是由同一个最终工作流校验器消费的真实文件。
                    archive = work / (mutation + ".zip")
                    with zipfile.ZipFile(archive, "w") as package:
                        for name, contents in files.items():
                            if mutation == "missing linked document" and name == "docs/embedded-validation.md":
                                continue
                            package.writestr("examples-package/" + name, contents)
                        if mutation == "development file":
                            package.writestr("examples-package/examples/embedded_lifecycle/main.go", "package main")
                        if mutation == "development directory":
                            package.writestr("examples-package/examples/embedded_lifecycle/", "")
                    if mutation == "none":
                        # report proves the prepared source excludes development API files and preserves relative file links.
                        # report 证明已准备源码排除开发 API 文件并保留相对文件链接。
                        report = verify_archive(archive, "examples-package")
                        self.assertEqual(len(report["examples"]), 6)
                        self.assertNotIn("examples/embedded_lifecycle/main.go", files)
                        self.assertGreater(report["relative_document_links"], 0)
                        with zipfile.ZipFile(archive) as packaged:
                            release_document = packaged.read("examples-package/EXAMPLES_RELEASE.md").decode("utf-8")
                        self.assertIn("npx @luaskills/sdk@" + npm_version + " install-runtime", release_document)
                        self.assertNotIn("npx @luaskills/sdk@" + version + " install-runtime", release_document)
                    else:
                        with self.assertRaises(ValueError):
                            verify_archive(archive, "examples-package")


if __name__ == "__main__":
    unittest.main()
