"""Keep signed candidate bytes separate from an explicitly authorized, independently signed completion.
将签名候选字节与显式授权、独立签名的完成凭证分开。
"""
import argparse
import hashlib
import importlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile

import go_release as gate

# Physical subject names are owned by this SDK; common recovery rules own their identity and inventory algorithm.
# 物理 subject 名称归本 SDK；共同恢复规则拥有其身份及清单算法。
WORKFLOW = ".github/workflows/sdk-release.yml"
CANDIDATE_MANIFEST = "go-candidate.json"
CANDIDATE_BUNDLE = "go-candidate-proof.tar.gz"
CANDIDATE_ATTESTATION = "go-candidate-attestation.jsonl"
CANDIDATE_PAYLOAD = {"go-module.zip", "go-aggregate.json", "go-core-prerequisites.json", CANDIDATE_BUNDLE, CANDIDATE_MANIFEST}
COMPLETION_MANIFEST = "go-completion.json"
COMPLETION_BUNDLE = "go-completion-proof.tar.gz"
COMPLETION_ATTESTATION = "go-completion-attestation.jsonl"
COMPLETION_PAYLOAD = {COMPLETION_MANIFEST, COMPLETION_BUNDLE, "go-completion-consumer.json",
                      "go-completion-core-prerequisites.json", "go-completion-prerequisites.json"}
# Examples have their own candidate signature and retain the original deterministic ZIP and sidecar.
# 示例拥有独立候选签名，并保留原确定性 ZIP 及 sidecar。
EXAMPLES_WORKFLOW = ".github/workflows/examples-release.yml"
EXAMPLES_MANIFEST = "go-examples-candidate.json"
EXAMPLES_ATTESTATION = "go-examples-attestation.jsonl"


def authorities(root, commit):
    """Return core and recovery modules from root's exact committed authority bytes at commit.
    返回 root 在精确 commit 上已提交权威字节对应的核心及恢复模块。
    """
    core = gate.core_authority(root, commit)
    relative = "scripts/release/sdk_recovery.py"
    expected = subprocess.check_output(["git", "show", commit + ":" + relative], cwd=root)
    gate.require((Path(root) / relative).read_bytes() == expected, "Recovery authority differs from frozen core commit")
    sys.modules.pop("sdk_recovery", None)
    return core, importlib.import_module("sdk_recovery")


def positive(value):
    """Parse a canonical positive numeric workflow input; return its integer identity.
    解析规范正数工作流输入；返回整数身份。
    """
    gate.require(str(value).isdigit() and str(value) == str(int(value)) and int(value) > 0, "Explicit positive run/attempt/artifact ID required")
    return int(value)


def digest(body):
    """Return the SHA-256 of exact byte body, never a reserialized JSON representation.
    返回精确字节 body 的 SHA-256，绝不对重序列化 JSON 求摘要。
    """
    return hashlib.sha256(body).hexdigest()


def fresh(directory):
    """Create a new directory for reviewable evidence; reject any prior output.
    为可审核凭证创建新目录；拒绝任何既有输出。
    """
    gate.require(not directory.exists(), "Recovery evidence output must be fresh")
    directory.mkdir(parents=True)


def root_files(directory):
    """Read only ordinary root files from directory; reject directories and symbolic links.
    仅读取 directory 的普通根文件；拒绝目录及符号链接。
    """
    result = {}
    for path in directory.iterdir():
        gate.require(path.is_file() and not path.is_symlink(), "Recovery root contains a non-file")
        result[path.name] = path.read_bytes()
    return result


def save_files(directory, files):
    """Persist exact safe root byte files into a fresh directory; return nothing.
    将精确安全根字节文件保存到新目录；无返回值。
    """
    fresh(directory)
    for name, body in files.items():
        gate.require(Path(name).name == name and name not in (".", ".."), "Evidence filename is not a root member")
        (directory / name).write_bytes(body)


def candidate_jobs(core):
    """Derive exact successful candidate job names from core's matrix and this SDK's frozen job declaration.
    从核心矩阵及本 SDK 冻结作业声明派生精确成功候选作业名称。
    """
    return ["freeze", "aggregate", "candidate-evidence"] + [gate.NATIVE_JOB_PREFIX + name for name in core.candidate.PLATFORMS]


def official_subjects(directory, names, attestation, source, run_id, attempt, workflow=WORKFLOW):
    """Cryptographically verify every physical subject and return identities normalized from actual verified output.
    对每个物理 subject 执行密码学验证，并返回由实际已验证输出归一化的身份。
    """
    subjects = {name: gate.sha256(directory / name) for name in names}
    environment = dict(os.environ, GH_TOKEN=os.environ["GITHUB_TOKEN"])
    normalized = None
    for name in sorted(names):
        output = gate.run(["gh", "attestation", "verify", directory / name, "--bundle", attestation,
                           "--repo", gate.REPOSITORY, "--signer-workflow", gate.REPOSITORY + "/" + workflow,
                           "--source-digest", source, "--signer-digest", source, "--deny-self-hosted-runners", "--format", "json"],
                          environment=environment)
        current = gate.validate_attestation_results(json.loads(output, object_pairs_hook=gate.pairs), subjects, source,
                                                    str(run_id), str(attempt), workflow=workflow)
        gate.require(normalized is None or current == normalized, "Official subject verification identities differ")
        normalized = current
    gate.require(normalized is not None, "No verified official subjects")
    return normalized


def tar_members(path, members):
    """Write exact selected members to a reviewable bundle; return each member's actual byte inventory.
    将所选精确成员写入可审核 bundle；返回每个成员的实际字节清单。
    """
    with tarfile.open(path, "w:gz") as archive:
        for name, source in sorted(members.items()):
            gate.require(source.is_file() and not source.is_symlink(), "Bundle member must be an actual ordinary file")
            archive.add(source, arcname="go-sdk-proof/" + name, recursive=False)
    return {name: {"size": source.stat().st_size, "sha256": gate.sha256(source)} for name, source in sorted(members.items())}


def directory_members(members, prefix, directory):
    """Extend selected members with ordinary files beneath one explicit evidence directory; return nothing.
    使用一个显式凭证目录下的普通文件扩展所选成员；无返回值。
    """
    for path in sorted(directory.rglob("*")):
        gate.require(not path.is_symlink(), "Evidence bundle cannot contain symbolic links")
        if path.is_file():
            members[prefix + "/" + path.relative_to(directory).as_posix()] = path


