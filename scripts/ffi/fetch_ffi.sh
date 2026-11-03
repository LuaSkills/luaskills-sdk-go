#!/usr/bin/env bash
set -euo pipefail

# ProjectRoot points at the repository root regardless of the caller location.
# ProjectRoot 指向仓库根目录，避免调用方当前位置影响路径解析。
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# RuntimeRoot receives installed FFI headers, libraries, and license files.
# RuntimeRoot 接收已安装的 FFI 头文件、动态库与授权文件。
RUNTIME_ROOT="${RUNTIME_ROOT:-output}"

# LuaSkillsRepo stores the GitHub repository for LuaSkills FFI SDK assets.
# LuaSkillsRepo 保存 LuaSkills FFI SDK 资产所在的 GitHub 仓库。
LUASKILLS_REPO="${LUASKILLS_REPO:-LuaSkills/luaskills}"

# LuaSkillsVersion stores an optional exact vX.Y.Z core tag; empty resolves the newest published patch of the contract series.
# LuaSkillsVersion 保存可选的精确 vX.Y.Z 核心标签；为空时解析契约系列最新已发布补丁版本。
LUASKILLS_VERSION="${LUASKILLS_VERSION:-}"

# ContractDir holds the packaged embedded contract that fixes the core series and the required contract digest.
# ContractDir 保存包内嵌入式契约，它确定核心系列及所需契约摘要。
CONTRACT_DIR="$PROJECT_ROOT/contracts/embedded/v1"

ensure_dir() {
  # Create one directory when it does not exist.
  # 在目录不存在时创建该目录。
  mkdir -p "$1"
}

platform_key() {
  # Resolve the current luaskills FFI SDK asset platform key.
  # 解析当前 luaskills FFI SDK 资产平台标识。
  case "$(uname -s)" in
    Linux) os_key="linux" ;;
    Darwin) os_key="macos" ;;
    *) echo "Unsupported operating system: $(uname -s)" >&2; return 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch_key="x64" ;;
    aarch64|arm64) arch_key="arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; return 1 ;;
  esac
  printf '%s-%s\n' "$os_key" "$arch_key"
}

release_asset_info() {
  # Find one exact GitHub Release asset download URL and API digest.
  # 查找一个精确 GitHub Release 资产下载地址与 API 摘要。
  local repo="$1"
  local tag="$2"
  local asset_name="$3"
  local api_url="https://api.github.com/repos/${repo}/releases/tags/${tag}"
  curl -fsSL "$api_url" | python3 -c '
import json
import sys
asset_name = sys.argv[1]
data = json.load(sys.stdin)
for asset in data.get("assets", []):
    if asset.get("name") == asset_name:
        print(asset.get("browser_download_url", ""))
        print(asset.get("digest", ""))
        raise SystemExit(0)
print(f"asset not found: {asset_name}", file=sys.stderr)
raise SystemExit(2)
' "$asset_name"
}

save_release_asset_with_digest() {
  # Download one GitHub Release asset and verify its GitHub API digest.
  # 下载单个 GitHub Release 资产并校验其 GitHub API 摘要。
  local repo="$1"
  local tag="$2"
  local asset_name="$3"
  local destination="$4"
  # Lookup payload and status kept separately so only an exact missing-asset result may use compatibility fallback.
  # 分开保存查询载荷与状态，确保只有明确的资产缺失结果可以使用兼容回退。
  local asset_url asset_digest expected actual asset_info_text lookup_status
  if asset_info_text="$(release_asset_info "$repo" "$tag" "$asset_name")"; then
    lookup_status=0
  else
    lookup_status=$?
  fi
  if [ "$lookup_status" -ne 0 ]; then
    if [ "$lookup_status" -eq 2 ]; then
      return 2
    fi
    return 1
  fi
  # Plain reads keep this working on the Bash 3.2 that ships with macOS, which has no mapfile.
  # 使用普通 read，使脚本在 macOS 自带、没有 mapfile 的 Bash 3.2 上同样可用。
  asset_url=""
  asset_digest=""
  { IFS= read -r asset_url || true; IFS= read -r asset_digest || true; } <<< "$asset_info_text"
  if [[ "$asset_digest" != sha256:* ]]; then
    echo "GitHub API digest for $asset_name is missing or unsupported: $asset_digest" >&2
    return 1
  fi
  expected="${asset_digest#sha256:}"
  curl -fSL "$asset_url" -o "$destination"
  actual="$(python3 - "$destination" <<'PY'
import hashlib
import sys
digest = hashlib.sha256()
with open(sys.argv[1], "rb") as handle:
    for chunk in iter(lambda: handle.read(1024 * 1024), b""):
        digest.update(chunk)
print(digest.hexdigest())
PY
)"
  if [ "$expected" != "$actual" ]; then
    echo "SHA-256 mismatch for $asset_name. Expected $expected, got $actual" >&2
    return 1
  fi
}

