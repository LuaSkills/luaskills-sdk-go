"""Compile actual CGO-generated C declarations without linking or loading a native runtime.
编译真实 CGO 生成的 C 声明，不链接或加载原生运行时。
"""

import json
from pathlib import Path
import subprocess
import tempfile


def main():
    """Resolve this module's active CGO files and syntax-check every generated C unit; return nothing.
    解析本模块启用的 CGO 文件并检查所有生成 C 单元的语法；无返回值。
    Requires CGO_ENABLED=1 and Go's configured CC executable on PATH; any compiler failure propagates.
    要求 CGO_ENABLED=1 且 Go 配置的 CC 可执行文件位于 PATH；所有编译错误直接传播。
    """
    # The module and compiler are queried from Go, so new exported callbacks automatically join this gate.
    # 模块与编译器由 Go 查询，确保新增导出回调自动进入此门禁。
    root = Path(__file__).resolve().parents[1]
    module = json.loads(subprocess.check_output(['go', 'list', '-json', '.'], cwd=root, text=True))
    compiler = subprocess.check_output(['go', 'env', 'CC'], cwd=root, text=True).strip()
    if not module['CgoFiles']:
        raise RuntimeError('CGO declaration verification requires active CGO sources')
    # Generate into an owned temporary directory; do not alter tracked sources or require the core library.
    # 生成到独占临时目录；不修改受跟踪源码，也不要求核心动态库。
    with tempfile.TemporaryDirectory(prefix='luaskills-cgo-declarations-') as directory:
        subprocess.run(['go', 'tool', 'cgo', '-objdir', directory, '--', '-I', str(root),
                        *module['CgoFiles']], cwd=root, check=True)
        # CGO concatenates exported preambles here; checking all units also catches ordinary declaration drift.
        # CGO 在此合并导出前导声明；检查所有单元也能发现普通声明偏移。
        generated = sorted(Path(directory).glob('*.c'))
        if not generated:
            raise RuntimeError('CGO produced no C translation units')
        for source in generated:
            subprocess.run([compiler, '-fsyntax-only', '-I', str(root), '-I', directory, str(source)],
                           cwd=root, check=True)
        print(f'Checked {len(generated)} generated C units from {len(module["CgoFiles"])} CGO sources')


if __name__ == '__main__':
    main()