def checksum_subjects(directory, names):
    """Write the official action's exact filename/hash subjects outside the uploaded payload; return nothing.
    在上传载荷之外写入官方 action 的精确文件名／摘要 subjects；无返回值。
    """
    (directory / "subjects.sha256").write_text("".join(gate.sha256(directory / name) + "  " + name + "\n" for name in sorted(names)), encoding="utf-8")


def candidate_stage(args):
    """Freeze tested package/native/core evidence and a direct schema-2 binding before any public mutation.
    在任何公开写入前冻结已测试包／原生／核心证据及直接 schema-2 绑定。
    """
    core, recovery = authorities(args.core_root, gate.read_json(args.plan)["core_commit"])
    plan, aggregate = gate.read_json(args.plan), gate.read_json(args.aggregate)
    gate.require(aggregate["accepted"] is True and aggregate["plan"] == plan, "Candidate requires the exact complete native aggregate")
    gate.module_bytes(plan, args.module)
    gate.require(args.mode in ("artifact-only", "publish"), "Original candidate mode cannot be recovery")
    gate.require(plan["sdk_sha"] == os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA")
                 and os.environ.get("GITHUB_JOB") == "candidate-evidence", "Candidate signing must use the exact source evidence job")
    run_id, attempt = positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"])
    fresh(args.output)
    # Sign only a matrix rebuilt from actual raw native logs, rather than trusting an earlier JSON assertion.
    # 仅签署从实际原始原生日志重建的矩阵，避免信任先前 JSON 断言。
    rebuilt = gate.aggregate(argparse.Namespace(core_root=args.core_root, plan=args.plan, module=args.module,
                                               prerequisites=args.prerequisites, reports=args.reports,
                                               output=args.output / "checked-aggregate.json"))
    gate.require(rebuilt == aggregate, "Candidate signing aggregate differs from actual native receipts")
    members = {"plan.json": args.plan, "aggregate.json": args.aggregate}
    for prefix, directory in (("module", args.module), ("core", args.prerequisites.parent), ("reports", args.reports)):
        directory_members(members, prefix, directory)
    inventory = tar_members(args.output / CANDIDATE_BUNDLE, members)
    for name, path in (("go-module.zip", args.module / "module.zip"), ("go-aggregate.json", args.aggregate),
                       ("go-core-prerequisites.json", args.prerequisites)):
        shutil.copyfile(path, args.output / name)
    manifest = {"schema_version": recovery.SCHEMA_VERSION, "kind": "go-sdk-candidate", "sdk_source_sha": plan["sdk_sha"],
                "sdk_version": plan["sdk_version"], "core_tag": plan["core_tag"], "core_commit": plan["core_commit"],
                "candidate_run_id": str(run_id), "candidate_run_attempt": attempt, "candidate_mode": args.mode,
                "required_jobs": candidate_jobs(core), "members": inventory}
    gate.write_json(args.output / CANDIDATE_MANIFEST, manifest)
    payload = {name: (args.output / name).read_bytes() for name in CANDIDATE_PAYLOAD}
    binding = {"schema_version": recovery.SCHEMA_VERSION, "kind": "sdk-candidate", "repository": gate.REPOSITORY,
               "workflow_path": WORKFLOW, "source_sha": plan["sdk_sha"], "run_id": run_id, "run_attempt": attempt,
               "artifact_name": recovery.candidate_artifact_name(run_id, attempt), "inventory": recovery.inventory_for(payload)}
    gate.write_json(args.output / recovery.BINDING_FILENAME, binding)
    checksum_subjects(args.output, CANDIDATE_PAYLOAD | {recovery.BINDING_FILENAME})
    return {"artifact_name": binding["artifact_name"], "candidate_only": True}


def verify_candidate(directory, core, recovery, source, run_id, attempt):
    """Verify exact candidate signatures, independent binding and all physical payload bytes; return authenticated facts.
    验证精确候选签名、独立绑定及全部物理载荷字节；返回已认证事实。
    """
    files = root_files(directory)
    gate.require(set(files) == CANDIDATE_PAYLOAD | {recovery.BINDING_FILENAME, CANDIDATE_ATTESTATION}, "Candidate physical subject inventory differs")
    normalized = official_subjects(directory, CANDIDATE_PAYLOAD | {recovery.BINDING_FILENAME}, directory / CANDIDATE_ATTESTATION,
                                   source, run_id, attempt)
    verified = recovery.verify_signed_binding(files[recovery.BINDING_FILENAME], **normalized, repository=gate.REPOSITORY,
                                              workflow_path=WORKFLOW, source_sha=source, run_id=run_id, run_attempt=attempt,
                                              artifact_name=recovery.candidate_artifact_name(run_id, attempt))
    inventory = recovery.verify_inventory(verified, files, attestation_filename=CANDIDATE_ATTESTATION)
    manifest = json.loads(files[CANDIDATE_MANIFEST], object_pairs_hook=gate.pairs)
    gate.require(manifest["schema_version"] == recovery.SCHEMA_VERSION and manifest["kind"] == "go-sdk-candidate"
                 and manifest["sdk_source_sha"] == source and manifest["candidate_run_id"] == str(run_id)
                 and type(manifest["candidate_run_attempt"]) is int and manifest["candidate_run_attempt"] == attempt
                 and manifest["candidate_mode"] in ("artifact-only", "publish") and manifest["required_jobs"] == candidate_jobs(core),
                 "Authenticated candidate source/attempt/mode/job identity differs")
    return {"manifest": manifest, "binding": verified, "inventory": inventory}


