"""Exercise real frozen ZIPs and mocked publication APIs without remote writes or Cargo.
无需远端写入或 Cargo，验证真实冻结 ZIP 与模拟发布 API。
"""
import copy
import hashlib
import io
import json
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch
import urllib.error
import zipfile

import go_release as gate


class Api:
    """Provide a deterministic read-only API fixture with explicit routes.
    提供具有显式路由的确定性只读 API 夹具。
    """

    def __init__(self, routes):
        """Retain exact routes; every unexpected lookup fails.
        保留精确 routes；任何非预期查询均失败。
        """
        self.routes = routes

    def request(self, path, **kwargs):
        """Return a deep copy of one exact fixture route.
        返回一个精确夹具路由的深拷贝。
        """
        return copy.deepcopy(self.routes[path])


class DraftApi(Api):
    """Model authenticated draft listing/creation while published-by-tag correctly returns 404 for drafts.
    模拟认证草稿列表／创建，且已发布 Release 按 tag 查询对草稿正确返回 404。
    """

    def __init__(self, routes, release):
        """Retain routes and optional exact release record; record created request bodies.
        保留 routes 及可选精确 Release 记录；记录创建请求 body。
        """
        super().__init__(routes)
        self.release = release
        self.created = []

    def request(self, path, **kwargs):
        """Return authenticated records or create one draft; reject draft lookup through published-by-tag.
        返回认证记录或创建一个草稿；拒绝通过已发布按 tag 端点查询草稿。
        """
        if path.startswith("releases/tags/"):
            if self.release is None or self.release["draft"]:
                raise urllib.error.HTTPError(path, 404, "Only published releases", {}, None)
            return copy.deepcopy(self.release)
        if path == "releases?per_page=100&page=1":
            return [] if self.release is None else [copy.deepcopy(self.release)]
        if path == "releases":
            body = kwargs["body"]
            self.created.append(copy.deepcopy(body))
            self.release = {"id": 1, "immutable": False, **body}
            return copy.deepcopy(self.release)
        if path == "releases/1":
            return copy.deepcopy(self.release)
        return super().request(path, **kwargs)


def fixture():
    """Return small synthetic authority/plan/receipt fixtures without reproducing release platform tables.
    返回小型合成权威／计划／凭证夹具，不复写发布平台表。
    """
    authority = SimpleNamespace(candidate=SimpleNamespace(MANIFEST_VERSION=1, PLATFORMS={"test-a": (), "test-b": ()}),
                                REPOSITORY="LuaSkills/luaskills", REGISTRY_SOURCE="registry+https://github.com/rust-lang/crates.io-index")
    plan = {"sdk_sha": "a" * 40, "sdk_version": "0.5.9", "module_version": "v0.5.9", "module_zip_sha256": "b" * 64,
            "module_members": 1, "core_commit": "c" * 40, "core_tag": "v0.5.9", "npm_version": "0.6.1", "npm_sha": "d" * 40,
            "pypi_version": "0.7.2", "pypi_sha": "e" * 40}
    for kind, candidate_run, completion_run in (("npm", "123", "321"), ("pypi", "456", "654")):
        plan.update({kind + "_candidate_run_id": candidate_run, kind + "_candidate_run_attempt": 1,
                     kind + "_completion_run_id": completion_run, kind + "_completion_run_attempt": 2,
                     kind + "_completion_source_sha": plan[kind + "_sha"]})
    consumer = {"source": authority.REGISTRY_SOURCE, "executable": "/new/luaskills-registry-consumer", "executable_sha256": "f" * 64,
                "commands": [{"command": ["cargo", command], "exit_code": 0} for command in ("generate-lockfile", "metadata", "build")]
                            + [{"command": ["/tmp/luaskills-registry-consumer", "challenge"], "exit_code": 0}],
                "result": {"runtime": True, "pool_reuse": True, "drained": True, "capability_calls": 2}}
    inputs = {"library_sha256": "1" * 64, "description_sha256": "2" * 64, "archive_sha256": "3" * 64, "build": {"identity": "test"}}
    proof = {"schema_version": 1, "phase": "complete", "complete": True, "core_commit": plan["core_commit"], "core_tag": plan["core_tag"],
             "github": {"repository": authority.REPOSITORY}, "registry": {"consumer": consumer},
             "sdk_inputs": {key: copy.deepcopy(inputs) for key in authority.candidate.PLATFORMS}}
    records = [{"schema_version": 1, **{key: plan[key] for key in ("sdk_sha", "sdk_version", "module_version", "module_zip_sha256", "core_commit", "core_tag")},
                "platform": name, "accepted": True, "inputs": copy.deepcopy(inputs),
                "facts": {"race": True, "replace": False, "tests": {"pass": 2, "fail": 0, "skip": 0}, "lifecycle_tests": {"pass": 1, "fail": 0, "skip": 0}}}
               for name in authority.candidate.PLATFORMS]
    return authority, plan, proof, records


