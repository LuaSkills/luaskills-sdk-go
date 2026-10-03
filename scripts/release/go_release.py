"""Freeze and verify the Go SDK release chain; publication requires explicit workflow authorization.
冻结并验证 Go SDK 发布链；发布需要显式工作流授权。
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shutil
import subprocess
import sys
import tarfile
import time
import urllib.error
import urllib.parse
import urllib.request

# Local package tools remain the sole module ZIP and native acceptance implementation.
# 本地包工具仍是模块 ZIP 与原生验收的唯一实现。
ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from embedded_module_archive import MODULE_PATH, validate_zip
from freeze_embedded_module import freeze as freeze_module

# Public repository identities are fixed independently of each SDK's exact version.
# 公共仓库身份固定，且各 SDK 的精确版本独立。
REPOSITORY = "LuaSkills/luaskills-sdk-go"
API = "https://api.github.com/repos/" + REPOSITORY
NPM_REPOSITORY = "LuaSkills/luaskills-sdk-typescript"
PYPI_REPOSITORY = "LuaSkills/luaskills-sdk-python"
SCHEMA = 1
# Formal recovery headers have independent candidate and completion identities.
# 正式恢复验收头分别保留候选与完成身份。
SDK_FORMAL_SCHEMA = 2
# Matrix job names derive once here and are passed to the workflow as data.
# 矩阵作业名称在此唯一派生，并作为数据传给工作流。
NATIVE_JOB_PREFIX = "native-"
# Every GitHub response is subject to the same exact-byte download ceiling.
# 每个 GitHub 响应均受同一个精确字节下载上限约束。
MAX_HTTP_BYTES = 1024 * 1024 * 1024


def require(condition, message):
    """Reject a false condition with its explicit message; return nothing on success.
    条件为假时以明确消息拒绝；成功时无返回值。
    """
    if not condition:
        raise ValueError(message)


def sha256(path):
    """Return SHA-256 of path's exact bytes.
    返回 path 精确字节的 SHA-256。
    """
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def pairs(entries):
    """Decode unique JSON entries into an object, rejecting duplicate keys.
    将唯一 JSON 条目解码为对象，拒绝重复键。
    """
    result = {}
    for key, value in entries:
        require(key not in result, "Duplicate release JSON key: " + key)
        result[key] = value
    return result


def read_json(path):
    """Read path's strict UTF-8 JSON evidence and return its object.
    读取 path 的严格 UTF-8 JSON 凭证并返回对象。
    """
    return json.loads(Path(path).read_text(encoding="utf-8"), object_pairs_hook=pairs)


def write_json(path, value):
    """Write value to a new evidence file at path, refusing replacement.
    将 value 写入 path 的新凭证文件，拒绝替换。
    """
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("x", encoding="utf-8") as output:
        output.write(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def run(command, cwd=None, environment=None):
    """Execute exact command with a finite deadline; return UTF-8 output or fail.
    在有限期限内执行精确 command；返回 UTF-8 输出或失败。
    """
    try:
        result = subprocess.run([str(value) for value in command], cwd=cwd, env=environment,
                                check=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=1800)
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
        # The exception owns the original merged bytes; expose them before propagating the unchanged failure.
        # 异常持有原始合并字节；传播不变的失败前展示原件。
        if error.output is not None:
            sys.stderr.buffer.write(error.output)
            sys.stderr.buffer.flush()
        raise
    return result.stdout.decode("utf-8")


def full_sha(value):
    """Validate value as an exact lowercase Git commit identity and return it.
    验证 value 为精确小写 Git 提交身份并返回原值。
    """
    require(isinstance(value, str) and re.fullmatch(r"[0-9a-f]{40}", value), "Full lowercase Git SHA required")
    return value


def core_authority(root, commit):
    """Load public core tools from root at exactly commit; reject changed authority files.
    从精确 commit 的 root 加载公共核心工具；拒绝变动的权威文件。
    """
    root = Path(root).resolve()
    full_sha(commit)
    require(run(["git", "rev-parse", "HEAD"], root).strip() == commit, "Core checkout differs from frozen commit")
    for name in ("candidate.py", "sdk_prerequisites.py"):
        relative = "scripts/release/" + name
        committed = subprocess.check_output(["git", "show", commit + ":" + relative], cwd=root)
        require((root / relative).read_bytes() == committed, "Core release authority has uncommitted changes")
    sys.path.insert(0, str(root / "scripts/release"))
    for name in ("candidate", "sdk_prerequisites"):
        sys.modules.pop(name, None)
    import sdk_prerequisites
    return sdk_prerequisites


def runner_for(os_name, architecture):
    """Derive a native runner label from core's OS/architecture tuple, never a copied platform table.
    从核心操作系统／架构元组派生原生 runner 标签，不复制平台表。
    """
    require(architecture in ("x86_64", "aarch64"), "Unsupported core architecture")
    if os_name == "linux":
        return "ubuntu-24.04" if architecture == "x86_64" else "ubuntu-24.04-arm"
    if os_name == "macos":
        return "macos-15-intel" if architecture == "x86_64" else "macos-15"
    require(os_name == "windows" and architecture == "x86_64", "Unsupported core runner OS")
    return "windows-2025"


def source_metadata(root, core_tag):
    """Read Go VERSION, generated version mirror and asset tag from root; return the final SDK version.
    从 root 读取 Go VERSION、生成版本镜像及资产标签；返回正式 SDK 版本。
    """
    version = (root / "VERSION").read_text(encoding="utf-8").strip()
    require(re.fullmatch(r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)", version), "Final Go VERSION required")
    embedded = re.findall(r'^const EmbeddedCoreVersion = "([^"]+)"$', (root / "embedded_contract_generated.go").read_text(encoding="utf-8"), re.M)
    require(embedded == [version], "Go VERSION and embedded VERSION mirror differ")
    asset = re.findall(r'^const DefaultLuaSkillsVersion = "([^"]+)"$', (root / "runtime_assets.go").read_text(encoding="utf-8"), re.M)
    require(asset == [core_tag], "Default core asset tag differs from explicit core prerequisite tag")
    return version


def host_platform(authority, requested):
    """Resolve host to exactly one core platform, or validate the explicit platform string.
    将 host 解析为唯一核心平台，或校验显式平台字符串。
    """
    if requested != "host":
        require(requested in authority.candidate.PLATFORMS, "Unknown core platform")
        return requested
    os_name = {"win32": "windows", "darwin": "macos", "linux": "linux"}[sys.platform]
    architecture = platform.machine().lower()
    matches = [name for name, spec in authority.candidate.PLATFORMS.items() if spec[1] == os_name
               and architecture in ({"x86_64", "amd64"} if spec[2] == "x86_64" else {"aarch64", "arm64"})]
    require(len(matches) == 1, "Native host must match exactly one authoritative core platform")
    return matches[0]


def complete_core(proof, plan, authority):
    """Validate complete core receipt against plan and the sole platform authority; return the receipt.
    对照 plan 及唯一平台权威校验完整核心凭证；返回凭证。
    """
    require(proof["schema_version"] == authority.candidate.MANIFEST_VERSION and proof["phase"] == "complete"
            and proof["complete"] is True, "A complete real core prerequisite gate is mandatory")
    require(proof["core_tag"] == plan["core_tag"] and proof["core_commit"] == plan["core_commit"]
            and proof["github"]["repository"] == authority.REPOSITORY, "Core prerequisite identity mismatch")
    require(set(proof["sdk_inputs"]) == set(authority.candidate.PLATFORMS), "Missing mandatory core platform")
    consumer = proof["registry"]["consumer"]
    require(consumer["source"] == authority.REGISTRY_SOURCE and len(consumer["commands"]) == 4
            and all(type(item["exit_code"]) is int and item["exit_code"] == 0 for item in consumer["commands"])
            and all(item["command"][0] == "cargo" for item in consumer["commands"][:3])
            and consumer["commands"][3]["command"][0].replace("\\", "/").split("/")[-1] == consumer["executable"].replace("\\", "/").split("/")[-1]
            and re.fullmatch(r"[0-9a-f]{64}", consumer["executable_sha256"]), "Actual Cargo consumer command proof is missing")
    require(all(consumer["result"][key] is True for key in ("runtime", "pool_reuse", "drained"))
            and type(consumer["result"]["capability_calls"]) is int and consumer["result"]["capability_calls"] == 2,
            "Actual Cargo runtime result is incomplete")
    return proof


class GitHub:
    """Use bounded GitHub API reads and explicitly authorized writes with the workflow's GITHUB_TOKEN.
    使用工作流 GITHUB_TOKEN 执行有界 GitHub API 读取及显式授权写入。
    """

    def request(self, path, body=None, binary=False, repository=REPOSITORY):
        """Request fixed repository path, optionally posting body; return bytes or strict JSON.
        请求固定仓库 path，可选提交 body；返回字节或严格 JSON。
        """
        require(repository in (REPOSITORY, NPM_REPOSITORY, PYPI_REPOSITORY), "Unexpected SDK repository")
        token = os.environ.get("GITHUB_TOKEN")
        require(token, "GITHUB_TOKEN is required; no alternate publisher token is supported")
        if body is not None:
            require(os.environ.get("GITHUB_ACTIONS") == "true", "Remote mutation is allowed only in the authorized release workflow")
        request = urllib.request.Request("https://api.github.com/repos/" + repository + ("/" + path if path else ""),
                                        data=None if body is None else json.dumps(body).encode(),
                                        headers={"Authorization": "Bearer " + token,
                                                 "Accept": "application/octet-stream" if binary else "application/vnd.github+json",
                                                 "Content-Type": "application/json", "X-GitHub-Api-Version": "2022-11-28"})
        # The API download redirect must not forward the runner token to storage hosts.
        # API 下载重定向不得将 runner token 转发给存储主机。
        class Redirect(urllib.request.HTTPRedirectHandler):
            """Strip credentials on cross-host HTTPS redirects.
            在跨主机 HTTPS 重定向时移除凭证。
            """

            def redirect_request(self, req, fp, code, msg, headers, newurl):
                """Return redirected request after checking HTTPS and removing cross-host authorization.
                检查 HTTPS 并移除跨主机授权后返回重定向请求。
                """
                require(urllib.parse.urlsplit(newurl).scheme == "https", "GitHub redirected outside HTTPS")
                redirected = super().redirect_request(req, fp, code, msg, headers, newurl)
                if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(newurl).netloc:
                    redirected.remove_header("Authorization")
                return redirected
        with urllib.request.build_opener(Redirect()).open(request, timeout=60) as response:
            content = response.read(MAX_HTTP_BYTES + 1)
            require(len(content) <= MAX_HTTP_BYTES, "GitHub response exceeds bounded download")
        return content if binary else json.loads(content, object_pairs_hook=pairs)


def resolve_tag(api, tag, commit):
    """Resolve exact public tag recursively to commit, returning authenticated object chain.
    将精确公共标签递归解析到 commit，返回已认证对象链。
    """
    reference = api.request("git/ref/tags/" + urllib.parse.quote(tag, safe=""))
    require(reference["ref"] == "refs/tags/" + tag, "Unexpected Go tag reference")
    obj, chain = reference["object"], []
    while True:
        sha = full_sha(obj["sha"])
        require(sha not in {entry["sha"] for entry in chain} and len(chain) < 32, "Invalid Go annotated tag chain")
        chain.append({"sha": sha, "type": obj["type"]})
        if obj["type"] == "commit":
            require(sha == commit and api.request("git/commits/" + sha)["sha"] == commit, "Go tag resolves to a different commit")
            return chain
        require(obj["type"] == "tag", "Go tag does not resolve to a commit")
        annotation = api.request("git/tags/" + sha)
        require(annotation["sha"] == sha, "Go annotation identity mismatch")
        obj = annotation["object"]


def optional_tag(api, tag, commit):
    """Return the authenticated chain or None only when the exact ref is absent; broken existing objects fail.
    返回已认证链；仅精确 ref 不存在时返回 None；已有对象损坏则失败。
    """
    try:
        api.request("git/ref/tags/" + urllib.parse.quote(tag, safe=""))
    except urllib.error.HTTPError as error:
        require(error.code == 404, "Exact Go tag lookup failed")
        return None
    return resolve_tag(api, tag, commit)


def trusted_workflow(api, commit):
    """Require target in public default-branch history with identical workflow subtree before token publication.
    token 发布前要求目标位于公共默认分支历史，且工作流子树相同。
    """
    repository = api.request("")
    require(repository["full_name"] == REPOSITORY, "Go publishing repository mismatch")
    branch = api.request("branches/" + urllib.parse.quote(repository["default_branch"], safe=""))["commit"]["sha"]
    comparison = api.request("compare/" + commit + "..." + branch)
    require(comparison["merge_base_commit"]["sha"] == commit, "Frozen Go commit is not in trusted default-branch history")
    trees = []
    for sha in (commit, branch):
        record = api.request("git/trees/" + sha + "?recursive=1")
        require(record["truncated"] is False, "Cannot authenticate truncated workflow tree")
        entries = {entry["path"]: entry["sha"] for entry in record["tree"] if entry["path"].startswith(".github/workflows/")}
        require(".github/workflows/sdk-release.yml" in entries and ".github/workflows/examples-release.yml" in entries,
                "Trusted release workflows must be present on default branch before dispatch")
        trees.append(entries)
    require(trees[0] == trees[1], "GITHUB_TOKEN cannot authorize a new or modified workflow subtree at the frozen tag")
    return {"default_branch": repository["default_branch"], "default_commit": branch, "workflows": trees[0]}


def bootstrap(args):
    """Authenticate fresh non-authorizing public core inputs and derive this host's exact installable Rust toolchain.
    认证新鲜、不授权发布的公共核心输入，并派生本宿主精确可安装 Rust 工具链。
    """
    authority = core_authority(args.core_root, args.core_commit)
    selected = host_platform(authority, args.platform)
    require(selected == host_platform(authority, "host"), "Rust bootstrap platform differs from native host")
    require(not args.output.exists(), "Toolchain bootstrap output must be fresh")
    args.output.mkdir(parents=True)
    proof = authority.run_gate(args.core_tag, args.core_commit, args.output / "core", "github-only")
    require(proof["phase"] == "github-only" and proof["complete"] is False and proof["registry"] is None
            and proof["core_tag"] == args.core_tag and proof["core_commit"] == args.core_commit,
            "Rust bootstrap must retain exact non-authorizing public core identity")
    toolchain = authority.toolchain_inputs(args.output / "core/prerequisites.json", selected)
    require(toolchain["phase"] == "github-only" and toolchain["complete"] is False
            and toolchain["core_tag"] == args.core_tag and toolchain["core_commit"] == args.core_commit
            and toolchain["platform"] == selected and re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", toolchain["toolchain"]),
            "Public authority returned invalid Rust toolchain inputs")
    write_json(args.output / "toolchain.json", toolchain)
    return toolchain


def freeze(args):
    """Freeze exact clean SDK commit, one module ZIP and core receipt into args.output.
    将精确干净 SDK 提交、一个模块 ZIP 及核心凭证冻结到 args.output。
    """
    root = Path(args.root).resolve()
    full_sha(args.sdk_sha)
    require(args.sdk_sha == args.workflow_sha == os.environ.get("GITHUB_SHA", args.workflow_sha)
            == os.environ.get("GITHUB_WORKFLOW_SHA", args.workflow_sha), "SDK/workflow/github SHA mismatch")
    require(run(["git", "rev-parse", "HEAD"], root).strip() == args.sdk_sha, "SDK checkout differs from frozen SHA")
    require(not run(["git", "status", "--porcelain", "--untracked-files=all"], root).strip(), "Formal freeze requires clean committed SDK source")
    version = source_metadata(root, args.core_tag)
    authority = core_authority(args.core_root, args.core_commit)
    plan = {"schema_version": SCHEMA, "sdk_sha": args.sdk_sha, "sdk_version": version,
            "module_version": "v" + version, "core_tag": args.core_tag, "core_commit": args.core_commit,
            "npm_version": args.npm_version, "npm_sha": full_sha(args.npm_sha),
            "pypi_version": args.pypi_version, "pypi_sha": full_sha(args.pypi_sha)}
    for kind in ("npm", "pypi"):
        for phase in ("candidate", "completion"):
            run_id, attempt = getattr(args, kind + "_" + phase + "_run_id"), getattr(args, kind + "_" + phase + "_run_attempt")
            require(re.fullmatch(r"[1-9][0-9]*", str(run_id)) and re.fullmatch(r"[1-9][0-9]*", str(attempt)), "Explicit positive SDK run and attempt required")
            plan[kind + "_" + phase + "_run_id"] = str(run_id)
            plan[kind + "_" + phase + "_run_attempt"] = int(attempt)
        plan[kind + "_completion_source_sha"] = full_sha(getattr(args, kind + "_completion_source_sha"))
        require(plan[kind + "_completion_source_sha"] == plan[kind + "_sha"], "Initial prerequisite completion source must equal original SDK source")
    require(all(re.fullmatch(r"\d+\.\d+\.\d+", plan[field]) for field in ("npm_version", "pypi_version")), "Explicit formal NPM/PyPI versions required")
    require(not args.output.exists(), "Release output must be fresh")
    args.output.mkdir(parents=True)
    require(args.mode in ("artifact-only", "publish"), "Candidate freeze mode must be artifact-only or publish")
    if args.mode == "publish":
        plan["trusted_workflow"] = trusted_workflow(GitHub(), args.sdk_sha)
    frozen = freeze_module(root, args.output / "module")
    plan.update(module_zip_sha256=frozen["module_zip_sha256"], module_members=frozen["module_members"])
    proof = authority.run_gate(args.core_tag, args.core_commit, args.output / "core", "complete")
    complete_core(proof, plan, authority)
    plan["core_version"] = proof["core_version"]
    plan["prerequisites_sha256"] = sha256(args.output / "core/prerequisites.json")
    plan["matrix"] = {"include": [{"platform": name, "runner": runner_for(spec[1], spec[2]), "job": NATIVE_JOB_PREFIX + name}
                                   for name, spec in authority.candidate.PLATFORMS.items()]}
    write_json(args.output / "plan.json", plan)
    if args.github_output:
        with args.github_output.open("a", encoding="utf-8") as output:
            output.write("matrix=" + json.dumps(plan["matrix"]) + "\n")
    return plan


def module_bytes(plan, directory):
    """Validate directory's one frozen ZIP and exact manifest against plan; return source bytes.
    对照 plan 校验 directory 中唯一冻结 ZIP 及精确清单；返回源码字节。
    """
    sources, manifest = validate_zip(Path(directory) / "module.zip", plan["module_version"], plan["module_zip_sha256"])
    require(len(sources) == plan["module_members"] and manifest == read_json(Path(directory) / "module-members.json"), "Frozen Go module member manifest mismatch")
    return sources


def parse_events(content, required):
    """Require real test2json passes without skips/failures and specified receipt tests; return event counts.
    要求真实 test2json 通过且无跳过／失败，以及指定凭证测试；返回事件计数。
    """
    counts, passes = {"pass": 0, "fail": 0, "skip": 0}, set()
    for line in content.decode("utf-8").splitlines():
        if not line.startswith("{"):
            continue
        event = json.loads(line, object_pairs_hook=pairs)
        if "Test" in event and event["Action"] in counts:
            counts[event["Action"]] += 1
            if event["Action"] == "pass":
                passes.add(event["Test"])
    require(counts["pass"] > 0 and counts["fail"] == counts["skip"] == 0 and set(required) <= passes,
            "Native test receipt missing, skipped or failed")
    return counts


def validate_native(summary, plan, inputs, files):
    """Validate actual candidate receipts and logs against exact module/core inputs; return accepted facts.
    对照精确模块／核心输入验证实际候选凭证与日志；返回验收事实。
    """
    require(summary["mode"] == "frozen-module-zip" and summary["replace"] is False and summary["race"] is True,
            "Formal native gate requires actual ZIP imports and cgo race")
    require(summary["module_zip_sha256"] == plan["module_zip_sha256"] and summary["module_version"] == plan["module_version"]
            and summary["module_members"] == plan["module_members"], "Native module identity mismatch")
    require(summary["sha256"] == inputs["library_sha256"] and summary["description_sha256"] == inputs["description_sha256"], "Native core bytes mismatch")
    require(Path(summary["module"]).is_relative_to(Path(summary["gomodcache"])), "Native module import is outside the empty cache")
    tests = parse_events(files["tests.jsonl"], ("TestEmbeddedCandidateEvidence", "TestEmbeddedCandidateDescriptionExtensions"))
    lifecycle = parse_events(files["example-tests.jsonl"], ("TestEmbeddedReserveCancelledObserver",))
    require(tests == summary["tests"] and lifecycle == summary["lifecycle_tests"], "Native receipt count differs from logs")
    require(b"Runtime scope, callback pump, driver receipts and transport closed." in files["example.log"], "Actual closed lifecycle example receipt missing")
    return {"tests": tests, "lifecycle_tests": lifecycle, "race": True, "replace": False,
            "import_directory": summary["module"], "gomodcache": summary["gomodcache"], "module_sum": summary["module_sum"]}


def native(args):
    """Run the existing native ZIP gate on this native runner and persist portable receipts.
    在此原生 runner 执行既有 ZIP 原生门禁，并保存可移植凭证。
    """
    plan = read_json(args.plan)
    authority = core_authority(args.core_root, plan["core_commit"])
    args.platform = host_platform(authority, args.platform)
    proof = complete_core(read_json(args.prerequisites), plan, authority)
    require(sha256(args.prerequisites) == plan["prerequisites_sha256"], "Original complete core evidence changed")
    require(args.platform in authority.candidate.PLATFORMS, "Unexpected platform")
    spec = authority.candidate.PLATFORMS[args.platform]
    host_os = {"win32": "windows", "darwin": "macos", "linux": "linux"}[sys.platform]
    host_arch = platform.machine().lower()
    require(host_os == spec[1] and host_arch in ({"x86_64", "amd64"} if spec[2] == "x86_64" else {"aarch64", "arm64"}), "Cross-compilation cannot count as native acceptance")
    inputs = authority.resolve_sdk_inputs(args.prerequisites, args.platform)
    module_bytes(plan, args.module)
    require(not args.output.exists(), "Native evidence output must be fresh")
    args.output.mkdir(parents=True)
    try:
        output = run([sys.executable, ROOT / "scripts/verify_embedded_candidate.py", "--module-zip", (args.module / "module.zip").resolve(),
                      "--module-zip-sha256", plan["module_zip_sha256"], "--module-version", plan["module_version"],
                      "--library", inputs["library"], "--library-sha256", inputs["library_sha256"],
                      "--description", inputs["description"], "--race"])
    except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
        # Error retains this credential-free child's merged raw bytes; preserve failure instead of accepting it.
        # Error 保留本无凭据子进程的合并原始字节；保留失败而非将其接受。
        if error.output is not None:
            (args.output / "native.log").write_bytes(error.output)
        raise
    summary = json.loads(output.splitlines()[-1], object_pairs_hook=pairs)
    work = Path(summary["evidence"])
    files = {name: (work / name).read_bytes() for name in ("tests.jsonl", "example-tests.jsonl", "example.log")}
    facts = validate_native(summary, plan, inputs, files)
    for name, content in files.items():
        (args.output / name).write_bytes(content)
    (args.output / "native.log").write_text(output, encoding="utf-8")
    report = {"schema_version": SCHEMA, "sdk_sha": plan["sdk_sha"], "sdk_version": plan["sdk_version"],
              "module_zip_sha256": plan["module_zip_sha256"], "module_version": plan["module_version"], "platform": args.platform,
              "core_commit": plan["core_commit"], "core_tag": plan["core_tag"], "accepted": True, "facts": facts,
              "inputs": {key: inputs[key] for key in ("library_sha256", "description_sha256", "archive_sha256", "build")},
              "logs": {name: hashlib.sha256(body).hexdigest() for name, body in files.items()}}
    write_json(args.output / "report.json", report)
    return report


def aggregate_records(plan, proof, records, authority):
    """Require every mandatory native platform and all immutable identities; return aggregate.
    要求全部必需原生平台及所有不可变身份；返回聚合凭证。
    """
    complete_core(proof, plan, authority)
    seen = set()
    for record in records:
        name = record["platform"]
        require(name in authority.candidate.PLATFORMS and name not in seen, "Unknown or duplicate Go native platform")
        seen.add(name)
        require(record["schema_version"] == SCHEMA and record["accepted"] is True, "Native platform not accepted")
        require(all(record[field] == plan[field] for field in ("sdk_sha", "sdk_version", "module_zip_sha256", "module_version", "core_commit", "core_tag")), "SDK native report identity mismatch")
        require(record["inputs"] == {key: proof["sdk_inputs"][name][key] for key in ("library_sha256", "description_sha256", "archive_sha256", "build")}, "Frozen platform core input mismatch")
        require(record["facts"]["race"] is True and record["facts"]["replace"] is False, "Native cgo race/ZIP receipt missing")
        for field in ("tests", "lifecycle_tests"):
            counts = record["facts"][field]
            require(counts["pass"] > 0 and counts["fail"] == counts["skip"] == 0, "Native tests failed or skipped")
    require(seen == set(authority.candidate.PLATFORMS), "Missing mandatory Go native platform")
    return {"schema_version": SCHEMA, "plan": plan, "platforms": sorted(records, key=lambda item: item["platform"]), "accepted": True}


def aggregate(args):
    """Read actual native logs and module bytes before aggregating to a new output.
    聚合到新 output 前读取实际原生日志及模块字节。
    """
    plan = read_json(args.plan)
    authority = core_authority(args.core_root, plan["core_commit"])
    require(sha256(args.prerequisites) == plan["prerequisites_sha256"], "Core prerequisite receipt changed")
    module_bytes(plan, args.module)
    records = []
    for path in sorted(args.reports.rglob("report.json")):
        record = read_json(path)
        for name, expected in record["logs"].items():
            require(name in ("tests.jsonl", "example-tests.jsonl", "example.log") and sha256(path.parent / name) == expected, "Native log checksum mismatch")
        require(set(record["logs"]) == {"tests.jsonl", "example-tests.jsonl", "example.log"}, "Missing native receipt logs")
        require(parse_events((path.parent / "tests.jsonl").read_bytes(), ("TestEmbeddedCandidateEvidence", "TestEmbeddedCandidateDescriptionExtensions")) == record["facts"]["tests"], "Root native log mismatch")
        require(parse_events((path.parent / "example-tests.jsonl").read_bytes(), ("TestEmbeddedReserveCancelledObserver",)) == record["facts"]["lifecycle_tests"], "Actual reservation cleanup receipt missing")
        require(b"Runtime scope, callback pump, driver receipts and transport closed." in (path.parent / "example.log").read_bytes(), "Actual closed receipt missing")
        records.append(record)
    result = aggregate_records(plan, read_json(args.prerequisites), records, authority)
    write_json(args.output, result)
    return result




def fixed_sdk_script(root, commit):
    """Return the committed SDK release helper at root's exact commit, rejecting dirty scripts.
    返回 root 精确提交中的 SDK 发布工具，拒绝变动脚本。
    """
    require(run(["git", "rev-parse", "HEAD"], root).strip() == full_sha(commit), "Prerequisite SDK checkout/source mismatch")
    require(not run(["git", "status", "--porcelain", "--untracked-files=all"], root).strip(), "Prerequisite SDK source must be clean and committed")
    relative = "scripts/release/sdk_release.py"
    script = root / relative
    require(script.read_bytes() == subprocess.check_output(["git", "show", commit + ":" + relative], cwd=root), "Prerequisite SDK helper differs from frozen source")
    return script


def validate_sdk_header(header, plan, kind, repository):
    """Bind another SDK's actual formal-consumer acceptance header to this independent release batch.
    将其他 SDK 的实际正式消费者验收头绑定到此独立发布批次。
    """
    require(header["schema_version"] == SDK_FORMAL_SCHEMA and header["accepted"] is True and header["repository"] == repository,
            "Nonformal SDK prerequisite receipt")
    require(header["sdk_source_sha"] == plan[kind + "_sha"] and header["sdk_version"] == plan[kind + "_version"]
            and header["completion_source_sha"] == plan[kind + "_completion_source_sha"] == plan[kind + "_sha"]
            and header["core_commit"] == plan["core_commit"] and header["core_tag"] == plan["core_tag"],
            "Formal prerequisite SDK source/version/core/batch mismatch")
    for phase in ("candidate", "completion"):
        require(type(header[phase + "_run_id"]) is str and header[phase + "_run_id"] == plan[kind + "_" + phase + "_run_id"]
                and type(header[phase + "_run_attempt"]) is int and header[phase + "_run_attempt"] == plan[kind + "_" + phase + "_run_attempt"],
                "Formal prerequisite SDK exact attempt mismatch")
    require(re.fullmatch(r"[0-9a-f]{64}", header["registry_consumer_sha256"]), "Actual registry consumer digest missing")
    require(isinstance(header["registry_consumer_file"], str)
            and re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*\.json", header["registry_consumer_file"]), "Registry consumer must be one explicit safe JSON basename")


def sdk_identity_flags(plan, kind):
    """Return the authoritative SDK formal-proof CLI's explicit dual run/attempt/source flags for kind.
    返回 kind 所属 SDK 正式证明 CLI 的显式双运行／轮次／源码参数。
    """
    flags = []
    for phase in ("candidate", "completion"):
        for suffix in ("run-id", "run-attempt"):
            flags.extend(("--" + phase + "-" + suffix, str(plan[kind + "_" + phase + "_" + suffix.replace("-", "_")])))
    return flags + ["--completion-source-sha", plan[kind + "_completion_source_sha"]]


def prerequisites(args):
    """Re-run complete core and independent formal NPM/PyPI consumers immediately before tagging.
    打标签前重新执行完整核心及独立正式 NPM/PyPI 消费者。
    """
    aggregate = read_json(args.aggregate)
    require(aggregate["accepted"] is True, "Go five-platform aggregate is mandatory")
    plan = aggregate["plan"]
    authority = core_authority(args.core_root, plan["core_commit"])
    args.platform = host_platform(authority, args.platform)
    module_bytes(plan, args.module)
    require(not args.output.exists(), "Publication recheck output must be fresh")
    args.output.mkdir(parents=True)
    proof = authority.recheck(args.prerequisites, args.output / "core")
    complete_core(proof, plan, authority)
    npm_script = fixed_sdk_script(args.npm_root, plan["npm_sha"])
    pypi_script = fixed_sdk_script(args.pypi_root, plan["pypi_sha"])
    environment = dict(os.environ, GH_TOKEN=os.environ["GITHUB_TOKEN"])
    run([sys.executable, npm_script, "formal-proof", "--core-root", args.core_root,
         *sdk_identity_flags(plan, "npm"), "--repository", NPM_REPOSITORY, "--sdk-source-sha", plan["npm_sha"],
         "--sdk-version", plan["npm_version"], "--output", args.output / "npm"], environment=environment)
    run([sys.executable, pypi_script, "formal-proof", "--core-root", args.core_root,
         *sdk_identity_flags(plan, "pypi"), "--source-sha", plan["pypi_sha"], "--sdk-version", plan["pypi_version"],
         "--core-tag", plan["core_tag"], "--core-commit", plan["core_commit"], "--platform", args.platform,
         "--output", args.output / "pypi"], environment=environment)
    # Each SDK owns durable publication authentication and its real cold registry consumer.
    # 每个 SDK 自行负责持久发布认证及其真实冷 registry 消费者。
    npm, pypi = read_json(args.output / "npm/accepted.json"), read_json(args.output / "pypi/accepted.json")
    validate_sdk_header(npm, plan, "npm", NPM_REPOSITORY)
    validate_sdk_header(pypi, plan, "pypi", PYPI_REPOSITORY)
    require(sha256(args.output / "npm" / npm["registry_consumer_file"]) == npm["registry_consumer_sha256"], "Actual NPM consumer receipt changed")
    require(sha256(args.output / "pypi" / pypi["registry_consumer_file"]) == pypi["registry_consumer_sha256"], "Actual PyPI consumer receipt changed")
    module_bytes(plan, args.module)
    result = {"schema_version": SCHEMA, "plan": plan, "aggregate_sha256": sha256(args.aggregate), "core": proof,
              "npm": npm, "pypi": pypi, "accepted": True}
    write_json(args.output / "publish-proof.json", result)
    return result


def publish_tag(args):
    """Create one immutable Go tag after all current prerequisites, or authenticate identical reuse.
    完成所有当前前置验收后创建一个不可变 Go 标签，或认证相同复用。
    """
    proof = read_json(args.proof)
    require(proof["schema_version"] == SCHEMA and proof["accepted"] is True, "Current complete publication proof required")
    plan = proof["plan"]
    require(os.environ.get("GITHUB_SHA") == plan["sdk_sha"] == os.environ.get("GITHUB_WORKFLOW_SHA")
            and os.environ.get("GITHUB_JOB") == "completion", "Go tag creation must use the exact trusted completion job")
    authority = core_authority(args.core_root, plan["core_commit"])
    complete_core(proof["core"], plan, authority)
    validate_sdk_header(proof["npm"], plan, "npm", NPM_REPOSITORY)
    validate_sdk_header(proof["pypi"], plan, "pypi", PYPI_REPOSITORY)
    module_bytes(plan, args.module)
    api = GitHub()
    trusted_workflow(api, plan["sdk_sha"])
    chain = optional_tag(api, plan["module_version"], plan["sdk_sha"])
    created = chain is None
    if created:
        api.request("git/refs", {"ref": "refs/tags/" + plan["module_version"], "sha": plan["sdk_sha"]})
        chain = resolve_tag(api, plan["module_version"], plan["sdk_sha"])
    result = {"schema_version": SCHEMA, "plan": plan, "created_this_run": created, "tag_resolution": chain,
              "publish_proof_sha256": sha256(args.proof)}
    write_json(args.output, result)
    return result


def public_module(args):
    """Cold-download exact public module and compare all member bytes before native reacceptance.
    冷下载精确公共模块，并在原生重新验收前比较所有成员字节。
    """
    plan = read_json(args.plan)
    frozen = module_bytes(plan, args.module)
    chain = resolve_tag(GitHub(), plan["module_version"], plan["sdk_sha"])
    require(not args.output.exists(), "Public module proof requires a fresh cache/output")
    consumer, cache = args.output / "consumer", args.output / "cache"
    consumer.mkdir(parents=True)
    (consumer / "go.mod").write_text(f"module public-go-sdk-consumer\n\ngo 1.22\n\nrequire {MODULE_PATH} {plan['module_version']}\n", encoding="utf-8")
    environment = dict(os.environ, GOPROXY="https://proxy.golang.org", GOSUMDB="sum.golang.org", GOMODCACHE=str(cache.resolve()),
                       GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="", GOPRIVATE="", GONOPROXY="", GONOSUMDB="")
    download = json.loads(run(["go", "mod", "download", "-json", MODULE_PATH + "@" + plan["module_version"]], consumer, environment), object_pairs_hook=pairs)
    require(download["Path"] == MODULE_PATH and download["Version"] == plan["module_version"] and "Error" not in download,
            "Public proxy returned another module/version")
    actual_archive = Path(download["Zip"])
    actual, manifest = validate_zip(actual_archive, plan["module_version"], sha256(actual_archive))
    require(actual == frozen, "Public Go module member content differs from frozen ZIP")
    directory = Path(download["Dir"])
    require(directory.resolve().is_relative_to(cache.resolve()), "Public module did not import from the empty cache")
    imported = json.loads(run(["go", "list", "-json", MODULE_PATH], consumer, environment), object_pairs_hook=pairs)
    require(imported["ImportPath"] == MODULE_PATH and Path(imported["Dir"]).resolve() == directory.resolve()
            and imported["Module"]["Version"] == plan["module_version"] and "Replace" not in imported["Module"], "Public import override detected")
    for name, body in frozen.items():
        require((directory / name).read_bytes() == body, "Public extracted module bytes changed: " + name)
    authority = core_authority(args.core_root, plan["core_commit"])
    args.platform = host_platform(authority, args.platform)
    complete_core(read_json(args.prerequisites), plan, authority)
    inputs = authority.resolve_sdk_inputs(args.prerequisites, args.platform)
    output = run([sys.executable, ROOT / "scripts/verify_embedded_candidate.py", "--module-zip", actual_archive.resolve(),
                  "--module-zip-sha256", sha256(actual_archive), "--module-version", plan["module_version"],
                  "--library", inputs["library"], "--library-sha256", inputs["library_sha256"], "--description", inputs["description"], "--race"])
    summary = json.loads(output.splitlines()[-1], object_pairs_hook=pairs)
    # Compression bytes may differ publicly; exact source members remain the publication authority.
    # 公共压缩字节允许不同；精确源码成员仍是发布权威。
    public_plan = {**plan, "module_zip_sha256": sha256(actual_archive)}
    work = Path(summary["evidence"])
    files = {name: (work / name).read_bytes() for name in ("tests.jsonl", "example-tests.jsonl", "example.log")}
    facts = validate_native(summary, public_plan, inputs, files)
    for name, body in files.items():
        (args.output / name).write_bytes(body)
    write_json(args.output / "go-mod-download.json", download)
    write_json(args.output / "go-list.json", imported)
    write_json(args.output / "module-members.json", manifest)
    write_json(args.output / "native-summary.json", summary)
    result = {"schema_version": SCHEMA, "plan": plan, "accepted": True, "public_proxy": environment["GOPROXY"],
              "tag_resolution": chain, "public_zip_sha256": sha256(actual_archive), "module_sum": download["Sum"],
              "public_import_directory": str(directory), "gomodcache": str(cache), "replace": False,
              "native": facts, "platform": args.platform, "library_sha256": inputs["library_sha256"]}
    write_json(args.output / "public-module.json", result)
    return result


def assets_to_upload(existing, files, download):
    """Compare all existing asset bytes before any upload; return only missing files.
    在任何上传前比较全部已有资产字节；仅返回缺失文件。
    """
    require(len({entry["name"] for entry in existing}) == len(existing), "Duplicate existing release assets")
    missing = []
    for path in files:
        selected = [entry for entry in existing if entry["name"] == path.name]
        if selected:
            require(download(selected[0]) == path.read_bytes(), "Existing release asset differs: " + path.name)
        else:
            missing.append(path)
    return missing


def release_by_tag(api, tag):
    """Find one exact tag across authenticated release pages, then authenticate its actual release ID; return None if absent.
    在认证后的全部 Release 分页中查找唯一精确 tag，再认证其实际 Release ID；不存在时返回 None。
    """
    matches, ids = [], set()
    for page in range(1, 101):
        entries = api.request(f"releases?per_page=100&page={page}")
        require(type(entries) is list, "Release listing is not an array")
        for entry in entries:
            release_id = entry["id"]
            require(type(release_id) is int and release_id > 0 and release_id not in ids, "Duplicate or invalid release ID in listing")
            ids.add(release_id)
            if entry["tag_name"] == tag:
                matches.append(entry)
        if len(entries) < 100:
            break
    else:
        raise ValueError("Unbounded authenticated release listing")
    require(len(matches) <= 1, "Multiple releases match the exact tag")
    if not matches:
        return None
    release = api.request("releases/" + str(matches[0]["id"]))
    require(release["id"] == matches[0]["id"] and release["tag_name"] == tag, "Release ID/tag differs from authenticated listing")
    return release


def publish_assets(api, tag, commit, files, title, notes):
    """Publish only missing assets after authenticating exact tag and all conflict bytes; return status.
    认证精确标签及全部冲突字节后仅发布缺失资产；返回状态。
    """
    environment = dict(os.environ, GH_TOKEN=os.environ["GITHUB_TOKEN"])
    require(os.environ.get("GITHUB_ACTIONS") == "true", "Remote publication requires authorized workflow")
    created = False
    release = release_by_tag(api, tag)
    # Check the actual ref before creating, uploading or publishing any draft, independent of target_commitish.
    # 无论 target_commitish 如何，创建、上传或发布任何草稿前均检查实际 ref。
    chain = optional_tag(api, tag, commit)
    if release is None:
        # A draft is not returned by the published-release-by-tag endpoint; creation owns its actual ID.
        # 已发布 Release 的按 tag 端点不返回草稿；创建响应提供其实际 ID。
        response = api.request("releases", body={"tag_name": tag, "target_commitish": commit, "name": title,
                                                  "body": notes.read_text(encoding="utf-8"), "draft": True, "prerelease": False})
        require(type(response["id"]) is int and response["id"] > 0, "Draft creation did not return a valid release ID")
        release = api.request("releases/" + str(response["id"]))
        require(release["id"] == response["id"] and release["draft"] is True, "Created draft ID/state differs")
        created = True
    require(release["prerelease"] is False and release["tag_name"] == tag, "Unexpected release tag/state")
    if release["draft"] is False:
        require(chain is not None, "Final release tag is missing")
    else:
        require(release["draft"] is True and release["target_commitish"] == commit, "Draft target differs from frozen commit")
    existing = []
    for page in range(1, 101):
        entries = api.request(f"releases/{release['id']}/assets?per_page=100&page={page}")
        existing.extend(entries)
        if len(entries) < 100:
            break
    else:
        raise ValueError("Unbounded existing release assets")
    missing = assets_to_upload(existing, files, lambda entry: api.request("releases/assets/" + str(entry["id"]), binary=True))
    require(set(entry["name"] for entry in existing) <= {path.name for path in files}, "Release contains unexpected assets")
    if release["draft"] is False:
        require(not missing, "Immutable final release cannot receive additional assets")
        return {"created": False, "uploaded": [], "reused": [path.name for path in files], "server_immutable": release["immutable"]}
    for filename in missing:
        run(["gh", "release", "upload", tag, filename, "--repo", REPOSITORY], environment=environment)
    # Re-read all pages immediately before publication; a concurrent extra asset cannot enter a final release.
    # 在公开前立即重读全部分页，防止并发额外资产进入 final Release。
    uploaded = []
    for page in range(1, 101):
        entries = api.request(f"releases/{release['id']}/assets?per_page=100&page={page}")
        uploaded.extend(entries)
        if len(entries) < 100:
            break
    else:
        raise ValueError("Unbounded pre-publication release assets")
    require({entry["name"] for entry in uploaded} == {path.name for path in files}
            and not assets_to_upload(uploaded, files, lambda entry: api.request("releases/assets/" + str(entry["id"]), binary=True)),
            "Draft is missing exact publication bytes or contains unexpected assets")
    optional_tag(api, tag, commit)
    run(["gh", "release", "edit", tag, "--draft=false", "--repo", REPOSITORY], environment=environment)
    final = api.request("releases/" + str(release["id"]))
    require(final["id"] == release["id"] and final["tag_name"] == tag and final["draft"] is False
            and final["prerelease"] is False, "Published release is not final")
    resolve_tag(api, tag, commit)
    return {"created": created, "uploaded": [path.name for path in missing], "reused": [path.name for path in files if path not in missing], "server_immutable": final["immutable"]}




def validate_attestation_results(results, subjects, commit, run_id, attempt, workflow=".github/workflows/sdk-release.yml"):
    """Bind already cryptographically verified subjects and certificate identity to this exact issuer attempt.
    将已完成密码学验证的 subjects 及证书身份绑定到此精确 issuer attempt。
    """
    invocation = f"https://github.com/{REPOSITORY}/actions/runs/{run_id}/attempts/{attempt}"
    signer = f"https://github.com/{REPOSITORY}/{workflow}@"
    require(isinstance(results, list) and results, "No verified Go provenance results")
    for result in results:
        verified = result["verificationResult"]
        certificate = verified["signature"]["certificate"]
        require(certificate["issuer"] == "https://token.actions.githubusercontent.com"
                and certificate["sourceRepositoryURI"] == "https://github.com/" + REPOSITORY
                and certificate["sourceRepositoryDigest"] == certificate["buildSignerDigest"] == commit
                and certificate["buildSignerURI"].startswith(signer)
                and certificate["runnerEnvironment"] == "github-hosted"
                and certificate["runInvocationURI"] == invocation, "Go signed certificate issuer/source/workflow/run mismatch")
        statement = verified["statement"]
        require(statement["predicateType"] == "https://slsa.dev/provenance/v1"
                and statement["predicate"]["runDetails"]["metadata"]["invocationId"] == invocation,
                "Go signed invocation identity mismatch")
        signed = {subject["name"]: subject["digest"]["sha256"] for subject in statement["subject"]}
        require(len(signed) == len(statement["subject"]) and signed == subjects, "Go signed subject inventory mismatch")
    return {"verified_subjects": signed, "verified_invocation_uri": certificate["runInvocationURI"],
            "verified_source_sha": certificate["sourceRepositoryDigest"]}










def examples(args):
    """Build, install and run the six published examples from the authenticated formal Go batch.
    从已认证正式 Go 批次构建、安装并运行六个已发布示例。
    """
    formal = args.formal.resolve()
    proof = read_json(formal / "current/publish-proof.json")
    public = read_json(formal / "public/public-module.json")
    require(proof["accepted"] is True and public["accepted"] is True and public["plan"] == proof["plan"], "Current formal batch proof required for examples")
    plan = proof["plan"]
    validate_sdk_header(proof["npm"], plan, "npm", NPM_REPOSITORY)
    validate_sdk_header(proof["pypi"], plan, "pypi", PYPI_REPOSITORY)
    require(run(["git", "rev-parse", "HEAD"], ROOT).strip() == plan["sdk_sha"], "Examples source differs from formal SDK commit")
    authority = core_authority(args.core_root, plan["core_commit"])
    selected = host_platform(authority, "host")
    inputs = authority.resolve_sdk_inputs(formal / "current/core/prerequisites.json", selected)
    require(not args.output.exists(), "Examples output must be fresh")
    args.output.mkdir(parents=True)
    name = "luaskills-sdk-go-examples-" + plan["sdk_version"]
    package = args.output / name
    from prepare_examples_release import PUBLISHED_EXAMPLES, prepare, verify_archive, write_archive
    prepare(ROOT, package, plan["sdk_version"], plan["module_version"], proof["npm"]["sdk_version"])
    environment = {key: value for key, value in os.environ.items() if not key.lower().startswith("npm_config_")}
    environment.update(NPM_CONFIG_USERCONFIG=str(args.output / "empty.npmrc"), NPM_CONFIG_GLOBALCONFIG=str(args.output / "empty-global.npmrc"))
    (args.output / "empty.npmrc").write_text("", encoding="utf-8")
    (args.output / "empty-global.npmrc").write_text("", encoding="utf-8")
    installer = run(["npm", "exec", "--yes", "--registry=https://registry.npmjs.org", "--cache", args.output / "npm-cache",
                     "--package=@luaskills/sdk@" + plan["npm_version"], "--", "luaskills", "install-runtime", "--database", "none",
                     "--runtime-root", "examples/fixture-runtime"], package, environment)
    (args.output / "installer.log").write_text(installer, encoding="utf-8")
    library = package / "examples/fixture-runtime/libs" / Path(inputs["library"]).name
    require(library.is_file() and sha256(library) == inputs["library_sha256"], "Actual installed runtime library differs from formal core bytes")
    runtime = package / "examples/fixture-runtime"
    cache = args.output / "go-cache"
    environment.update(GOPROXY="https://proxy.golang.org", GOSUMDB="sum.golang.org", GOMODCACHE=str(cache.resolve()),
                       GOPRIVATE="", GONOPROXY="", GONOSUMDB="", GOTOOLCHAIN="local", GOWORK="off", GOFLAGS="", CGO_ENABLED="1",
                       CGO_LDFLAGS='"-L' + library.parent.as_posix() + '"', LD_LIBRARY_PATH=str(library.parent),
                       LUASKILLS_RUNTIME_ROOT=str(runtime), LUASKILLS_EXAMPLE_RUNTIME_ROOT=str(runtime))
    (args.output / "tidy.log").write_text(run(["go", "mod", "tidy"], package, environment), encoding="utf-8")
    imported = json.loads(run(["go", "list", "-json", MODULE_PATH], package, environment), object_pairs_hook=pairs)
    require(imported["Module"]["Version"] == plan["module_version"] and imported["Module"]["Path"] == MODULE_PATH
            and "Replace" not in imported["Module"] and Path(imported["Dir"]).resolve().is_relative_to(cache.resolve()), "Examples do not import the exact public module from a new cache")
    source = module_bytes(plan, formal / "proof/module")
    for member, body in source.items():
        require((Path(imported["Dir"]) / member).read_bytes() == body, "Examples public module member differs: " + member)
    for example in PUBLISHED_EXAMPLES:
        (args.output / (example + ".log")).write_text(run(["go", "run", "./examples/" + example], package, environment), encoding="utf-8")
    require(sha256(library) == inputs["library_sha256"], "Installed runtime bytes changed during examples")
    archive = args.output / (name + ".zip")
    write_archive(package, archive, name)
    validated = verify_archive(archive, name)
    sidecar = archive.with_suffix(".zip.sha256")
    sidecar.write_text(sha256(archive) + "  " + archive.name + "\n", encoding="utf-8")
    notes = args.output / "notes.md"
    notes.write_text(f"Go SDK {plan['sdk_version']} examples from {plan['sdk_sha']}; public Go module {plan['module_version']}, TypeScript runtime installer {plan['npm_version']}, core {plan['core_tag']} {plan['core_commit']}. All six bundled examples completed against exact installed core bytes.\n", encoding="utf-8")
    result = {"schema_version": SCHEMA, "plan": plan, "accepted": True, "archive": str(archive), "sha256": sha256(archive),
              "npm_version": plan["npm_version"], "library_sha256": inputs["library_sha256"], "validation": validated, "replace": False}
    write_json(args.output / "examples.json", result)
    return result


def main():
    """Parse strict evidence commands; no skip, fake consumer or alternate token switches exist.
    解析严格凭证命令；不存在跳过、伪消费者或备用 token 开关。
    """
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    flags = {
        "bootstrap": ("core-root", "core-tag", "core-commit", "platform", "output"),
        "freeze": ("root", "core-root", "core-tag", "core-commit", "sdk-sha", "workflow-sha", "mode",
                   "npm-version", "npm-sha", "npm-candidate-run-id", "npm-candidate-run-attempt", "npm-completion-run-id", "npm-completion-run-attempt", "npm-completion-source-sha",
                   "pypi-version", "pypi-sha", "pypi-candidate-run-id", "pypi-candidate-run-attempt", "pypi-completion-run-id", "pypi-completion-run-attempt", "pypi-completion-source-sha", "output"),
        "native": ("core-root", "plan", "module", "prerequisites", "platform", "output"),
        "aggregate": ("core-root", "plan", "module", "prerequisites", "reports", "output"),
        "prerequisites": ("core-root", "aggregate", "module", "prerequisites", "npm-root", "pypi-root", "platform", "output"),
        "publish-tag": ("core-root", "proof", "module", "output"),
        "public-module": ("core-root", "plan", "module", "prerequisites", "platform", "output"),
        "candidate-stage": ("core-root", "plan", "aggregate", "module", "prerequisites", "reports", "mode", "output"),
        "candidate-finalize": ("core-root", "core-commit", "directory", "bundle", "output"),
        "candidate-consume": ("core-root", "core-tag", "core-commit", "sdk-sha", "candidate-run-id", "candidate-run-attempt", "candidate-artifact-id", "output"),
        "completion-prepare": ("candidate", "core-root", "npm-root", "pypi-root", "intent", "output"),
        "completion-publish": ("directory", "bundle"),
        "formal-proof": ("core-root", "core-tag", "core-commit", "sdk-sha", "sdk-version", "candidate-run-id", "candidate-run-attempt",
                         "completion-run-id", "completion-run-attempt", "completion-source-sha", "npm-root", "pypi-root", "platform", "output"),
        "examples-stage": ("core-root", "formal", "examples", "mode", "output"),
        "examples-finalize": ("core-root", "formal", "directory", "bundle", "output"),
        "examples-publish": ("core-root", "formal", "intent", "examples-run-id", "examples-run-attempt", "examples-artifact-id", "output"),
        "examples": ("core-root", "formal", "output"),
    }
    path_flags = {"root", "core-root", "output", "plan", "module", "prerequisites", "reports", "aggregate", "npm-root", "npm-proof", "pypi-root", "pypi-proof", "proof", "public-proof", "publish-proof", "archive", "sidecar", "notes", "formal", "directory", "bundle", "candidate", "examples"}
    for name, options in flags.items():
        command = commands.add_parser(name)
        for option in options:
            command.add_argument("--" + option, required=True, type=Path if option in path_flags else str)
        if name == "freeze":
            command.add_argument("--github-output", type=Path)
    args = parser.parse_args()
    try:
        # Dual-chain commands live together; old single-run aliases are deliberately absent from this CLI.
        # 双链命令统一归属；此 CLI 明确不保留旧单运行别名。
        recovery_commands = {"candidate-stage", "candidate-finalize", "candidate-consume", "completion-prepare", "completion-publish", "formal-proof",
                             "examples-stage", "examples-finalize", "examples-publish"}
        if args.command in recovery_commands:
            import go_recovery
            result = getattr(go_recovery, args.command.replace("-", "_"))(args)
        else:
            result = globals()[args.command.replace("-", "_")](args)
        if result is not None:
            print(json.dumps(result, ensure_ascii=False))
    except (ValueError, KeyError, TypeError, OSError, subprocess.SubprocessError) as error:
        parser.exit(1, f"Go release gate failed: {error}\n")


if __name__ == "__main__":
    main()