def candidate_finalize(args):
    """Verify the signing job's own completed action output and persist only exact candidate artifact files.
    验证签名作业自身已完成 action 输出，并仅保存精确候选制品文件。
    """
    core, recovery = authorities(args.core_root, args.core_commit)
    manifest = gate.read_json(args.directory / CANDIDATE_MANIFEST)
    gate.require(manifest["sdk_source_sha"] == os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA")
                 and manifest["candidate_run_id"] == os.environ.get("GITHUB_RUN_ID")
                 and manifest["candidate_run_attempt"] == positive(os.environ["GITHUB_RUN_ATTEMPT"]), "Candidate action output belongs to another invocation")
    files = {name: (args.directory / name).read_bytes() for name in CANDIDATE_PAYLOAD | {recovery.BINDING_FILENAME}}
    files[CANDIDATE_ATTESTATION] = args.bundle.read_bytes()
    save_files(args.output, files)
    verify_candidate(args.output, core, recovery, manifest["sdk_source_sha"], positive(manifest["candidate_run_id"]), manifest["candidate_run_attempt"])
    # The evidence job is still running, so its own future success is not asserted here.
    # 凭证作业仍在运行，因此此处不声明其未来成功。
    return {"artifact_name": recovery.candidate_artifact_name(positive(manifest["candidate_run_id"]), manifest["candidate_run_attempt"])}


def extract_bundle(core, body, manifest, directory):
    """Extract only exact authenticated bundle members matching manifest; reject any member/content difference.
    仅解压匹配 manifest 的精确已认证 bundle 成员；拒绝任何成员／内容差异。
    """
    files = core.rooted_archive(body, "go-sdk-proof")
    gate.require(set(files) == set(manifest["members"]), "Signed evidence bundle member inventory differs")
    fresh(directory)
    for name, body in files.items():
        gate.require(len(body) == manifest["members"][name]["size"] and digest(body) == manifest["members"][name]["sha256"], "Signed bundle member changed: " + name)
        path = directory / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(body)


def reaggregate_candidate(core_root, original, proof, output):
    """Recheck exact flat package/core/aggregate originals and native receipts from the authenticated bundle.
    从已认证 bundle 重新检查精确平面包／核心／汇总原件及原生凭证。
    """
    for asset, member in (("go-module.zip", "module/module.zip"), ("go-aggregate.json", "aggregate.json"),
                          ("go-core-prerequisites.json", "core/prerequisites.json")):
        gate.require((original / asset).read_bytes() == (proof / member).read_bytes(), "Signed flat subject differs from bundled original: " + asset)
    rebuilt = gate.aggregate(argparse.Namespace(core_root=core_root, plan=proof / "plan.json", module=proof / "module",
                                               prerequisites=proof / "core/prerequisites.json", reports=proof / "reports", output=output))
    gate.require(rebuilt == gate.read_json(proof / "aggregate.json"), "Original matrix receipts differ from signed aggregate")
    return rebuilt


def check_attempt_intent(attempt, source, mode):
    """Require actual SDK dispatch event and source-owned displayed mode without rewriting original status.
    要求实际 SDK dispatch 事件及源码所属展示模式，不改写原始状态。
    """
    record = attempt["attempt"]
    gate.require(record["event"] == "workflow_dispatch" and record["display_title"] == f"Go SDK {source} mode={mode}",
                 "SDK attempt event or explicit invocation mode differs")


def candidate_consume(args):
    """Download one explicit original artifact and authenticate its exact successful gate jobs/signatures before writes.
    下载一个显式原制品，在写入前认证其精确成功门禁作业及签名。
    """
    core, recovery = authorities(args.core_root, args.core_commit)
    run_id, attempt, artifact_id = positive(args.candidate_run_id), positive(args.candidate_run_attempt), positive(args.candidate_artifact_id)
    artifact_name = recovery.candidate_artifact_name(run_id, attempt)
    metadata, files = recovery.download_artifact(core.Http(), repository=gate.REPOSITORY, source_sha=args.sdk_sha,
                                               run_id=run_id, artifact_id=artifact_id, artifact_name=artifact_name)
    fresh(args.output)
    save_files(args.output / "original", files)
    facts = verify_candidate(args.output / "original", core, recovery, args.sdk_sha, run_id, attempt)
    manifest = facts["manifest"]
    gate.require(manifest["core_tag"] == args.core_tag and manifest["core_commit"] == args.core_commit, "Candidate core identity differs from explicit input")
    actual = recovery.verify_attempt(core.Http(), repository=gate.REPOSITORY, workflow_path=WORKFLOW, source_sha=args.sdk_sha,
                                     run_id=run_id, run_attempt=attempt, required_jobs=candidate_jobs(core), phase="candidate")
    check_attempt_intent(actual, args.sdk_sha, manifest["candidate_mode"])
    extract_bundle(core, files[CANDIDATE_BUNDLE], manifest, args.output / "proof")
    reaggregate_candidate(args.core_root, args.output / "original", args.output / "proof", args.output / "aggregate.json")
    result = {"schema_version": recovery.SCHEMA_VERSION, "manifest": manifest, "binding": facts["binding"], "attempt": actual,
              "artifact": metadata, "artifact_id": artifact_id, "physical_inventory": recovery.inventory_for(files)}
    gate.write_json(args.output / "candidate.json", result)
    return result