contract_core_series() {
  # Read the core release series (major.minor) declared by the packaged embedded contract.
  # 读取包内嵌入式契约声明的核心发布系列（major.minor）。
  python3 - "$CONTRACT_DIR/contract.json" <<'PY'
import json
import re
import sys
series = json.load(open(sys.argv[1], encoding="utf-8")).get("core_series", "")
if not isinstance(series, str) or not re.fullmatch(r"(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)", series):
    raise SystemExit(f"embedded contract {sys.argv[1]} has no valid core_series")
print(series)
PY
}

contract_digest() {
  # Print the first token of one contract.sha256 sidecar.
  # 输出单个 contract.sha256 旁路文件的第一个字段。
  local sidecar="$1"
  if [ ! -f "$sidecar" ]; then
    echo "Embedded contract digest not found: $sidecar" >&2
    return 1
  fi
  awk 'NR == 1 { print $1 }' "$sidecar"
}

resolve_release_tag_for_series() {
  # Resolve the newest published GitHub release tag inside one major.minor series.
  # 解析一个 major.minor 协议线内最新的已发布 GitHub release 标签。
  local repo="$1"
  local series="$2"
  curl -fsSL "https://api.github.com/repos/${repo}/releases?per_page=100" | python3 -c '
import json
import re
import sys

series = sys.argv[1]
repo = sys.argv[2]
if not re.fullmatch(r"\d+\.\d+", series):
    raise SystemExit(f"unsupported core series: {series}")

matches = []
for release in json.load(sys.stdin):
    if release.get("draft") or release.get("prerelease"):
        continue
    tag = str(release.get("tag_name", ""))
    normalized = tag[1:] if tag.startswith("v") else tag
    if not re.fullmatch(r"\d+\.\d+\.\d+", normalized):
        continue
    major, minor, patch = (int(part) for part in normalized.split("."))
    if f"{major}.{minor}" != series:
        continue
    matches.append(((major, minor, patch), tag))

if not matches:
    raise SystemExit(f"no published release found for {repo} series {series}")

matches.sort(key=lambda item: item[0], reverse=True)
print(matches[0][1])
' "$series" "$repo"
}

resolve_luaskills_version() {
  # Resolve the core tag: an explicit tag must belong to the contract series; empty selects its newest published patch.
  # 解析核心标签：显式标签必须属于契约系列；为空时选择该系列最新已发布补丁版本。
  local explicit_tag="$1"
  local series
  series="$(contract_core_series)"
  if [ -z "$explicit_tag" ]; then
    resolve_release_tag_for_series "$LUASKILLS_REPO" "$series"
    return
  fi
  if ! [[ "$explicit_tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || [ "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}" != "$series" ]; then
    echo "LuaSkills core tag '$explicit_tag' is not an exact vX.Y.Z tag of core series $series supported by this SDK." >&2
    return 1
  fi
  printf '%s\n' "$explicit_tag"
}

verify_archive_contract() {
  # Require the extracted FFI SDK archive to package the same embedded contract digest as this SDK.
  # 要求已解压的 FFI SDK 归档打包与本 SDK 相同的嵌入式契约摘要。
  local extract_dir="$1"
  local asset_name="$2"
  local expected actual
  expected="$(contract_digest "$CONTRACT_DIR/contract.sha256")"
  actual="$(contract_digest "$extract_dir/contracts/embedded/v1/contract.sha256")" || {
    echo "Downloaded $asset_name does not contain contracts/embedded/v1/contract.sha256" >&2
    return 1
  }
  if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
    echo "Embedded contract mismatch in $asset_name ($LUASKILLS_REPO@$LUASKILLS_VERSION): expected $expected, got $actual" >&2
    return 1
  fi
}

luaskills_library_candidates() {
  # Return candidate LuaSkills dynamic library names for the current platform.
  # 返回当前平台对应的 LuaSkills 动态库候选名称。
  local platform
  platform="$(platform_key)"
  case "$platform" in
    linux-x64|linux-arm64) printf '%s\n' "libluaskills.so" "luaskills.so" ;;
    macos-x64|macos-arm64) printf '%s\n' "libluaskills.dylib" "luaskills.dylib" ;;
    *) echo "Unsupported LuaSkills FFI platform: $platform" >&2; return 1 ;;
  esac
}