class ReleaseTests(unittest.TestCase):
    """Test actual artifact boundaries and immutable publication acceptance failures.
    测试实际产物边界及不可变发布验收失败。
    """

    def test_run_real_failure_exposes_merged_bytes(self):
        """Run a real failed child and require its exact merged bytes, command and exit code.
        运行真实失败子进程，并要求其精确合并字节、命令与退出码。
        Self owns this test; no value is returned and no remote operation occurs.
        Self 持有本测试；无返回值，不执行远端操作。
        """
        # Command emits ordered raw streams without decoding, credentials or network activity.
        # Command 发出有序原始双流，不解码、不使用凭据或网络。
        command = [gate.sys.executable, "-c", 'import os; os.write(1, b"prerequisite stdout\\n"); os.write(2, b"prerequisite stderr\\xff\\n"); raise SystemExit(19)']
        # Displayed receives the real binary diagnostic stream exactly once.
        # Displayed 精确接收一次真实二进制诊断流。
        displayed = io.BytesIO()
        with patch.object(gate.sys, "stderr", SimpleNamespace(buffer=displayed)):
            with self.assertRaises(gate.subprocess.CalledProcessError) as failure:
                gate.run(command)
        self.assertEqual(failure.exception.returncode, 19)
        self.assertEqual(failure.exception.cmd, command)
        self.assertEqual(failure.exception.output, b"prerequisite stdout\nprerequisite stderr\xff\n")
        self.assertEqual(displayed.getvalue(), failure.exception.output)

    def test_run_real_timeout_exposes_partial_bytes(self):
        """Expire a real child after output and retain its partial bytes and original timeout exception.
        在真实子进程输出后使其超时，保留部分字节及原超时异常。
        Self owns this bounded test; no value is returned and production's deadline stays unchanged.
        Self 持有本有界测试；无返回值，生产期限保持不变。
        """
        # Command flushes raw bytes before waiting so expiration has an original diagnostic receipt.
        # Command 在等待前写出原始字节，使超时具有原始诊断凭证。
        command = [gate.sys.executable, "-c", 'import os, time; os.write(1, b"partial stdout\\xff\\n"); time.sleep(30)']
        # Execute retains the real subprocess implementation while shortening only this test's deadline.
        # Execute 保留真实子进程实现，仅缩短本测试期限。
        execute = gate.subprocess.run
        # Displayed receives the exception's actual binary output without replacement.
        # Displayed 接收异常的实际二进制输出，不替换字节。
        displayed = io.BytesIO()

        def expire(arguments, **options):
            """Execute arguments with a short test deadline after verifying production's exact timeout.
            核实生产精确超时后，以短测试期限执行 arguments。
            Options retain all original guards; the real child must raise TimeoutExpired.
            Options 保留全部原护栏；真实子进程必须抛出 TimeoutExpired。
            """
            self.assertEqual(options["timeout"], 1800)
            options["timeout"] = 1
            return execute(arguments, **options)

        with patch.object(gate.subprocess, "run", side_effect=expire), \
                patch.object(gate.sys, "stderr", SimpleNamespace(buffer=displayed)):
            with self.assertRaises(gate.subprocess.TimeoutExpired) as failure:
                gate.run(command)
        self.assertEqual(failure.exception.cmd, command)
        self.assertEqual(failure.exception.timeout, 1)
        self.assertEqual(failure.exception.output, b"partial stdout\xff\n")
        self.assertEqual(displayed.getvalue(), failure.exception.output)

    def test_real_frozen_zip_and_modified_member(self):
        """Freeze the actual SDK ZIP, then reject a real rewritten member against its original manifest.
        冻结实际 SDK ZIP，再拒绝真实改写成员与原始清单的差异。
        """
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary) / "module"
            frozen = gate.freeze_module(gate.ROOT, directory)
            plan = {"module_version": frozen["module_version"], "module_zip_sha256": frozen["module_zip_sha256"], "module_members": frozen["module_members"]}
            actual = gate.module_bytes(plan, directory)
            self.assertIn("examples/embedded_lifecycle/main_native_test.go", actual)
            archive = directory / "module.zip"
            with zipfile.ZipFile(archive) as package:
                entries = [(member, package.read(member)) for member in package.infolist()]
            with zipfile.ZipFile(archive, "w") as package:
                for member, body in entries:
                    package.writestr(member, body + b"tampered" if member.filename.endswith("/embedded_client.go") else body)
            plan["module_zip_sha256"] = gate.sha256(archive)
            with self.assertRaisesRegex(ValueError, "member manifest mismatch"):
                gate.module_bytes(plan, directory)

    def test_version_mirror_and_default_tag(self):
        """Read real metadata files and reject mismatched embedded mirror or explicit default asset tag.
        读取真实元数据文件，拒绝嵌入镜像或显式默认资产标签不符。
        """
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root / "VERSION").write_text("0.5.9\n")
            (root / "embedded_contract_generated.go").write_text('const EmbeddedCoreVersion = "0.5.9"\n')
            (root / "runtime_assets.go").write_text('const DefaultLuaSkillsVersion = "v0.5.9"\n')
            self.assertEqual(gate.source_metadata(root, "v0.5.9"), "0.5.9")
            with self.assertRaisesRegex(ValueError, "asset tag"):
                gate.source_metadata(root, "v0.6.0")
            (root / "embedded_contract_generated.go").write_text('const EmbeddedCoreVersion = "0.5.8"\n')
            with self.assertRaisesRegex(ValueError, "mirror differ"):
                gate.source_metadata(root, "v0.5.9")

    def test_missing_platform_sha_and_nonformal_cargo(self):
        """Reject omitted native platforms, swapped SDK hashes, nonformal phases and fake Cargo receipts.
        拒绝遗漏原生平台、替换 SDK 摘要、非正式阶段及伪 Cargo 凭证。
        """
        authority, plan, proof, records = fixture()
        self.assertTrue(gate.aggregate_records(plan, proof, records, authority)["accepted"])
        mutations = [("missing", lambda p, r: r.pop()),
                     ("sha", lambda p, r: r[0].update(sdk_sha="9" * 40)),
                     ("nonformal", lambda p, r: p.update(complete=False, phase="github-only")),
                     ("fakecargo", lambda p, r: p["registry"]["consumer"]["commands"][0].update(command=["echo", "success"])),
                     ("typedclosed", lambda p, r: p["registry"]["consumer"]["result"].update(drained="true"))]
        for name, mutate in mutations:
            with self.subTest(name=name):
                changed, rows = copy.deepcopy(proof), copy.deepcopy(records)
                mutate(changed, rows)
                with self.assertRaises(ValueError):
                    gate.aggregate_records(plan, changed, rows, authority)

    def test_native_actual_closed_receipt(self):
        """Require raw native description and cancelled-admission cleanup passes, plus actual closed example output.
        要求原始原生描述及入场后取消清理通过，以及实际关闭示例输出。
        """
        _, plan, proof, _ = fixture()
        inputs = proof["sdk_inputs"]["test-a"]
        summary = {"mode": "frozen-module-zip", "replace": False, "race": True, "module_zip_sha256": plan["module_zip_sha256"],
                   "module_version": plan["module_version"], "module_members": 1, "sha256": inputs["library_sha256"],
                   "description_sha256": inputs["description_sha256"], "module": "/cache/module", "gomodcache": "/cache", "module_sum": "h1:test",
                   "tests": {"pass": 2, "fail": 0, "skip": 0}, "lifecycle_tests": {"pass": 1, "fail": 0, "skip": 0}}
        def events(names):
            """Encode successful fixture test2json events for the named mandatory receipts.
            为指定必需凭证编码成功夹具 test2json 事件。
            """
            return b"\n".join(json.dumps({"Test": name, "Action": "pass"}).encode() for name in names)
        files = {"tests.jsonl": events(("TestEmbeddedCandidateEvidence", "TestEmbeddedCandidateDescriptionExtensions")),
                 "example-tests.jsonl": events(("TestEmbeddedReserveCancelledObserver",)),
                 "example.log": b"Runtime scope, callback pump, driver receipts and transport closed."}
        gate.validate_native(summary, plan, inputs, files)
        files["example.log"] = b"closed waiter cancelled"
        with self.assertRaisesRegex(ValueError, "Actual closed"):
            gate.validate_native(summary, plan, inputs, files)
        files["example-tests.jsonl"] = events(("TestSomeOtherCase",))
        with self.assertRaisesRegex(ValueError, "receipt missing"):
            gate.validate_native(summary, plan, inputs, files)

    def test_native_real_failure_preserves_raw_output(self):
        """Run a real failing Python child through native; retain exact merged bytes and the original exit code.
        通过 native 运行真实失败 Python 子进程；保留精确合并字节及原退出码。
        Self is this test case; no value is returned and no successful receipt may be created.
        Self 是本测试用例；无返回值，不得创建成功凭证。
        """
        # Temporary owns the real child source and fresh diagnostic output without touching SDK inputs.
        # Temporary 拥有真实子进程源码与新诊断输出，不触碰 SDK 输入。
        with tempfile.TemporaryDirectory() as temporary:
            # Root supplies a real verify script at the exact existing native call path.
            # Root 在原 native 精确调用路径提供真实 verify 脚本。
            root = Path(temporary)
            (root / "scripts").mkdir()
            # Raw contains ordered stdout and stderr bytes, including a byte invalid in UTF-8.
            # Raw 包含有序 stdout 与 stderr 字节，包括一个非法 UTF-8 字节。
            raw = b"native stdout\n" + b"native stderr\xff\n"
            (root / "scripts/verify_embedded_candidate.py").write_text(
                '"""Emit real ordered failure output; return no value and exit with status 17.\n'
                '输出真实有序失败内容；无返回值，以状态 17 退出。\n"""\n'
                'import os\n'
                'os.write(1, b"native stdout\\n")\n'
                'os.write(2, b"native stderr\\xff\\n")\n'
                'raise SystemExit(17)\n', encoding="utf-8")
            # Plan and inputs model only pre-child publication authority; the child itself is never mocked.
            # Plan 与 inputs 仅模拟子进程前发布权威；子进程本身绝不模拟。
            plan = {"core_commit": "c" * 40, "prerequisites_sha256": "f" * 64,
                    "module_zip_sha256": "b" * 64, "module_version": "v0.6.1"}
            inputs = {"library": root / "libluaskills", "library_sha256": "1" * 64,
                      "description": root / "description.json"}
            # Host_os and architecture match this actual test host for the unchanged native platform guard.
            # Host_os 与 architecture 匹配实际测试宿主，供未改 native 平台护栏使用。
            host_os = {"win32": "windows", "darwin": "macos", "linux": "linux"}[gate.sys.platform]
            architecture = "x86_64" if gate.platform.machine().lower() in {"x86_64", "amd64"} else "aarch64"

            def resolve_inputs(prerequisites, selected_platform):
                """Return fixture inputs for the exact pre-child platform; no real Core operation is performed.
                返回精确子进程前平台的夹具输入；不执行真实 Core 操作。
                Prerequisites names the fixture receipt; selected_platform must be test-host.
                Prerequisites 指定夹具凭证；selected_platform 必须为 test-host。
                """
                self.assertEqual(selected_platform, "test-host")
                return inputs

            # Authority provides only the pre-child platform and library-input lookup.
            # Authority 仅提供子进程前平台与库输入查询。
            authority = SimpleNamespace(candidate=SimpleNamespace(PLATFORMS={"test-host": ("test", host_os, architecture)}),
                                        resolve_sdk_inputs=resolve_inputs)
            # Arguments names the fresh output used by the production native boundary.
            # Arguments 指定生产 native 边界使用的新输出。
            arguments = SimpleNamespace(plan=root / "plan.json", core_root=root / "core", platform="test-host",
                                        prerequisites=root / "prerequisites.json", module=root / "module", output=root / "native")
            gate.write_json(arguments.plan, plan)
            gate.write_json(arguments.prerequisites, {})
            # Displayed captures the real binary stderr display without decoding or replacing bytes.
            # Displayed 捕获真实二进制 stderr 展示，不解码或替换字节。
            displayed = io.BytesIO()
            with patch.object(gate, "ROOT", root), patch.object(gate, "core_authority", return_value=authority), \
                    patch.object(gate, "host_platform", return_value="test-host"), patch.object(gate, "complete_core"), \
                    patch.object(gate, "sha256", return_value="f" * 64), patch.object(gate, "module_bytes"), \
                    patch.object(gate.sys, "stderr", SimpleNamespace(buffer=displayed)):
                # Failure is the real CalledProcessError emitted by the unchanged subprocess runner.
                # Failure 是未改子进程运行器发出的真实 CalledProcessError。
                with self.assertRaises(gate.subprocess.CalledProcessError) as failure:
                    gate.native(arguments)
            self.assertEqual(failure.exception.returncode, 17)
            self.assertEqual(failure.exception.output, raw)
            self.assertEqual((arguments.output / "native.log").read_bytes(), raw)
            self.assertEqual(displayed.getvalue(), raw)
            self.assertFalse((arguments.output / "report.json").exists())

    def test_independent_formal_sdk_header(self):
        """Accept different exact TS/Python versions and reject candidate receipts or wrong source/core identity.
        接受不同精确 TS／Python 版本，拒绝候选凭证或错误源码／核心身份。
        """
        _, plan, _, _ = fixture()
        header = {"schema_version": 2, "sdk_source_sha": plan["npm_sha"], "sdk_version": plan["npm_version"],
                  "core_tag": plan["core_tag"], "core_commit": plan["core_commit"],
                  "candidate_run_id": plan["npm_candidate_run_id"], "candidate_run_attempt": 1,
                  "completion_run_id": plan["npm_completion_run_id"], "completion_run_attempt": 2,
                  "completion_source_sha": plan["npm_completion_source_sha"],
                  "repository": gate.NPM_REPOSITORY, "accepted": True, "registry_consumer_file": "fresh-formal-consumer.json", "registry_consumer_sha256": "1" * 64}
        gate.validate_sdk_header(header, plan, "npm", gate.NPM_REPOSITORY)
        for field, value in (("accepted", "true"), ("sdk_version", plan["sdk_version"]), ("sdk_source_sha", "0" * 40), ("core_commit", "0" * 40),
                             ("candidate_run_attempt", 2), ("completion_run_attempt", 1), ("completion_run_id", "123"), ("completion_source_sha", "0" * 40), ("schema_version", 1)):
            with self.subTest(field=field), self.assertRaises(ValueError):
                gate.validate_sdk_header({**header, field: value}, plan, "npm", gate.NPM_REPOSITORY)
        for kind in ("npm", "pypi"):
            flags = gate.sdk_identity_flags(plan, kind)
            self.assertNotIn("--run-id", flags)
            self.assertIn("--candidate-run-id", flags)
            self.assertIn("--candidate-run-attempt", flags)
            self.assertIn("--completion-run-id", flags)
            self.assertIn("--completion-run-attempt", flags)
            self.assertIn("--completion-source-sha", flags)

    def test_recursive_tag_and_workflow_permission_gate(self):
        """Authenticate annotated tags and default-branch workflow trees; reject different commits or new workflows.
        认证附注标签及默认分支工作流树；拒绝不同提交或新增工作流。
        """
        commit, annotation, default = "a" * 40, "b" * 40, "c" * 40
        workflows = [{"path": ".github/workflows/" + name, "sha": "d" * 40} for name in ("sdk-release.yml", "examples-release.yml")]
        api = Api({"git/ref/tags/v0.5.9": {"ref": "refs/tags/v0.5.9", "object": {"type": "tag", "sha": annotation}},
                   "git/tags/" + annotation: {"sha": annotation, "object": {"type": "commit", "sha": commit}},
                   "git/commits/" + commit: {"sha": commit}, "": {"full_name": gate.REPOSITORY, "default_branch": "main"},
                   "branches/main": {"commit": {"sha": default}}, "compare/" + commit + "..." + default: {"merge_base_commit": {"sha": commit}},
                   "git/trees/" + commit + "?recursive=1": {"truncated": False, "tree": workflows},
                   "git/trees/" + default + "?recursive=1": {"truncated": False, "tree": workflows}})
        self.assertEqual(gate.resolve_tag(api, "v0.5.9", commit)[-1]["sha"], commit)
        gate.trusted_workflow(api, commit)
        with self.assertRaisesRegex(ValueError, "different commit"):
            gate.resolve_tag(api, "v0.5.9", default)
        api.routes["git/trees/" + default + "?recursive=1"]["tree"] = workflows[:1]
        with self.assertRaisesRegex(ValueError, "default branch"):
            gate.trusted_workflow(api, commit)

    def test_immutable_asset_conflicts_before_upload(self):
        """Reuse identical real asset bytes and reject any differing existing asset before uploading missing ones.
        复用相同实际资产字节，并在上传缺失资产前拒绝任何不同已有资产。
        """
        with tempfile.TemporaryDirectory() as temporary:
            first, second = Path(temporary) / "first.zip", Path(temporary) / "second.sha256"
            first.write_bytes(b"zip-bytes")
            second.write_bytes(b"sha-bytes")
            assets = [{"name": first.name, "id": 1}]
            self.assertEqual(gate.assets_to_upload(assets, (first, second), lambda entry: b"zip-bytes"), [second])
            with self.assertRaisesRegex(ValueError, "asset differs"):
                gate.assets_to_upload(assets, (second, first), lambda entry: b"changed")

    def test_exact_workflow_sha_before_any_tool(self):
        """Reject SDK/workflow SHA mismatch before source reads, Cargo or remote lookup.
        在源码读取、Cargo 或远端查询前拒绝 SDK／工作流 SHA 不一致。
        """
        args = SimpleNamespace(root=gate.ROOT, sdk_sha="a" * 40, workflow_sha="b" * 40)
        with patch.dict(gate.os.environ, {}, clear=True), patch.object(gate, "run") as command:
            with self.assertRaisesRegex(ValueError, "SHA mismatch"):
                gate.freeze(args)
            command.assert_not_called()


    def test_signed_certificate_and_complete_subject_inventory(self):
        """Reject wrong signed run/attempt, workflow, source or subject even when other provenance claims match.
        即使其他 provenance 声明匹配，也拒绝错误签名 run／attempt、工作流、源码或 subject。
        """
        commit, subjects = "a" * 40, {"module.zip": "b" * 64, "aggregate.json": "c" * 64}
        invocation = "https://github.com/" + gate.REPOSITORY + "/actions/runs/123/attempts/2"
        verified = {"signature": {"certificate": {"issuer": "https://token.actions.githubusercontent.com",
                    "sourceRepositoryURI": "https://github.com/" + gate.REPOSITORY, "sourceRepositoryDigest": commit,
                    "buildSignerDigest": commit, "buildSignerURI": "https://github.com/" + gate.REPOSITORY + "/.github/workflows/sdk-release.yml@refs/heads/main",
                    "runnerEnvironment": "github-hosted", "runInvocationURI": invocation}},
                    "statement": {"predicateType": "https://slsa.dev/provenance/v1", "subject": [{"name": name, "digest": {"sha256": digest}} for name, digest in subjects.items()],
                                  "predicate": {"runDetails": {"metadata": {"invocationId": invocation}}}}}
        gate.validate_attestation_results([{"verificationResult": verified}], subjects, commit, "123", "2")
        for key, value in (("runInvocationURI", invocation.replace("attempts/2", "attempts/1")),
                           ("sourceRepositoryDigest", "d" * 40), ("buildSignerDigest", "d" * 40),
                           ("buildSignerURI", "https://github.com/" + gate.REPOSITORY + "/.github/workflows/other.yml@refs/heads/main"),
                           ("runnerEnvironment", "self-hosted")):
            changed = copy.deepcopy(verified)
            changed["signature"]["certificate"][key] = value
            with self.subTest(key=key), self.assertRaises(ValueError):
                gate.validate_attestation_results([{"verificationResult": changed}], subjects, commit, "123", "2")
        changed = copy.deepcopy(verified)
        changed["statement"]["subject"].pop()
        with self.assertRaisesRegex(ValueError, "inventory mismatch"):
            gate.validate_attestation_results([{"verificationResult": changed}], subjects, commit, "123", "2")
        changed = copy.deepcopy(verified)
        changed["statement"]["predicate"]["runDetails"]["metadata"]["invocationId"] = invocation.replace("runs/123", "runs/456")
        with self.assertRaisesRegex(ValueError, "invocation identity"):
            gate.validate_attestation_results([{"verificationResult": changed}], subjects, commit, "123", "2")

    def test_complete_draft_before_publish_and_no_final_append(self):
        """Upload exact bytes to a draft before publish, and refuse missing assets on an existing final release.
        发布前向 draft 上传精确字节，并拒绝已有 final Release 的缺失资产。
        """
        commit = "a" * 40
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            files = (directory / "module.zip", directory / "proof.json")
            notes = directory / "notes.md"
            notes.write_text("Actual frozen batch", encoding="utf-8")
            for index, filename in enumerate(files):
                filename.write_bytes(str(index).encode())
            for tag in ("v0.5.9", "examples-v0.5.9"):
                for resume in (False, True):
                    with self.subTest(tag=tag, resume=resume):
                        routes = {"git/ref/tags/" + tag: {"ref": "refs/tags/" + tag, "object": {"type": "commit", "sha": commit}},
                                  "git/commits/" + commit: {"sha": commit}, "releases/1/assets?per_page=100&page=1": []}
                        release = {"id": 1, "draft": True, "prerelease": False, "tag_name": tag,
                                   "target_commitish": commit, "immutable": False}
                        api, operations = DraftApi(routes, release if resume else None), []
                        with self.assertRaises(urllib.error.HTTPError) as missing:
                            api.request("releases/tags/" + tag)
                        self.assertEqual(missing.exception.code, 404)
                        missing.exception.close()
                        if resume:
                            routes["releases/1/assets?per_page=100&page=1"] = [{"id": 1, "name": files[0].name}]
                            routes["releases/assets/1"] = files[0].read_bytes()

                        def execute(command, **kwargs):
                            """Record simulated uploads/edit and require complete draft bytes before publication.
                            记录模拟上传／编辑，并要求发布前草稿字节完整。
                            """
                            operation = command[2]
                            operations.append(operation)
                            if operation == "upload":
                                filename = Path(command[4])
                                asset_id = len(routes["releases/1/assets?per_page=100&page=1"]) + 1
                                routes["releases/1/assets?per_page=100&page=1"].append({"name": filename.name, "id": asset_id})
                                routes["releases/assets/" + str(asset_id)] = filename.read_bytes()
                            elif operation == "edit":
                                self.assertEqual(len(routes["releases/1/assets?per_page=100&page=1"]), len(files))
                                api.release["draft"] = False
                            return ""

                        with patch.dict(gate.os.environ, {"GITHUB_TOKEN": "test-only", "GITHUB_ACTIONS": "true"}), patch.object(gate, "run", side_effect=execute):
                            result = gate.publish_assets(api, tag, commit, files, "Test", notes)
                        self.assertEqual(operations, (["upload"] if resume else ["upload", "upload"]) + ["edit"])
                        self.assertEqual(result["created"], not resume)
                        self.assertEqual(len(api.created), int(not resume))
                        self.assertFalse(result["server_immutable"])
                        self.assertFalse(api.request("releases/tags/" + tag)["draft"])
                        routes["releases/1/assets?per_page=100&page=1"].pop()
                        with patch.dict(gate.os.environ, {"GITHUB_TOKEN": "test-only", "GITHUB_ACTIONS": "true"}), patch.object(gate, "run") as process:
                            with self.assertRaisesRegex(ValueError, "cannot receive additional"):
                                gate.publish_assets(api, tag, commit, files, "Test", notes)
                            process.assert_not_called()

    def test_draft_lookup_requires_all_pages_and_unique_id(self):
        """Discover drafts beyond the first page and reject duplicate tags or ID substitutions before writes.
        发现第一页之后的草稿，并在写入前拒绝重复 tag 或 ID 替换。
        """
        tag = "v0.5.9"
        selected = {"id": 101, "tag_name": tag, "draft": True}
        first = [{"id": index + 1, "tag_name": "other-" + str(index)} for index in range(100)]
        routes = {"releases?per_page=100&page=1": first, "releases?per_page=100&page=2": [selected], "releases/101": copy.deepcopy(selected)}
        self.assertEqual(gate.release_by_tag(Api(routes), tag), selected)
        changed = copy.deepcopy(routes)
        changed["releases?per_page=100&page=1"][0]["tag_name"] = tag
        with self.assertRaisesRegex(ValueError, "Multiple releases"):
            gate.release_by_tag(Api(changed), tag)
        changed = copy.deepcopy(routes)
        changed["releases/101"]["id"] = 102
        with self.assertRaisesRegex(ValueError, "ID/tag differs"):
            gate.release_by_tag(Api(changed), tag)

    def test_draft_wrong_actual_tag_ref_before_any_mutation(self):
        """Reject a draft's different actual tag despite a matching target_commitish before uploads or edits.
        即使 target_commitish 匹配，也在上传或编辑前拒绝草稿的不同实际 tag。
        """
        commit = "a" * 40
        for tag in ("v0.5.9", "examples-v0.5.9"):
            with self.subTest(tag=tag):
                release = {"id": 1, "draft": True, "prerelease": False, "tag_name": tag, "target_commitish": commit, "immutable": False}
                routes = {"git/ref/tags/" + tag: {"ref": "refs/tags/" + tag, "object": {"type": "commit", "sha": "b" * 40}}}
                api = DraftApi(routes, release)
                with patch.dict(gate.os.environ, {"GITHUB_TOKEN": "test-only", "GITHUB_ACTIONS": "true"}), patch.object(gate, "run") as process:
                    with self.assertRaisesRegex(ValueError, "different commit"):
                        gate.publish_assets(api, tag, commit, (), "Test", Path("unused-notes.md"))
                    process.assert_not_called()
                self.assertEqual(api.created, [])
                self.assertTrue(api.release["draft"])

    def test_extra_asset_before_or_during_finalize_never_publishes(self):
        """Reject extra SDK/completion/examples assets on draft or final and immediately before edit, with zero publication.
        拒绝 draft／final 及紧邻 edit 的额外 SDK／完成／示例资产，确保零公开写入。
        """
        commit = "a" * 40
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            file, notes = root / "original.zip", root / "notes.md"
            file.write_bytes(b"original")
            notes.write_text("Original", encoding="utf-8")
            for tag in ("v0.5.9", "examples-v0.5.9", "recovery-v0.5.9-r2-a1"):
                for phase in ("initial-draft", "initial-final", "pre-edit"):
                    with self.subTest(tag=tag, phase=phase):
                        release = {"id": 1, "draft": phase != "initial-final", "prerelease": False,
                                   "tag_name": tag, "target_commitish": commit, "immutable": False}
                        expected = [{"name": file.name, "id": 1}]
                        extra = {"name": "orphan-proof.json", "id": 2}
                        routes = {"git/ref/tags/" + tag: {"ref": "refs/tags/" + tag, "object": {"type": "commit", "sha": commit}},
                                  "git/commits/" + commit: {"sha": commit}, "releases/assets/1": file.read_bytes()}
                        api = DraftApi(routes, release)
                        original_request, asset_reads = api.request, 0

                        def request(path, **kwargs):
                            """Inject an extra asset at the selected real inventory read, preserving exact original bytes.
                            在指定实际清单读取时注入额外资产，保留精确原字节。
                            """
                            nonlocal asset_reads
                            if path == "releases/1/assets?per_page=100&page=1":
                                asset_reads += 1
                                return expected + ([extra] if phase != "pre-edit" or asset_reads >= 2 else [])
                            return original_request(path, **kwargs)

                        with patch.dict(gate.os.environ, {"GITHUB_TOKEN": "fixture", "GITHUB_ACTIONS": "true"}), \
                             patch.object(api, "request", side_effect=request), patch.object(gate, "run") as command:
                            with self.assertRaisesRegex(ValueError, "unexpected assets"):
                                gate.publish_assets(api, tag, commit, (file,), "Test", notes)
                            command.assert_not_called()
                        self.assertEqual(api.created, [])

    def test_optional_tag_does_not_hide_broken_existing_annotation(self):
        """Treat only an absent exact ref as missing; reject an existing annotation whose object returns 404.
        仅将不存在的精确 ref 视为缺失；拒绝对象返回 404 的已有注解。
        """
        tag, commit = "v0.5.9", "a" * 40
        error = urllib.error.HTTPError("exact-ref", 404, "Missing", {}, None)
        with patch.object(Api, "request", side_effect=error):
            self.assertIsNone(gate.optional_tag(Api({}), tag, commit))
        error.close()
        reference = {"ref": "refs/tags/" + tag, "object": {"type": "tag", "sha": "b" * 40}}
        error = urllib.error.HTTPError("annotation", 404, "Missing", {}, None)
        with patch.object(Api, "request", side_effect=[reference, reference, error]):
            with self.assertRaises(urllib.error.HTTPError):
                gate.optional_tag(Api({}), tag, commit)
        error.close()

    def test_toolchain_bootstrap_uses_non_authorizing_public_authority(self):
        """Use fresh public authority's explicit host toolchain and reject formal/cross-host/floating overrides.
        使用新公共权威的显式宿主工具链，拒绝正式、跨宿主或浮动覆盖。
        """
        commit = "a" * 40
        proof = {"phase": "github-only", "complete": False, "registry": None, "core_tag": "v0.5.9", "core_commit": commit}
        inputs = {**proof, "platform": "test-host", "toolchain": "1.2.3"}
        with tempfile.TemporaryDirectory() as temporary:
            for field, value in ((None, None), ("toolchain", "stable"), ("complete", True), ("platform", "foreign"), ("core_commit", "b" * 40)):
                current = copy.deepcopy(inputs)
                if field:
                    current[field] = value
                authority = SimpleNamespace(run_gate=lambda *arguments: proof, toolchain_inputs=lambda *arguments: current)
                args = SimpleNamespace(core_root=Path("unused"), core_tag="v0.5.9", core_commit=commit,
                                       platform="host", output=Path(temporary) / (field or "valid"))
                with patch.object(gate, "core_authority", return_value=authority), patch.object(gate, "host_platform", return_value="test-host"), patch.object(gate, "run") as process:
                    if field:
                        with self.assertRaisesRegex(ValueError, "invalid Rust toolchain"):
                            gate.bootstrap(args)
                        self.assertFalse((args.output / "toolchain.json").exists())
                    else:
                        self.assertEqual(gate.bootstrap(args), inputs)
                        self.assertEqual(gate.read_json(args.output / "toolchain.json"), inputs)
                    process.assert_not_called()


if __name__ == "__main__":
    unittest.main()