def production_context(plan, intent):
    """Authenticate this actual production dispatch and OIDC-capable completion job; return its unmodified attempt record.
    认证本次实际生产 dispatch 及具备 OIDC 权限的 completion 作业；返回未改写轮次记录。
    """
    gate.require(intent in ("publish", "recover") and os.environ.get("GITHUB_ACTIONS") == "true"
                 and os.environ.get("GITHUB_JOB") == "completion" and os.environ.get("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
                 and os.environ.get("ACTIONS_ID_TOKEN_REQUEST_URL"), "Explicit production completion with official OIDC capability required")
    source = plan["sdk_sha"]
    gate.require(source == os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA"), "Initial completion source must equal original SDK source")
    gate.require(gate.run(["git", "rev-parse", "HEAD"], gate.ROOT).strip() == source
                 and not gate.run(["git", "status", "--porcelain", "--untracked-files=all"], gate.ROOT).strip()
                 and gate.source_metadata(gate.ROOT, plan["core_tag"]) == plan["sdk_version"], "Completion checkout/version is not the exact clean original source")
    api = gate.GitHub()
    gate.trusted_workflow(api, source)
    run_id, attempt = positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"])
    record = api.request(f"actions/runs/{run_id}/attempts/{attempt}")
    gate.require(type(record["id"]) is int and record["id"] == run_id and type(record["run_attempt"]) is int
                 and record["run_attempt"] == attempt and record["head_sha"] == source
                 and record["repository"]["full_name"] == record["head_repository"]["full_name"] == gate.REPOSITORY
                 and record["event"] == "workflow_dispatch" and record["display_title"] == f"Go SDK {source} mode={intent}",
                 "Current production dispatch source/attempt/intent differs")
    gate.require(api.request("actions/workflows/" + str(record["workflow_id"]))["path"] == WORKFLOW,
                 "Current completion uses another workflow")
    return record


def release_files(api, tag, source, expected_names):
    """Authenticate a public exact tag/release and every expected asset byte; return release and physical files.
    认证公开精确 tag／Release 及每个预期资产字节；返回 Release 与物理文件。
    """
    gate.resolve_tag(api, tag, source)
    release = api.request("releases/tags/" + tag)
    gate.require(release["tag_name"] == tag and release["draft"] is False and release["prerelease"] is False
                 and type(release["id"]) is int and release["id"] > 0, "Exact public release identity required")
    assets = []
    for page in range(1, 101):
        entries = api.request(f"releases/{release['id']}/assets?per_page=100&page={page}")
        assets.extend(entries)
        if len(entries) < 100:
            break
    else:
        raise ValueError("Unbounded SDK release asset pages")
    gate.require(len({entry["name"] for entry in assets}) == len(assets)
                 and {entry["name"] for entry in assets} == set(expected_names), "Public SDK release has missing/extra/duplicate assets")
    files = {}
    for entry in assets:
        body = api.request("releases/assets/" + str(entry["id"]), binary=True)
        gate.require(type(body) is bytes and len(body) == entry["size"], "Actual public release asset size differs")
        files[entry["name"]] = body
    return release, files


def candidate_links(recovery, receipt, files):
    """Return immutable original source/artifact identities and exact signed bundle/manifest/signature digests.
    返回不可变原始源码／制品身份及精确签名 bundle／清单／签名摘要。
    """
    manifest = receipt["manifest"]
    return {"sdk_source_sha": manifest["sdk_source_sha"], "candidate_run_id": manifest["candidate_run_id"],
            "candidate_run_attempt": manifest["candidate_run_attempt"], "candidate_artifact_id": receipt["artifact_id"],
            "candidate_artifact_name": recovery.candidate_artifact_name(positive(manifest["candidate_run_id"]), manifest["candidate_run_attempt"]),
            "candidate_binding_sha256": digest(files[recovery.BINDING_FILENAME]), "candidate_manifest_sha256": digest(files[CANDIDATE_MANIFEST]),
            "candidate_bundle_sha256": digest(files[CANDIDATE_BUNDLE]), "candidate_attestation_sha256": digest(files[CANDIDATE_ATTESTATION]),
            "candidate_physical_inventory": recovery.inventory_for(files)}


def completion_prepare(args):
    """Publish only original authenticated candidates, then cold-consume public Go and stage independent completion subjects.
    仅发布原已认证候选，随后冷消费公开 Go 并准备独立完成 subjects。
    """
    receipt = gate.read_json(args.candidate / "candidate.json")
    manifest = receipt["manifest"]
    core, recovery = authorities(args.core_root, manifest["core_commit"])
    original, proof = args.candidate / "original", args.candidate / "proof"
    facts = verify_candidate(original, core, recovery, manifest["sdk_source_sha"], positive(manifest["candidate_run_id"]), manifest["candidate_run_attempt"])
    gate.require(facts["manifest"] == manifest and facts["binding"] == receipt["binding"], "Original authenticated receipt changed before completion")
    files = root_files(original)
    recovery.compare_inventory(receipt["physical_inventory"], files)
    plan = gate.read_json(proof / "plan.json")
    production_context(plan, args.intent)
    fresh(args.output)
    reaggregate_candidate(args.core_root, original, proof, args.output / "aggregate.json")
    # Fresh complete Cargo and exact formal SDK consumers precede the first tag or Release mutation.
    # 新完整 Cargo 及精确正式 SDK 消费者先于首个 tag 或 Release 写入。
    gate.prerequisites(argparse.Namespace(core_root=args.core_root, aggregate=args.output / "aggregate.json", module=proof / "module",
                                         prerequisites=proof / "core/prerequisites.json", npm_root=args.npm_root, pypi_root=args.pypi_root,
                                         platform="host", output=args.output / "current"))
    gate.publish_tag(argparse.Namespace(core_root=args.core_root, proof=args.output / "current/publish-proof.json", module=proof / "module",
                                       output=args.output / "tag.json"))
    notes = args.output / "main-notes.md"
    notes.write_text(f"Original signed Go SDK {plan['sdk_version']} candidate {manifest['candidate_run_id']}/{manifest['candidate_run_attempt']} from {plan['sdk_sha']}. Completion evidence is published independently.\n", encoding="utf-8")
    gate.publish_assets(gate.GitHub(), plan["module_version"], plan["sdk_sha"], tuple(original / name for name in sorted(files)),
                        "LuaSkills Go SDK " + plan["sdk_version"], notes)
    release, published = release_files(gate.GitHub(), plan["module_version"], plan["sdk_sha"], files)
    gate.require(published == files, "Main SDK release differs from original authenticated candidate bytes")
    public = gate.public_module(argparse.Namespace(core_root=args.core_root, plan=proof / "plan.json", module=proof / "module",
                                                  prerequisites=args.output / "current/core/prerequisites.json", platform="host", output=args.output / "public"))
    stage = args.output / "stage"
    fresh(stage)
    members = {"public/public-module.json": args.output / "public/public-module.json", "current/publish-proof.json": args.output / "current/publish-proof.json"}
    for name in ("go-mod-download.json", "go-list.json", "module-members.json", "native-summary.json", "tests.jsonl", "example-tests.jsonl", "example.log"):
        members["public/" + name] = args.output / "public" / name
    directory_members(members, "core", args.output / "current/core")
    for kind in ("npm", "pypi"):
        header = gate.read_json(args.output / "current" / kind / "accepted.json")
        members[kind + "/accepted.json"] = args.output / "current" / kind / "accepted.json"
        members[kind + "/" + header["registry_consumer_file"]] = args.output / "current" / kind / header["registry_consumer_file"]
    bundled = tar_members(stage / COMPLETION_BUNDLE, members)
    for name, source in (("go-completion-consumer.json", args.output / "public/public-module.json"),
                         ("go-completion-core-prerequisites.json", args.output / "current/core/prerequisites.json"),
                         ("go-completion-prerequisites.json", args.output / "current/publish-proof.json")):
        shutil.copyfile(source, stage / name)
    completion = {"schema_version": recovery.SCHEMA_VERSION, "kind": "go-sdk-completion", "sdk_version": plan["sdk_version"],
                  "core_tag": plan["core_tag"], "core_commit": plan["core_commit"], "original": candidate_links(recovery, receipt, files),
                  "completion_source_sha": plan["sdk_sha"], "completion_run_id": str(positive(os.environ["GITHUB_RUN_ID"])),
                  "completion_run_attempt": positive(os.environ["GITHUB_RUN_ATTEMPT"]), "completion_intent": args.intent,
                  "main_release": {"id": release["id"], "tag": release["tag_name"], "source_sha": plan["sdk_sha"],
                                   "inventory": recovery.inventory_for(published), "server_immutable": release["immutable"]},
                  "consumer_sha256": digest((stage / "go-completion-consumer.json").read_bytes()), "members": bundled}
    gate.write_json(stage / COMPLETION_MANIFEST, completion)
    checksum_subjects(stage, COMPLETION_PAYLOAD)
    return {"completion_source_sha": plan["sdk_sha"], "stage": str(stage), "original_candidate_mode": manifest["candidate_mode"], "public_consumer": public}


def completion_tag(version, run_id, attempt):
    """Return this SDK's independent completion release tag using its actual positive run and attempt.
    根据实际正数运行及轮次返回本 SDK 独立完成 Release 标签。
    """
    return f"recovery-v{version}-r{positive(run_id)}-a{positive(attempt)}"


def completion_publish(args):
    """Verify official completion subjects before publishing a full independent draft; return actual publication receipt.
    发布完整独立草稿前验证官方完成 subjects；返回实际发布凭证。
    """
    manifest = gate.read_json(args.directory / COMPLETION_MANIFEST)
    gate.require(manifest["completion_source_sha"] == manifest["original"]["sdk_source_sha"]
                 == os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA")
                 and manifest["completion_run_id"] == os.environ.get("GITHUB_RUN_ID")
                 and manifest["completion_run_attempt"] == positive(os.environ["GITHUB_RUN_ATTEMPT"]), "Completion signer identity differs from original/current source")
    production_context({"sdk_sha": manifest["completion_source_sha"], "sdk_version": manifest["sdk_version"], "core_tag": manifest["core_tag"]}, manifest["completion_intent"])
    attestation = args.directory / COMPLETION_ATTESTATION
    gate.require(not attestation.exists(), "Completion attestation output already exists")
    shutil.copyfile(args.bundle, attestation)
    official_subjects(args.directory, COMPLETION_PAYLOAD, attestation, manifest["completion_source_sha"],
                      positive(manifest["completion_run_id"]), manifest["completion_run_attempt"])
    tag = completion_tag(manifest["sdk_version"], manifest["completion_run_id"], manifest["completion_run_attempt"])
    notes = args.directory / "notes.md"
    notes.write_text(f"Explicit {manifest['completion_intent']} completion of original candidate {manifest['original']['candidate_run_id']}/{manifest['original']['candidate_run_attempt']}; independent signer {manifest['completion_run_id']}/{manifest['completion_run_attempt']}.\n", encoding="utf-8")
    files = tuple(args.directory / name for name in sorted(COMPLETION_PAYLOAD | {COMPLETION_ATTESTATION}))
    result = gate.publish_assets(gate.GitHub(), tag, manifest["completion_source_sha"], files, "LuaSkills Go SDK completion " + manifest["sdk_version"], notes)
    release, published = release_files(gate.GitHub(), tag, manifest["completion_source_sha"], COMPLETION_PAYLOAD | {COMPLETION_ATTESTATION})
    gate.require(all(published[path.name] == path.read_bytes() for path in files), "Published completion bytes differ from official signed originals")
    return {**result, "completion_release_id": release["id"], "completion_tag": tag}


def verify_completion_links(manifest, receipt, files, release, recovery, source, run_id, attempt):
    """Require signed completion to bind original hashes, exact main Release and separate completion identity.
    要求签名完成凭证绑定原始摘要、精确主 Release 及独立完成身份。
    """
    original = receipt["manifest"]
    gate.require(manifest["schema_version"] == recovery.SCHEMA_VERSION and manifest["kind"] == "go-sdk-completion"
                 and manifest["sdk_version"] == original["sdk_version"] and manifest["core_tag"] == original["core_tag"]
                 and manifest["core_commit"] == original["core_commit"]
                 and manifest["completion_source_sha"] == source == original["sdk_source_sha"]
                 and manifest["completion_run_id"] == str(run_id) and type(manifest["completion_run_attempt"]) is int
                 and manifest["completion_run_attempt"] == attempt and manifest["completion_intent"] in ("publish", "recover"),
                 "Signed completion source/attempt/intent identity differs")
    gate.require(manifest["original"] == candidate_links(recovery, receipt, files), "Signed completion original bundle/manifest/attestation/artifact binding differs")
    main = manifest["main_release"]
    gate.require(main["id"] == release["id"] and type(main["id"]) is int and main["tag"] == release["tag_name"]
                 and main["source_sha"] == source and main["inventory"] == recovery.inventory_for(files), "Completion main Release/source/package inventory differs")


def completion_attempt(core, recovery, source, run_id, attempt):
    """Wait only for this exact active attempt, then require actual completed success and the completion job.
    仅等待此精确活动轮次，随后要求实际 completed success 及 completion 作业。
    """
    http = core.Http()
    endpoint = f"https://api.github.com/repos/{gate.REPOSITORY}/actions/runs/{run_id}/attempts/{attempt}"
    deadline = gate.time.monotonic() + 600
    while True:
        record = http.json(endpoint)
        gate.require(record["head_sha"] == source and record["id"] == run_id and record["run_attempt"] == attempt,
                     "Completion attempt changed while waiting")
        if record["status"] == "completed":
            break
        gate.require(record["status"] == "in_progress" and record["conclusion"] is None and gate.time.monotonic() < deadline,
                     "Specified completion attempt is not successful or timed out")
        gate.time.sleep(5)
    return recovery.verify_attempt(http, repository=gate.REPOSITORY, workflow_path=WORKFLOW, source_sha=source,
                                   run_id=run_id, run_attempt=attempt, required_jobs=["completion"], phase="completion")


def validate_completion_evidence(core, candidate_proof, completion_proof, flat, manifest):
    """Validate signed fresh core/consumer bytes and actual native logs against the original plan; return public facts.
    对照原 plan 验证签名新核心／消费者字节及实际原生日志；返回公开消费事实。
    """
    plan = gate.read_json(candidate_proof / "plan.json")
    for asset, member in (("go-completion-consumer.json", "public/public-module.json"),
                          ("go-completion-core-prerequisites.json", "core/prerequisites.json"),
                          ("go-completion-prerequisites.json", "current/publish-proof.json")):
        gate.require((flat / asset).read_bytes() == (completion_proof / member).read_bytes(), "Completion flat subject differs from bundled actual consumer")
    public, current = gate.read_json(flat / "go-completion-consumer.json"), gate.read_json(flat / "go-completion-prerequisites.json")
    gate.require(public["accepted"] is True and current["accepted"] is True and public["plan"] == current["plan"] == plan
                 and public["public_proxy"] == "https://proxy.golang.org" and public["replace"] is False
                 and gate.sha256(flat / "go-completion-consumer.json") == manifest["consumer_sha256"], "Signed public Go consumer differs from original package/plan")
    gate.complete_core(gate.read_json(completion_proof / "core/prerequisites.json"), plan, core)
    gate.require(current["core"] == gate.read_json(completion_proof / "core/prerequisites.json"), "Completion current core consumer differs")
    for kind, repository in (("npm", gate.NPM_REPOSITORY), ("pypi", gate.PYPI_REPOSITORY)):
        gate.validate_sdk_header(current[kind], plan, kind, repository)
        header = gate.read_json(completion_proof / kind / "accepted.json")
        gate.require(header == current[kind] and gate.sha256(completion_proof / kind / header["registry_consumer_file"]) == header["registry_consumer_sha256"],
                     "Signed independent SDK consumer bytes differ")
    inputs = core.resolve_sdk_inputs(completion_proof / "core/prerequisites.json", public["platform"])
    logs = {name: (completion_proof / "public" / name).read_bytes() for name in ("tests.jsonl", "example-tests.jsonl", "example.log")}
    summary = gate.read_json(completion_proof / "public/native-summary.json")
    public_plan = {**plan, "module_zip_sha256": public["public_zip_sha256"]}
    gate.require(gate.validate_native(summary, public_plan, inputs, logs) == public["native"], "Signed public native summary differs from actual logs")
    return public


def formal_proof(args):
    """Authenticate permanent original and successful completion chains, then execute fresh public consumers.
    认证永久原始链及成功完成链，随后执行新鲜公开消费者。
    """
    gate.require(gate.run(["git", "rev-parse", "HEAD"], gate.ROOT).strip() == gate.full_sha(args.sdk_sha)
                 and not gate.run(["git", "status", "--porcelain", "--untracked-files=all"], gate.ROOT).strip()
                 and args.completion_source_sha == args.sdk_sha
                 and gate.source_metadata(gate.ROOT, args.core_tag) == args.sdk_version, "Formal dual-chain verifier source/version differs")
    core, recovery = authorities(args.core_root, args.core_commit)
    run_id, attempt = positive(args.candidate_run_id), positive(args.candidate_run_attempt)
    complete_id, complete_attempt = positive(args.completion_run_id), positive(args.completion_run_attempt)
    api = gate.GitHub()
    main, original = release_files(api, "v" + args.sdk_version, args.sdk_sha, CANDIDATE_PAYLOAD | {recovery.BINDING_FILENAME, CANDIDATE_ATTESTATION})
    complete_release, completion = release_files(api, completion_tag(args.sdk_version, complete_id, complete_attempt), args.completion_source_sha,
                                                 COMPLETION_PAYLOAD | {COMPLETION_ATTESTATION})
    fresh(args.output)
    save_files(args.output / "original", original)
    save_files(args.output / "completion-original", completion)
    facts = verify_candidate(args.output / "original", core, recovery, args.sdk_sha, run_id, attempt)
    manifest = facts["manifest"]
    gate.require(manifest["core_tag"] == args.core_tag and manifest["core_commit"] == args.core_commit and manifest["sdk_version"] == args.sdk_version,
                 "Original candidate core/version differs")
    actual_candidate = recovery.verify_attempt(core.Http(), repository=gate.REPOSITORY, workflow_path=WORKFLOW, source_sha=args.sdk_sha,
                                               run_id=run_id, run_attempt=attempt, required_jobs=candidate_jobs(core), phase="candidate")
    check_attempt_intent(actual_candidate, args.sdk_sha, manifest["candidate_mode"])
    official_subjects(args.output / "completion-original", COMPLETION_PAYLOAD, args.output / "completion-original" / COMPLETION_ATTESTATION,
                      args.completion_source_sha, complete_id, complete_attempt)
    completed = json.loads(completion[COMPLETION_MANIFEST], object_pairs_hook=gate.pairs)
    receipt = {"manifest": manifest, "binding": facts["binding"], "artifact_id": completed["original"]["candidate_artifact_id"]}
    positive(receipt["artifact_id"])
    verify_completion_links(completed, receipt, original, main, recovery, args.completion_source_sha, complete_id, complete_attempt)
    actual_completion = completion_attempt(core, recovery, args.completion_source_sha, complete_id, complete_attempt)
    check_attempt_intent(actual_completion, args.completion_source_sha, completed["completion_intent"])
    extract_bundle(core, original[CANDIDATE_BUNDLE], manifest, args.output / "proof")
    reaggregate_candidate(args.core_root, args.output / "original", args.output / "proof", args.output / "aggregate.json")
    extract_bundle(core, completion[COMPLETION_BUNDLE], completed, args.output / "completion-proof")
    validate_completion_evidence(core, args.output / "proof", args.output / "completion-proof", args.output / "completion-original", completed)
    # Permanent issuer receipts remain intact; this fresh consumer has a separate physical name.
    # 永久签发者回执保持原样；本次新消费者使用独立物理文件名。
    proof = args.output / "proof"
    gate.prerequisites(argparse.Namespace(core_root=args.core_root, aggregate=args.output / "aggregate.json", module=proof / "module",
                                         prerequisites=proof / "core/prerequisites.json", npm_root=args.npm_root, pypi_root=args.pypi_root,
                                         platform=args.platform, output=args.output / "current"))
    gate.public_module(argparse.Namespace(core_root=args.core_root, plan=proof / "plan.json", module=proof / "module",
                                         prerequisites=args.output / "current/core/prerequisites.json", platform=args.platform, output=args.output / "public"))
    consumer = args.output / "fresh-formal-consumer.json"
    shutil.copyfile(args.output / "public/public-module.json", consumer)
    accepted = {"schema_version": recovery.SCHEMA_VERSION, "sdk_source_sha": args.sdk_sha, "sdk_version": args.sdk_version,
                "core_tag": args.core_tag, "core_commit": args.core_commit, "candidate_run_id": str(run_id), "candidate_run_attempt": attempt,
                "completion_run_id": str(complete_id), "completion_run_attempt": complete_attempt, "completion_source_sha": args.completion_source_sha,
                "repository": gate.REPOSITORY, "accepted": True, "registry_consumer_file": consumer.name, "registry_consumer_sha256": gate.sha256(consumer)}
    gate.write_json(args.output / "formal-proof.json", {"accepted": accepted, "candidate_attempt": actual_candidate, "completion_attempt": actual_completion,
                                                      "main_release_id": main["id"], "completion_release_id": complete_release["id"],
                                                      "main_server_immutable": main["immutable"], "completion_server_immutable": complete_release["immutable"]})
    gate.write_json(args.output / "accepted.json", accepted)
    return accepted


def formal_header(formal):
    """Authenticate one completed local formal-consumer receipt and return its immutable dual identity fields.
    认证一个已完成本地正式消费者回执，并返回其不可变双身份字段。
    """
    header = gate.read_json(formal / "accepted.json")
    gate.require(header["schema_version"] == gate.SDK_FORMAL_SCHEMA and header["accepted"] is True
                 and header["repository"] == gate.REPOSITORY and header["sdk_source_sha"] == header["completion_source_sha"]
                 and header["registry_consumer_file"] == "fresh-formal-consumer.json"
                 and gate.sha256(formal / header["registry_consumer_file"]) == header["registry_consumer_sha256"],
                 "Actual fresh dual-chain Go consumer is mandatory for examples")
    keys = ("schema_version", "sdk_source_sha", "sdk_version", "core_tag", "core_commit", "candidate_run_id",
            "candidate_run_attempt", "completion_run_id", "completion_run_attempt", "completion_source_sha", "repository", "accepted")
    return {key: header[key] for key in keys}


def example_names(version):
    """Return exact published ZIP/sidecar/manifest subject names for the Go SDK version.
    根据 Go SDK version 返回精确已发布 ZIP／sidecar／清单 subject 名称。
    """
    archive = "luaskills-sdk-go-examples-" + version + ".zip"
    return {archive, archive + ".sha256", EXAMPLES_MANIFEST}


def examples_stage(args):
    """Stage one actually tested deterministic examples ZIP and original sidecar for independent official signing.
    为独立官方签名准备一个已实际测试的确定性示例 ZIP 及原 sidecar。
    """
    header = formal_header(args.formal)
    core, recovery = authorities(args.core_root, header["core_commit"])
    gate.require(args.mode in ("artifact-only", "publish") and os.environ.get("GITHUB_JOB") == "examples-candidate"
                 and os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA") == header["sdk_source_sha"],
                 "Examples signing source/job/mode differs")
    examples = gate.read_json(args.examples / "examples.json")
    gate.require(examples["accepted"] is True and examples["plan"]["sdk_sha"] == header["sdk_source_sha"]
                 and examples["plan"]["sdk_version"] == header["sdk_version"], "Actual six-example receipt differs from formal SDK")
    version = header["sdk_version"]
    archive = args.examples / ("luaskills-sdk-go-examples-" + version + ".zip")
    sidecar = archive.with_suffix(".zip.sha256")
    gate.require(gate.sha256(archive) == examples["sha256"] and sidecar.read_text(encoding="utf-8").strip()
                 == examples["sha256"] + "  " + archive.name, "Original examples ZIP/sidecar bytes differ")
    from prepare_examples_release import verify_archive
    verify_archive(archive, "luaskills-sdk-go-examples-" + version)
    fresh(args.output)
    shutil.copyfile(archive, args.output / archive.name)
    shutil.copyfile(sidecar, args.output / sidecar.name)
    run_id, attempt = positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"])
    manifest = {"schema_version": recovery.SCHEMA_VERSION, "kind": "go-sdk-examples", "sdk_formal": header,
                "candidate_run_id": str(run_id), "candidate_run_attempt": attempt, "candidate_mode": args.mode,
                "zip_sha256": examples["sha256"], "examples_receipt": examples}
    gate.write_json(args.output / EXAMPLES_MANIFEST, manifest)
    files = {name: (args.output / name).read_bytes() for name in example_names(version)}
    binding = {"schema_version": recovery.SCHEMA_VERSION, "kind": "sdk-candidate", "repository": gate.REPOSITORY,
               "workflow_path": EXAMPLES_WORKFLOW, "source_sha": header["sdk_source_sha"], "run_id": run_id, "run_attempt": attempt,
               "artifact_name": recovery.candidate_artifact_name(run_id, attempt), "inventory": recovery.inventory_for(files)}
    gate.write_json(args.output / recovery.BINDING_FILENAME, binding)
    checksum_subjects(args.output, example_names(version) | {recovery.BINDING_FILENAME})
    return {"artifact_name": binding["artifact_name"], "candidate_only": True}


def verify_examples(directory, core, recovery, header, run_id, attempt):
    """Verify original examples signature/inventory plus immutable parent SDK identities and actual archive boundary.
    验证原示例签名／清单、不可变父 SDK 身份及实际归档边界。
    """
    names = example_names(header["sdk_version"])
    files = root_files(directory)
    gate.require(set(files) == names | {recovery.BINDING_FILENAME, EXAMPLES_ATTESTATION}, "Original examples has missing/extra subject files")
    normalized = official_subjects(directory, names | {recovery.BINDING_FILENAME}, directory / EXAMPLES_ATTESTATION,
                                   header["sdk_source_sha"], run_id, attempt, workflow=EXAMPLES_WORKFLOW)
    verified = recovery.verify_signed_binding(files[recovery.BINDING_FILENAME], **normalized, repository=gate.REPOSITORY,
                                              workflow_path=EXAMPLES_WORKFLOW, source_sha=header["sdk_source_sha"], run_id=run_id,
                                              run_attempt=attempt, artifact_name=recovery.candidate_artifact_name(run_id, attempt))
    recovery.verify_inventory(verified, files, attestation_filename=EXAMPLES_ATTESTATION)
    manifest = json.loads(files[EXAMPLES_MANIFEST], object_pairs_hook=gate.pairs)
    gate.require(manifest["schema_version"] == recovery.SCHEMA_VERSION and manifest["kind"] == "go-sdk-examples"
                 and manifest["sdk_formal"] == header and manifest["candidate_run_id"] == str(run_id)
                 and type(manifest["candidate_run_attempt"]) is int and manifest["candidate_run_attempt"] == attempt
                 and manifest["candidate_mode"] in ("artifact-only", "publish"), "Original examples SDK/source/attempt identity differs")
    archive = directory / ("luaskills-sdk-go-examples-" + header["sdk_version"] + ".zip")
    sidecar = archive.with_suffix(".zip.sha256")
    gate.require(gate.sha256(archive) == manifest["zip_sha256"]
                 and sidecar.read_text(encoding="utf-8").strip() == manifest["zip_sha256"] + "  " + archive.name,
                 "Signed original examples ZIP/sidecar differs")
    from prepare_examples_release import verify_archive
    verify_archive(archive, "luaskills-sdk-go-examples-" + header["sdk_version"])
    return manifest


def examples_finalize(args):
    """Verify this examples signing job's action bundle and preserve only the frozen root artifact bytes.
    验证本示例签名作业的 action bundle，并仅保留冻结根制品字节。
    """
    header = formal_header(args.formal)
    core, recovery = authorities(args.core_root, header["core_commit"])
    files = {name: (args.directory / name).read_bytes() for name in example_names(header["sdk_version"]) | {recovery.BINDING_FILENAME}}
    files[EXAMPLES_ATTESTATION] = args.bundle.read_bytes()
    save_files(args.output, files)
    verify_examples(args.output, core, recovery, header, positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"]))
    return {"artifact_name": recovery.candidate_artifact_name(positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"]))}


def examples_publish(args):
    """Publish only a specified signed original ZIP/sidecar after current formal SDK reacceptance; never rebuild on recovery.
    在当前正式 SDK 重新验收后，仅发布指定签名原 ZIP／sidecar；恢复时绝不重建。
    """
    header = formal_header(args.formal)
    core, recovery = authorities(args.core_root, header["core_commit"])
    gate.require(args.intent in ("publish", "recover") and os.environ.get("GITHUB_ACTIONS") == "true"
                 and os.environ.get("GITHUB_JOB") == "examples-publish" and os.environ.get("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
                 and os.environ.get("GITHUB_SHA") == os.environ.get("GITHUB_WORKFLOW_SHA") == header["sdk_source_sha"],
                 "Explicit current production examples publication required")
    current_id, current_attempt = positive(os.environ["GITHUB_RUN_ID"]), positive(os.environ["GITHUB_RUN_ATTEMPT"])
    current = gate.GitHub().request(f"actions/runs/{current_id}/attempts/{current_attempt}")
    gate.require(current["id"] == current_id and current["run_attempt"] == current_attempt and current["head_sha"] == header["sdk_source_sha"]
                 and current["repository"]["full_name"] == current["head_repository"]["full_name"] == gate.REPOSITORY
                 and current["event"] == "workflow_dispatch"
                 and current["display_title"] == f"Go examples {header['sdk_source_sha']} mode={args.intent}"
                 and gate.GitHub().request("actions/workflows/" + str(current["workflow_id"]))["path"] == EXAMPLES_WORKFLOW,
                 "Current examples publication dispatch differs")
    run_id, attempt = positive(args.examples_run_id), positive(args.examples_run_attempt)
    metadata, files = recovery.download_artifact(core.Http(), repository=gate.REPOSITORY, source_sha=header["sdk_source_sha"],
                                               run_id=run_id, artifact_id=positive(args.examples_artifact_id),
                                               artifact_name=recovery.candidate_artifact_name(run_id, attempt))
    save_files(args.output, files)
    manifest = verify_examples(args.output, core, recovery, header, run_id, attempt)
    actual = recovery.verify_attempt(core.Http(), repository=gate.REPOSITORY, workflow_path=EXAMPLES_WORKFLOW,
                                     source_sha=header["sdk_source_sha"], run_id=run_id, run_attempt=attempt,
                                     required_jobs=["examples-candidate"], phase="candidate")
    gate.require(actual["attempt"]["event"] == "workflow_dispatch"
                 and actual["attempt"]["display_title"] == f"Go examples {header['sdk_source_sha']} mode={manifest['candidate_mode']}",
                 "Original examples attempt event/mode differs")
    api = gate.GitHub()
    gate.resolve_tag(api, "v" + header["sdk_version"], header["sdk_source_sha"])
    gate.trusted_workflow(api, header["sdk_source_sha"])
    notes = args.output.parent / "examples-notes.md"
    gate.require(not notes.exists(), "Examples publication notes output must be fresh")
    notes.write_text(f"Original signed six-example ZIP and sidecar from {run_id}/{attempt}; formally completed Go SDK {header['candidate_run_id']}/{header['candidate_run_attempt']} and {header['completion_run_id']}/{header['completion_run_attempt']}.\n", encoding="utf-8")
    result = gate.publish_assets(api, "examples-v" + header["sdk_version"], header["sdk_source_sha"],
                                tuple(args.output / name for name in sorted(files)), "LuaSkills Go SDK Examples " + header["sdk_version"], notes)
    return {**result, "original_attempt": actual, "artifact": metadata}
