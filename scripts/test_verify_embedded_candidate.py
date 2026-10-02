"""Prove invalid candidate inputs fail before any compiler or native execution.
证明非法候选输入在编译器或原生执行前失败。
"""
import argparse
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

from verify_embedded_candidate import digest, verify


class CandidateInputTests(unittest.TestCase):
    """Validate early local gate failures without loading synthetic library bytes.
    校验本地门禁的提前失败，不加载合成库字节。
    """

    def test_invalid_inputs_never_invoke_go(self) -> None:
        """Reject missing libraries, incorrect hashes and malformed description JSON before Go subprocesses.
        在 Go 子进程前拒绝缺失库、错误摘要及畸形描述 JSON。
        """
        # Files are synthetic input fixtures only; no compiler may consume them.
        # 文件仅是合成输入夹具；任何编译器都不得消费它们。
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            library = root / {"win32": "luaskills.dll", "linux": "libluaskills.so", "darwin": "libluaskills.dylib"}[sys.platform]
            library.write_bytes(b"invalid native fixture, never load")
            description = root / "description.json"
            description.write_text("{}", encoding="utf-8")
            arguments = argparse.Namespace(library=str(library), library_sha256=digest(library), description=str(description), go="must-not-run", race=False, timeout="1m", module_zip=None, module_zip_sha256=None, module_version=None, source_development=True)
            with patch("verify_embedded_candidate.subprocess.run") as compiler:
                with self.subTest("missing library"):
                    arguments.library = str(root / "missing")
                    with self.assertRaises(FileNotFoundError):
                        verify(arguments)
                arguments.library = str(library)
                with self.subTest("incorrect binary hash"):
                    arguments.library_sha256 = "0" * 64
                    with self.assertRaisesRegex(ValueError, "SHA-256 mismatch"):
                        verify(arguments)
                arguments.library_sha256 = digest(library)
                with self.subTest("missing description"):
                    arguments.description = str(root / "missing.json")
                    with self.assertRaises(FileNotFoundError):
                        verify(arguments)
                arguments.description = str(description)
                with self.subTest("malformed description"):
                    description.write_text("{", encoding="utf-8")
                    with self.assertRaises(ValueError):
                        verify(arguments)
                with self.subTest("non-object description"):
                    description.write_text("[]", encoding="utf-8")
                    with self.assertRaisesRegex(ValueError, "CoreDescription JSON object"):
                        verify(arguments)
                description.write_text("{}", encoding="utf-8")
                arguments.source_development = False
                with self.subTest("no mode fallback"):
                    with self.assertRaisesRegex(ValueError, "explicit --source-development"):
                        verify(arguments)
                arguments.module_zip = str(root / "module.zip")
                with self.subTest("partial frozen ZIP inputs"):
                    with self.assertRaisesRegex(ValueError, "all frozen module ZIP inputs"):
                        verify(arguments)
                compiler.assert_not_called()


if __name__ == "__main__":
    unittest.main()
