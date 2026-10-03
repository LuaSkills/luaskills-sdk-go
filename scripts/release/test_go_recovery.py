"""Exercise real SDK byte bundles and the actual shared recovery authority against mock exact-attempt APIs.
使用真实 SDK 字节 bundle 及实际共同恢复权威，验证模拟精确轮次 API。
"""
import argparse
import copy
import io
import json
import os
from pathlib import Path
import shutil
import sys
import tarfile
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch
import zipfile

import go_release as gate
import go_recovery as flow
from test_go_release import Api, DraftApi

# Offline tests load the same shared implementation; production only imports its exact committed authority.
# 离线测试加载同一共同实现；生产仅导入其精确已提交权威。
if "LUASKILLS_RELEASE_CORE_ROOT" not in os.environ:
    raise RuntimeError("Set LUASKILLS_RELEASE_CORE_ROOT to the one explicit core authority checkout for offline recovery tests")
CORE_RELEASE = Path(os.environ["LUASKILLS_RELEASE_CORE_ROOT"]).resolve(strict=True) / "scripts/release"
sys.path.insert(0, str(CORE_RELEASE))
import sdk_recovery as common
import sdk_prerequisites as core


class Http:
    """Return explicit fixture API JSON/bytes and record every URL; reject undeclared or latest lookups.
    返回显式夹具 API JSON／字节并记录每个 URL；拒绝未声明或最新查询。
    """

    def __init__(self, routes):
        """Retain exact route values and an initially empty request log.
        保留精确路由值及初始空请求日志。
        """
        self.routes, self.calls = routes, []

    def json(self, url):
        """Return an independent JSON object for exact url; record the actual lookup.
        返回精确 url 的独立 JSON 对象；记录实际查询。
        """
        self.calls.append(url)
        return copy.deepcopy(self.routes[url])

    def get(self, url, binary=False):
        """Return exact ZIP body/header tuple for declared url; require Actions JSON-media delegation.
        为已声明 url 返回精确 ZIP 正文／响应头元组；要求 Actions JSON 媒体委托。
        """
        self.calls.append(url)
        if binary is not False or not url.endswith("/zip"):
            raise AssertionError("Exact Actions ZIP must delegate Core's JSON Accept with raw bytes")
        return self.routes[url], {}


def encode(value):
    """Encode deterministic UTF-8 fixture JSON bytes with no duplicate-key ambiguity.
    编码无重复键歧义的确定性 UTF-8 夹具 JSON 字节。
    """
    return (json.dumps(value, sort_keys=True) + "\n").encode()


def signed_fixture():
    """Return an original flat candidate whose physical bundle/manifest/binding inventory is internally exact.
    返回物理 bundle／清单／绑定清单内部精确一致的原平面候选。
    """
    source, core_commit, run_id, attempt = "a" * 40, "b" * 40, 101, 1
    matrix = {"test-host": ()}
    authority = SimpleNamespace(candidate=SimpleNamespace(PLATFORMS=matrix), rooted_archive=core.rooted_archive)
    plan = {"sdk_sha": source, "sdk_version": "0.5.9", "core_tag": "v0.5.9", "core_commit": core_commit}
    members = {"plan.json": encode(plan), "aggregate.json": encode({"accepted": True, "plan": plan}),
               "module/module.zip": b"actual-module-fixture-bytes", "core/prerequisites.json": encode({"phase": "complete"})}
    buffer = io.BytesIO()
    with tarfile.open(fileobj=buffer, mode="w:gz") as archive:
        for name, body in sorted(members.items()):
            info = tarfile.TarInfo("go-sdk-proof/" + name)
            info.size = len(body)
            archive.addfile(info, io.BytesIO(body))
    manifest = {"schema_version": 2, "kind": "go-sdk-candidate", "sdk_source_sha": source, "sdk_version": "0.5.9",
                "core_tag": "v0.5.9", "core_commit": core_commit, "candidate_run_id": str(run_id), "candidate_run_attempt": attempt,
                "candidate_mode": "artifact-only", "required_jobs": flow.candidate_jobs(authority),
                "members": {name: {"size": len(body), "sha256": flow.digest(body)} for name, body in members.items()}}
    files = {"go-module.zip": members["module/module.zip"], "go-aggregate.json": members["aggregate.json"],
             "go-core-prerequisites.json": members["core/prerequisites.json"], flow.CANDIDATE_BUNDLE: buffer.getvalue(),
             flow.CANDIDATE_MANIFEST: encode(manifest)}
    binding = {"schema_version": 2, "kind": "sdk-candidate", "repository": gate.REPOSITORY, "workflow_path": flow.WORKFLOW,
               "source_sha": source, "run_id": run_id, "run_attempt": attempt,
               "artifact_name": common.candidate_artifact_name(run_id, attempt), "inventory": common.inventory_for(files)}
    files[common.BINDING_FILENAME] = encode(binding)
    normalized = {"verified_subjects": {name: flow.digest(body) for name, body in files.items()},
                  "verified_invocation_uri": common.invocation_uri(gate.REPOSITORY, run_id, attempt), "verified_source_sha": source}
    files[flow.CANDIDATE_ATTESTATION] = b"offline-official-verifier-fixture"
    return authority, source, core_commit, run_id, attempt, manifest, files, normalized