has_existing_luaskills_ffi_content() {
  # Check whether the runtime root already contains one LuaSkills core dynamic library.
  # 检查运行根目录是否已经包含一个 LuaSkills core 动态库。
  local candidate=""
  while IFS= read -r candidate; do
    [ -n "$candidate" ] || continue
    if [ -f "$RUNTIME_ROOT/libs/$candidate" ]; then
      return 0
    fi
  done < <(luaskills_library_candidates)
  return 1
}

install_luaskills_ffi() {
  # Download and install one luaskills FFI SDK archive into the runtime root.
  # 下载并安装一个 luaskills FFI SDK 归档到运行根目录。
  local platform
  platform="$(platform_key)"
  local asset_name="luaskills-ffi-sdk-${platform}.tar.gz"
  local temp_dir
  temp_dir="$(mktemp -d)"
  trap 'rm -rf "$temp_dir"' RETURN
  local archive="$temp_dir/$asset_name"
  local extract_dir="$temp_dir/extract"
  ensure_dir "$extract_dir"
  # Classified download result used to separate exact asset absence from integrity or transport failures.
  # 用于区分明确资产缺失与完整性或传输失败的分类下载结果。
  local download_status=0
  save_release_asset_with_digest "$LUASKILLS_REPO" "$LUASKILLS_VERSION" "$asset_name" "$archive" || download_status=$?
  if [ "$download_status" -ne 0 ]; then
    if [ "$download_status" -eq 2 ] && has_existing_luaskills_ffi_content; then
      echo "WARNING: LuaSkills FFI SDK asset '$asset_name' was not found in $LUASKILLS_REPO@$LUASKILLS_VERSION. Existing packaged LuaSkills core content will be used." >&2
      return 0
    fi
    return 1
  fi
  tar -xzf "$archive" -C "$extract_dir"
  # Exact extracted core library required before any SDK content is copied into an existing runtime root.
  # 在任何 SDK 内容复制到现有运行根之前必须存在的精确解压核心库。
  local extracted_library="" candidate=""
  while IFS= read -r candidate; do
    if [ -f "$extract_dir/lib/$candidate" ]; then
      extracted_library="$extract_dir/lib/$candidate"
      break
    fi
  done < <(luaskills_library_candidates)
  if [ -z "$extracted_library" ]; then
    echo "LuaSkills dynamic library was not found in downloaded $asset_name" >&2
    return 1
  fi
  # The packaged contract digest is checked before anything is copied; a mismatch never falls back to another tag.
  # 在复制任何内容前校验包内契约摘要；不匹配时绝不回退到其他标签。
  verify_archive_contract "$extract_dir" "$asset_name"
  if [ -d "$extract_dir/include" ]; then
    ensure_dir "$RUNTIME_ROOT/include"
    cp -a "$extract_dir/include"/. "$RUNTIME_ROOT/include"/
  fi
  if [ -d "$extract_dir/lib" ]; then
    ensure_dir "$RUNTIME_ROOT/libs"
    cp -a "$extract_dir/lib"/. "$RUNTIME_ROOT/libs"/
  fi
  if [ -d "$extract_dir/licenses" ]; then
    ensure_dir "$RUNTIME_ROOT/licenses/luaskills-ffi"
    cp -a "$extract_dir/licenses"/. "$RUNTIME_ROOT/licenses/luaskills-ffi"/
  fi
}

cd "$PROJECT_ROOT"
LUASKILLS_VERSION="$(resolve_luaskills_version "$LUASKILLS_VERSION")"
ensure_dir "$RUNTIME_ROOT"
install_luaskills_ffi
echo "LuaSkills FFI SDK installed into $RUNTIME_ROOT"
