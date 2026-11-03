#!/usr/bin/env bash
set -euo pipefail

# ScriptRoot identifies the shared dependency script directory.
# ScriptRoot 标识共享依赖脚本目录。
SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# RepositoryRoot identifies the SDK or LuaSkills repository root.
# RepositoryRoot 标识 SDK 或 LuaSkills 仓库根目录。
REPOSITORY_ROOT="$(cd "$SCRIPT_ROOT/../.." && pwd)"
# Target selects the asset group synchronized into the runtime root.
# Target 选择要同步到运行根目录的资产分组。
TARGET="${1:-all}"
# RuntimeRoot receives all synchronized assets.
# RuntimeRoot 接收全部同步资产。
RUNTIME_ROOT="${RUNTIME_ROOT:-output}"
# LuaSkillsVersion stores an optional exact vX.Y.Z core tag; empty resolves the newest published patch of the contract series.
# LuaSkillsVersion 保存可选的精确 vX.Y.Z 核心标签；为空时解析契约系列最新已发布补丁版本。
LUASKILLS_VERSION="${LUASKILLS_VERSION:-}"

case "$TARGET" in
  all|luaskills)
    RUNTIME_ROOT="$RUNTIME_ROOT" LUASKILLS_VERSION="$LUASKILLS_VERSION" "$REPOSITORY_ROOT/scripts/ffi/fetch_ffi.sh"
    ;;
esac

case "$TARGET" in
  all|lua)
    RUNTIME_ROOT="$RUNTIME_ROOT" "$SCRIPT_ROOT/fetch_deps.sh" lua
    ;;
esac

case "$TARGET" in
  all|luaskills|lua) ;;
  *) echo "Usage: $0 [all|luaskills|lua]" >&2; exit 2 ;;
esac

echo "Runtime asset target '$TARGET' synchronized into $RUNTIME_ROOT"