def artifact_routes(authority, source, run_id, attempt, files, *, conclusion="failure"):
    """Create real artifact ZIP and exact original failed-attempt routes with independently successful required jobs.
    创建真实制品 ZIP 及精确原失败轮次路由，其中各必需作业独立成功。
    """
    output = io.BytesIO()
    with zipfile.ZipFile(output, "w") as archive:
        for name, body in files.items():
            archive.writestr(name, body)
    body, artifact_id = output.getvalue(), 777
    base = "https://api.github.com/repos/" + gate.REPOSITORY
    endpoint = f"{base}/actions/runs/{run_id}/attempts/{attempt}"
    actual = {"id": run_id, "run_attempt": attempt, "head_sha": source, "repository": {"full_name": gate.REPOSITORY},
              "head_repository": {"full_name": gate.REPOSITORY}, "path": flow.WORKFLOW, "workflow_id": 1,
              "status": "completed", "conclusion": conclusion, "event": "workflow_dispatch", "display_title": f"Go SDK {source} mode=artifact-only"}
    jobs = [{"id": index + 1, "run_id": run_id, "head_sha": source, "run_url": f"{base}/actions/runs/{run_id}",
             "name": name, "status": "completed", "conclusion": "success"} for index, name in enumerate(flow.candidate_jobs(authority))]
    artifact = {"id": artifact_id, "name": common.candidate_artifact_name(run_id, attempt), "expired": False,
                "workflow_run": {"id": run_id, "head_sha": source}, "size_in_bytes": len(body), "digest": "sha256:" + flow.digest(body)}
    routes = {endpoint: actual, endpoint + "/jobs?per_page=100&page=1": {"total_count": len(jobs), "jobs": jobs},
              f"{base}/actions/artifacts/{artifact_id}": artifact, f"{base}/actions/artifacts/{artifact_id}/zip": body}
    return routes, endpoint, artifact_id


