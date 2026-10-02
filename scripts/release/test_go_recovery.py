"""Exercise real SDK byte bundles and the actual shared recovery authority against mock exact-attempt APIs.
使用真实 SDK 字节 bundle 及实际共同恢复权威，验证模拟精确轮次 API。
"""
import argparse
import copy
import io
import json
import os
from pathlib import Path
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
        """Return exact byte body/header tuple for url; require the documented binary download mode.
        返回 url 的精确字节正文／响应头元组；要求文档规定的二进制下载模式。
        """
        self.calls.append(url)
        if not binary:
            raise AssertionError("Artifact download must be binary")
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


if __name__ == "__main__":
    unittest.main()