class RecoveryTests(unittest.TestCase):
    """Reject cross-attempt/partial/expired evidence and preserve original failure before any mutation.
    在任何写入前拒绝跨轮次／局部／过期证据，并保留原失败状态。
    """

    def test_exact_old_failed_candidate_and_prewrite_negatives(self):
        """Accept a signed old failed attempt's completed gates, then reject swaps, expiry, partial gates and byte changes.
        接受签名原失败轮次的已完成门禁，随后拒绝替换、过期、局部门禁及字节变化。
        """
        authority, source, commit, run_id, attempt, manifest, files, normalized = signed_fixture()
        routes, endpoint, artifact_id = artifact_routes(authority, source, run_id, attempt, files)
        for mutation in ("valid", "valid-active", "expired", "partial", "cross-attempt", "artifact-source", "byte-mismatch"):
            with self.subTest(mutation=mutation), tempfile.TemporaryDirectory() as temporary:
                changed = copy.deepcopy(routes)
                artifact_url = f"https://api.github.com/repos/{gate.REPOSITORY}/actions/artifacts/{artifact_id}"
                if mutation == "valid-active":
                    changed[endpoint]["status"], changed[endpoint]["conclusion"] = "in_progress", None
                elif mutation == "expired":
                    changed[artifact_url]["expired"] = True
                elif mutation == "partial":
                    changed[endpoint + "/jobs?per_page=100&page=1"]["jobs"][-1]["conclusion"] = "failure"
                elif mutation == "cross-attempt":
                    changed[endpoint]["run_attempt"] = 2
                elif mutation == "artifact-source":
                    changed[artifact_url]["workflow_run"]["head_sha"] = "c" * 40
                elif mutation == "byte-mismatch":
                    changed[artifact_url + "/zip"] += b"changed"
                http = Http(changed)
                authority.Http = lambda: http
                args = argparse.Namespace(core_root=Path("unused"), core_tag="v0.5.9", core_commit=commit, sdk_sha=source,
                                          candidate_run_id=str(run_id), candidate_run_attempt=str(attempt),
                                          candidate_artifact_id=str(artifact_id), output=Path(temporary) / "accepted")
                with patch.object(flow, "authorities", return_value=(authority, common)), \
                     patch.object(flow, "official_subjects", return_value=normalized), \
                     patch.object(flow, "reaggregate_candidate") as aggregate, patch.object(gate, "run") as command:
                    if mutation in ("valid", "valid-active"):
                        result = flow.candidate_consume(args)
                        self.assertEqual(result["attempt"]["attempt"]["conclusion"], None if mutation == "valid-active" else "failure")
                        self.assertEqual(result["manifest"]["candidate_mode"], "artifact-only")
                        self.assertEqual(flow.root_files(args.output / "original"), files)
                        aggregate.assert_called_once()
                    else:
                        with self.assertRaises(ValueError):
                            flow.candidate_consume(args)
                        self.assertFalse((args.output / "candidate.json").exists())
                        aggregate.assert_not_called()
                    command.assert_not_called()
                self.assertNotIn(f"https://api.github.com/repos/{gate.REPOSITORY}/actions/runs/{run_id}", http.calls)

    def test_signature_inventory_and_three_original_hashes(self):
        """Use actual shared binding validation and reject complete-looking rewritten originals or completion links.
        使用实际共同绑定验证，拒绝看似完整但已重写的原件或完成关联。
        """
        authority, source, commit, run_id, attempt, original, files, normalized = signed_fixture()
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary) / "original"
            flow.save_files(directory, files)
            with patch.object(flow, "official_subjects", return_value=normalized):
                facts = flow.verify_candidate(directory, authority, common, source, run_id, attempt)
            for name in ("go-module.zip", flow.CANDIDATE_BUNDLE, flow.CANDIDATE_MANIFEST, common.BINDING_FILENAME):
                changed = copy.deepcopy(files)
                changed[name] += b"changed"
                target = Path(temporary) / name
                flow.save_files(target, changed)
                with patch.object(flow, "official_subjects", return_value=normalized), self.assertRaises(ValueError):
                    flow.verify_candidate(target, authority, common, source, run_id, attempt)
            receipt = {"manifest": original, "binding": facts["binding"], "artifact_id": 777}
            release = {"id": 888, "tag_name": "v0.5.9"}
            completion = {"schema_version": 2, "kind": "go-sdk-completion", "sdk_version": "0.5.9", "core_tag": "v0.5.9", "core_commit": commit,
                          "original": flow.candidate_links(common, receipt, files), "completion_source_sha": source, "completion_run_id": "202",
                          "completion_run_attempt": 2, "completion_intent": "recover",
                          "main_release": {"id": 888, "tag": "v0.5.9", "source_sha": source, "inventory": common.inventory_for(files)}}
            flow.verify_completion_links(completion, receipt, files, release, common, source, 202, 2)
            for field in ("candidate_bundle_sha256", "candidate_manifest_sha256", "candidate_attestation_sha256"):
                changed = copy.deepcopy(completion)
                changed["original"][field] = "0" * 64
                with self.subTest(field=field), self.assertRaisesRegex(ValueError, "original bundle"):
                    flow.verify_completion_links(changed, receipt, files, release, common, source, 202, 2)
            changed = copy.deepcopy(completion)
            changed["completion_source_sha"] = "c" * 40
            with self.assertRaisesRegex(ValueError, "identity differs"):
                flow.verify_completion_links(changed, receipt, files, release, common, source, 202, 2)

    def test_completion_uses_specified_successful_attempt_not_latest(self):
        """Authenticate exact completed completion attempt while an unrelated later retry has failed; no latest lookup.
        在无关后续重试失败时认证精确已完成成功轮次；不查询最新轮次。
        """
        authority, source, _, run_id, attempt, _, files, _ = signed_fixture()
        routes, endpoint, _ = artifact_routes(authority, source, run_id, attempt, files, conclusion="success")
        routes[endpoint]["display_title"] = f"Go SDK {source} mode=recover"
        jobs = routes[endpoint + "/jobs?per_page=100&page=1"]["jobs"]
        jobs[0]["name"] = "completion"
        http = Http(routes)
        authority.Http = lambda: http
        result = flow.completion_attempt(authority, common, source, run_id, attempt)
        self.assertEqual(result["attempt"]["conclusion"], "success")
        self.assertEqual(result["run_attempt"], attempt)
        routes[endpoint]["conclusion"] = "failure"
        with self.assertRaisesRegex(ValueError, "did not succeed"):
            flow.completion_attempt(authority, common, source, run_id, attempt)
        self.assertTrue(all("/attempts/1" in url for url in http.calls))

    def test_publish_success_lost_readback_reuses_exact_final_bytes(self):
        """Lose the post-publication read once and prove retry reuses exact assets without any upload/edit.
        单次丢失公开后回读，证明重试复用精确资产且不再上传／编辑。
        """
        source, tag = "a" * 40, "v0.5.9"
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            file = root / "original.zip"
            file.write_bytes(b"fixed-original")
            notes = root / "notes.md"
            notes.write_text("Original", encoding="utf-8")
            release = {"id": 1, "draft": True, "prerelease": False, "tag_name": tag, "target_commitish": source, "immutable": False}
            routes = {"git/ref/tags/" + tag: {"ref": "refs/tags/" + tag, "object": {"type": "commit", "sha": source}},
                      "git/commits/" + source: {"sha": source}, "releases/1/assets?per_page=100&page=1": [{"name": file.name, "id": 1}],
                      "releases/assets/1": file.read_bytes()}
            api = DraftApi(routes, release)
            original_request = api.request
            lost = False

            def request(path, **kwargs):
                """Drop exactly one final by-ID read after the server changed state; preserve actual final record.
                在服务端状态改变后恰好丢失一次最终按 ID 回读；保留实际 final 记录。
                """
                nonlocal lost
                if path == "releases/1" and api.release["draft"] is False and not lost:
                    lost = True
                    raise OSError("Lost final readback")
                return original_request(path, **kwargs)

            def execute(command, **kwargs):
                """Simulate the actual final publication effect, refusing any asset mutation.
                模拟实际最终公开效果，并拒绝任何资产写入。
                """
                self.assertEqual(command[2], "edit")
                api.release["draft"] = False
                return ""

            with patch.dict(gate.os.environ, {"GITHUB_TOKEN": "fixture", "GITHUB_ACTIONS": "true"}), \
                 patch.object(api, "request", side_effect=request), patch.object(gate, "run", side_effect=execute) as command:
                with self.assertRaisesRegex(OSError, "Lost final"):
                    gate.publish_assets(api, tag, source, (file,), "Test", notes)
                self.assertFalse(api.release["draft"])
                self.assertEqual(command.call_count, 1)
                result = gate.publish_assets(api, tag, source, (file,), "Test", notes)
                self.assertEqual(result["uploaded"], [])
                self.assertEqual(command.call_count, 1)

    def test_signed_examples_retain_original_zip_sidecar_and_dual_identity(self):
        """Stage actual repository example bytes, verify the shared signed inventory, then reject swaps before publication.
        准备真实仓库示例字节、验证共同签名清单，并在发布前拒绝替换。
        """
        from prepare_examples_release import prepare, write_archive
        authority, source, commit, run_id, attempt, _, _, _ = signed_fixture()
        version = (gate.ROOT / "VERSION").read_text(encoding="utf-8").strip()
        header = {"schema_version": 2, "repository": gate.REPOSITORY, "accepted": True, "sdk_source_sha": source,
                  "sdk_version": version, "core_tag": "v0.5.9", "core_commit": commit, "candidate_run_id": "9",
                  "candidate_run_attempt": 1, "completion_run_id": "10", "completion_run_attempt": 2,
                  "completion_source_sha": source}
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            examples = root / "examples"
            examples.mkdir()
            package = examples / "package"
            prepare(gate.ROOT, package, version, "v" + version, "1.2.3")
            archive = examples / ("luaskills-sdk-go-examples-" + version + ".zip")
            write_archive(package, archive, archive.stem)
            sidecar = archive.with_suffix(".zip.sha256")
            sidecar.write_text(gate.sha256(archive) + "  " + archive.name + "\n", encoding="utf-8")
            gate.write_json(examples / "examples.json", {"accepted": True, "plan": {"sdk_sha": source, "sdk_version": version}, "sha256": gate.sha256(archive)})
            args = argparse.Namespace(core_root=Path("unused"), formal=root / "formal", examples=examples, mode="artifact-only", output=root / "stage")
            environment = {"GITHUB_SHA": source, "GITHUB_WORKFLOW_SHA": source, "GITHUB_JOB": "examples-candidate",
                           "GITHUB_RUN_ID": str(run_id), "GITHUB_RUN_ATTEMPT": str(attempt)}
            with patch.object(flow, "formal_header", return_value=header), patch.object(flow, "authorities", return_value=(authority, common)), \
                 patch.dict(os.environ, environment), patch.object(gate, "run") as command:
                flow.examples_stage(args)
                command.assert_not_called()
            files = {name: (args.output / name).read_bytes() for name in flow.example_names(version) | {common.BINDING_FILENAME}}
            normalized = {"verified_subjects": {name: flow.digest(body) for name, body in files.items()},
                          "verified_invocation_uri": common.invocation_uri(gate.REPOSITORY, run_id, attempt), "verified_source_sha": source}
            files[flow.EXAMPLES_ATTESTATION] = b"offline-official-verifier-fixture"
            original = root / "original"
            flow.save_files(original, files)
            with patch.object(flow, "official_subjects", return_value=normalized):
                accepted = flow.verify_examples(original, authority, common, header, run_id, attempt)
            self.assertEqual(accepted["candidate_mode"], "artifact-only")
            self.assertEqual((original / archive.name).read_bytes(), archive.read_bytes())
            self.assertEqual((original / sidecar.name).read_bytes(), sidecar.read_bytes())
            for mutation in ("zip", "sidecar", "extra", "parent-attempt", "original-attempt"):
                with self.subTest(mutation=mutation):
                    changed = copy.deepcopy(files)
                    parent = copy.deepcopy(header)
                    verify_attempt = attempt
                    if mutation == "zip":
                        changed[archive.name] += b"changed"
                    elif mutation == "sidecar":
                        changed[sidecar.name] = b"different sidecar\n"
                    elif mutation == "extra":
                        changed["extra.txt"] = b"extra"
                    elif mutation == "parent-attempt":
                        parent["completion_run_attempt"] += 1
                    else:
                        verify_attempt += 1
                    directory = root / mutation
                    flow.save_files(directory, changed)
                    with patch.object(flow, "official_subjects", return_value=normalized), patch.object(gate, "run") as command, self.assertRaises(ValueError):
                        flow.verify_examples(directory, authority, common, parent, run_id, verify_attempt)
                    command.assert_not_called()


class CoreProofSelectionTests(unittest.TestCase):
    """Verify actual Core byte selection and both production packing call sites without public writes.
    验证实际 Core 字节选择及两个生产打包调用点，不进行公开写入。
    """

    @unittest.skipUnless("SDK_RELEASE_TEST_CORE_PROOF" in os.environ, "Explicit original complete Core proof required")
    def test_audit_retention_missing_required_and_tampered_copy(self):
        """Retain original audit/native evidence and reject missing/tampered inputs through real Core validation.
        通过真实 Core 校验保留原审计、原生证据，并拒绝缺失、篡改输入。
        """
        # Original is read-only; an independent short-path copy avoids Windows hard-link path limits.
        # Original 只读；独立短路径副本避免 Windows 硬链接路径上限。
        original = Path(os.environ["SDK_RELEASE_TEST_CORE_PROOF"]).resolve(strict=True)
        with tempfile.TemporaryDirectory(prefix="cp-") as temporary:
            # Root preserves the actual producer bytes rather than a mirrored synthetic protocol.
            # Root 保留实际生产者字节，不采用镜像合成协议。
            root = Path(temporary) / "core"
            shutil.copytree(original, root, copy_function=shutil.copyfile)
            selected = flow.core_proof_members(root / "prerequisites.json", core)
            self.assertIn("candidate/luaskills-ffi-sdk-windows-x64.tar.gz", selected)
            self.assertNotIn("candidate/luaskills-demo-ffi-windows-x64.tar.gz", selected)
            self.assertNotIn("downloads/assets/luaskills-ffi-sdk-windows-x64.tar.gz", selected)
            for path in (root / "registry").rglob("*"):
                if path.is_file():
                    self.assertEqual(selected[path.relative_to(root).as_posix()].read_bytes(), path.read_bytes())
            # The required official manifest remains mandatory independently of the copied archive selection.
            # 必需正式清单仍为强制项，不因归档副本选择而改变。
            required = root / "downloads/candidate-manifest.json"
            required.unlink()
            with self.assertRaises(FileNotFoundError):
                flow.core_proof_members(root / "prerequisites.json", core)
            shutil.copyfile(original / "downloads/candidate-manifest.json", required)
            # A source copy that will be omitted still needs its original official asset checksum.
            # 即将省略的源码副本仍须符合原正式资产校验和。
            source = core.candidate.read_json(root / "candidate/candidate-manifest.json")["source_archive"]["name"]
            copy = root / "downloads/assets" / source
            copy.unlink()
            copy.write_bytes(b"tampered archive copy")
            with self.assertRaisesRegex(ValueError, "differs from official asset"):
                flow.core_proof_members(root / "prerequisites.json", core)

    def test_candidate_and_completion_propagate_core_selection_failure(self):
        """Execute both actual staging methods and require each to propagate the sole selector failure.
        执行两个实际准备方法，并要求各自传播唯一选择函数失败。
        """
        # Fixture signatures and public processes remain controlled unit-only boundaries.
        # 夹具签名及公开进程边界保持受控、仅限单测。
        authority, source, commit, run_id, attempt, manifest, files, normalized = signed_fixture()
        with tempfile.TemporaryDirectory() as temporary:
            # Root owns only these small orchestration fixtures, not an authenticated publication candidate.
            # Root 仅拥有这些小型编排夹具，不是已认证发布候选。
            root = Path(temporary)
            proof = root / "candidate/proof"
            proof.mkdir(parents=True)
            plan = json.loads(common.sdk_prerequisites.rooted_archive(files[flow.CANDIDATE_BUNDLE], "go-sdk-proof")["plan.json"])
            # This old recovery fixture needs its explicit module tag to reach the completion selector boundary.
            # 此旧恢复夹具须提供显式模块标签，才能执行至完成选择边界。
            plan["module_version"] = "v0.5.9"
            aggregate = {"accepted": True, "plan": plan}
            gate.write_json(proof / "plan.json", plan)
            gate.write_json(proof / "aggregate.json", aggregate)
            (proof / "module").mkdir()
            (proof / "module/module.zip").write_bytes(files["go-module.zip"])
            (proof / "core").mkdir()
            (proof / "core/prerequisites.json").write_bytes(files["go-core-prerequisites.json"])
            reports = root / "reports"
            reports.mkdir()
            original = root / "candidate/original"
            flow.save_files(original, files)
            receipt = {"manifest": manifest, "binding": {}, "physical_inventory": common.inventory_for(files)}
            gate.write_json(root / "candidate/candidate.json", receipt)
            environment = {"GITHUB_SHA": source, "GITHUB_WORKFLOW_SHA": source, "GITHUB_JOB": "candidate-evidence",
                           "GITHUB_RUN_ID": str(run_id), "GITHUB_RUN_ATTEMPT": str(attempt)}
            candidate = argparse.Namespace(core_root=root, plan=proof / "plan.json", aggregate=proof / "aggregate.json",
                module=proof / "module", prerequisites=proof / "core/prerequisites.json", reports=reports,
                mode="artifact-only", output=root / "candidate-stage")
            with patch.object(flow, "authorities", return_value=(authority, common)), patch.object(gate, "module_bytes"), \
                    patch.object(gate, "aggregate", return_value=aggregate), patch.dict(os.environ, environment), \
                    patch.object(flow, "core_proof_members", side_effect=ValueError("Required Core bytes missing")) as select:
                with self.assertRaisesRegex(ValueError, "Required Core bytes missing"):
                    flow.candidate_stage(candidate)
                select.assert_called_once_with(candidate.prerequisites, authority)
            # Every public/process boundary is mocked; completion still reaches its actual selector call site.
            # 所有公开、进程边界均有替身；completion 仍执行至实际选择函数调用点。
            completion = argparse.Namespace(candidate=root / "candidate", core_root=root, intent="recover",
                npm_root=root, pypi_root=root, output=root / "completion-stage")
            with patch.object(flow, "authorities", return_value=(authority, common)), \
                    patch.object(flow, "verify_candidate", return_value={"manifest": manifest, "binding": {}}), \
                    patch.object(flow, "production_context"), patch.object(flow, "reaggregate_candidate"), \
                    patch.object(gate, "prerequisites"), patch.object(gate, "publish_tag"), patch.object(gate, "publish_assets"), \
                    patch.object(gate, "GitHub"), patch.object(flow, "release_files", return_value=({"id": 17}, files)), \
                    patch.object(gate, "public_module"), patch.dict(os.environ, environment), \
                    patch.object(flow, "core_proof_members", side_effect=ValueError("Required Core bytes missing")) as select:
                with self.assertRaisesRegex(ValueError, "Required Core bytes missing"):
                    flow.completion_prepare(completion)
                select.assert_called_once_with(completion.output / "current/core/prerequisites.json", authority)


class ArtifactMediaTests(unittest.TestCase):
    """Exercise the SDK adapter through real Core HTTP Requests and artifact byte validation offline.
    离线通过真实 Core HTTP Request 及制品字节验证测试 SDK 适配器。
    """

    def test_bound_archive_media_and_original_core_byte_guards(self):
        """Prove original 415, exact bound-media repair and retained byte/endpoint guards; return nothing.
        证明原 415、精确绑定媒体修复及字节／端点护栏保留；无返回值。
        """
        # Standard-library fixtures replace only network I/O, never Core get/download or SDK media decisions.
        # 标准库夹具仅替换网络 I/O，绝不替换 Core get/download 或 SDK 媒体决策。
        import hashlib
        import urllib.error
        import zipfile
        # Bind the test to the explicit imported Core authority and this SDK recovery wrapper.
        # 将测试绑定到明确已导入 Core 权威及本 SDK 恢复包装。
        shared, recovery, wrapper = core, common, flow.ArtifactHttp
        # Offline identities are explicit fixture values and make no official signing claim.
        # 离线身份为明确夹具值，不声称官方签名。
        endpoint = "https://api.github.com/repos/test/repo/actions/artifacts/77"
        archive_url = endpoint + "/zip"
        buffer = io.BytesIO()
        with zipfile.ZipFile(buffer, "w") as archive:
            archive.writestr("original.txt", b"original ZIP bytes")
        # Actual ZIP bytes are consumed by the unmodified Core digest/size/unpack implementation.
        # 实际 ZIP 字节由未修改 Core 的摘要／大小／解包实现消费。
        content = buffer.getvalue()
        metadata = {"id": 77, "name": recovery.candidate_artifact_name(123, 1), "expired": False,
                    "workflow_run": {"id": 123, "head_sha": "a" * 40}, "size_in_bytes": len(content),
                    "digest": "sha256:" + hashlib.sha256(content).hexdigest()}
        arguments = dict(repository="test/repo", source_sha="a" * 40, run_id=123, artifact_id=77,
                         artifact_name=metadata["name"])
        # Exact unrelated routes expose accidental broader binary rewrites without probing a real network.
        # 精确无关路由暴露意外宽泛二进制改写，不探测真实网络。
        others = ("https://api.github.com/repos/test/repo/actions/artifacts/78/zip",
                  archive_url + "?part=1", "https://example.invalid/native.dll")
        routes = {archive_url: content, **{url: b"unrelated bytes" for url in others}}
        requests, reads = [], []

        class Response(io.BytesIO):
            """Retain byte-body fixture semantics while observing the real Core bounded read.
            保留字节正文夹具语义，同时观察真实 Core 有界读取。
            """

            def read(self, size=-1):
                """Record requested size and return original fixture bytes without altering the body.
                记录请求 size 并返回原夹具字节，不改变正文。
                """
                reads.append(size)
                return super().read(size)

        class Opener:
            """Serve declared routes and reject the bound ZIP's octet Accept with the original 415.
            提供已声明路由，并对绑定 ZIP 的 octet Accept 返回原 415。
            """

            def open(self, request, timeout):
                """Observe the real Request and timeout; return body or the explicit media rejection.
                观察真实 Request 及 timeout；返回正文或明确媒体拒绝。
                """
                # Store only safe URL/media/timeout facts, never the Request's authorization header.
                # 仅保存安全 URL／媒体／超时事实，绝不保存 Request 的授权头。
                accept = request.get_header("Accept")
                requests.append((request.full_url, accept, timeout))
                if request.full_url == archive_url and accept != "application/json":
                    raise urllib.error.HTTPError(request.full_url, 415, "Unsupported Accept", {}, io.BytesIO())
                # Metadata JSON encoding belongs only to its declared API route, not the ZIP body.
                # 元数据 JSON 编码仅属于已声明 API 路由，不属于 ZIP 正文。
                body = json.dumps(metadata).encode() if request.full_url == endpoint else routes[request.full_url]
                response = Response(body)
                response.status = 200
                response.headers = {"Content-Type": "application/json" if request.full_url == endpoint else "application/zip"}
                return response

        # The same real Core instance retains its original get/json implementations and bounded reads.
        # 同一真实 Core 实例保留原 get/json 实现及有界读取。
        http = shared.Http()
        opener = Opener()
        http.opener = opener
        with self.assertRaisesRegex(ValueError, "status 415"):
            recovery.download_artifact(http, **arguments)
        self.assertEqual(requests[-1], (archive_url, "application/octet-stream", 60))
        # Only the SDK wrapper changes bound media; Core still authenticates the ZIP digest/size and members.
        # 仅 SDK 包装改变绑定媒体；Core 仍认证 ZIP 摘要／大小及成员。
        adapted = wrapper(http, "test/repo", 77)
        evidence, files = recovery.download_artifact(adapted, **arguments)
        self.assertEqual(files, {"original.txt": b"original ZIP bytes"})
        self.assertEqual(evidence["archive_sha256"], hashlib.sha256(content).hexdigest())
        self.assertEqual(requests[-1], (archive_url, "application/json", 60))
        self.assertIs(http.opener, opener)
        self.assertEqual(adapted.json(endpoint), metadata)
        self.assertEqual(requests[-1], (endpoint, "application/json", 60))
        for url in others:
            with self.subTest(url=url):
                self.assertEqual(adapted.get(url, binary=True)[0], b"unrelated bytes")
                self.assertEqual(requests[-1], (url, "application/octet-stream", 60))
                self.assertEqual(adapted.get(url, binary=False)[0], b"unrelated bytes")
                self.assertEqual(requests[-1], (url, "application/json", 60))
        self.assertEqual(adapted.get(archive_url, binary=False)[0], content)
        # Wrong API size or digest must still fail inside real Core before its artifact members are accepted.
        # 错误 API 大小或摘要仍须在真实 Core 内失败，之后才可能接受制品成员。
        for field, changed in (("size_in_bytes", len(content) + 1), ("digest", "sha256:" + "0" * 64)):
            with self.subTest(field=field):
                original = metadata[field]
                metadata[field] = changed
                with self.assertRaisesRegex(ValueError, "downloaded bytes differ"):
                    recovery.download_artifact(adapted, **arguments)
                metadata[field] = original
        # Even correctly sized/hashed JSON cannot replace ZIP: real Core's ZIP parser rejects it.
        # 即使大小／摘要正确，JSON 也不能替代 ZIP：真实 Core ZIP 解析器拒绝它。
        routes[archive_url] = b"{}"
        metadata["size_in_bytes"] = len(routes[archive_url])
        metadata["digest"] = "sha256:" + hashlib.sha256(routes[archive_url]).hexdigest()
        with self.assertRaises((ValueError, zipfile.BadZipFile)):
            recovery.download_artifact(adapted, **arguments)
        self.assertTrue(reads)
        self.assertTrue(all(size == shared.MAX_BODY_BYTES + 1 for size in reads))


if __name__ == "__main__":
    unittest.main()
