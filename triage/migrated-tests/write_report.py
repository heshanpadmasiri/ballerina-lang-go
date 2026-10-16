# Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
"""Assemble execution evidence and independent spec assessments into Markdown."""
from collections import Counter
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re
import sys
from zoneinfo import ZoneInfo

BASE = Path(__file__).resolve().parent
ROOT = BASE.parents[1]
RESULTS = [json.loads(line) for line in (BASE/'results.jsonl').read_text().splitlines()]
inventory=json.loads((BASE/'inventory.json').read_text())
assert {c['Id'] for c in RESULTS} == {c['Id'] for c in inventory}, 'Execution inventory incomplete'
assert len(RESULTS) == len(inventory), 'Execution records missing or duplicated'
REV = '87c44c943d23ad9e99d61957d1e950610dd40f61'
ENGINE = '055c54a696a377438d632b1dbffead994fa5a223'

def text(value):
    if isinstance(value,dict):
        return '; '.join(f'{k}: {text(v)}' for k,v in value.items())
    if isinstance(value,list):
        return '; '.join(text(v) for v in value)
    return str(value)

def cell(value):
    return text(value).replace('|','\\|').replace('\n',' ').strip()

def load_capture(case):
    p=BASE/'evidence'/case['Id']/'capture.json'
    return json.loads(p.read_text()) if p.exists() else {}

def observed(case):
    cap=load_capture(case)
    stderr=cap.get('Stderr','')
    if case['Status']=='timeout':return 'Execution exceeded 25-second Go test timeout (30-second outer watchdog).'
    if not cap:
        return 'Process aborted with Go stack overflow; see run log.'
    headers=re.findall(r'(?m)^(?:fatal|error)\[[^\]]+\]:[^\n]*',stderr)
    if headers:return '; '.join(dict.fromkeys(headers))[:950]
    if stderr.strip():return stderr.strip().splitlines()[0][:700]
    stdout=cap.get('Stdout','')
    if 'panic:' in stdout:return next(line for line in stdout.splitlines() if 'panic:' in line)[:700]
    log=(BASE/'evidence'/case['Id']/'run.log').read_text()
    match=re.search(r'migration_triage_test.go:\d+: (.*)',log)
    return match.group(1) if match else 'See captured output and marker validation log.'

def evidence_kind(case):
    cap=load_capture(case);s=cap.get('Stderr','');o=cap.get('Stdout','')
    if case['Status']=='timeout':return 'Timeout'
    if not cap:return 'Go process crash'
    if 'INTERNAL_ERROR' in s:return 'Internal compiler error'
    if 'UNIMPLEMENTED_ERROR' in s:return 'Explicitly unimplemented'
    if 'unused variable' in s:return 'Unused-variable diagnostics'
    if s.startswith('error['):return 'Other compile diagnostics'
    if 'panic:' in o:return 'Go panic recovered by corpus harness'
    if 'error:' in s:return 'Runtime error or panic-marker mismatch'
    return 'Output or marker mismatch'

assessments={}
for p in (BASE/'assessments').glob('*.json'):
    a=json.loads(p.read_text());assert a['id']==p.stem
    assert a['classification'] in ('bug','feature','uncertain'),(p,a['classification'])
    assessments[p.stem]=a
failures=[c for c in RESULTS if c['Status'] in ('fail','timeout')]
pending=[c for c in failures if c['Id'] not in assessments]
if '--complete' in sys.argv:
    assert not pending, f'{len(pending)} independent assessments missing'
counts=Counter(c['Status'] for c in RESULTS)
kind_table=['### Results by corpus case kind', '', '| Kind | Discovered | Java excluded | Passed | Failed | Timed out |', '| --- | ---: | ---: | ---: | ---: | ---: |']
for kind,label in [('v','Valid (-v)'),('e','Expected compile error (-e)'),('p','Expected runtime panic (-p)')]:
    group=[c for c in RESULTS if c['Name'].removesuffix('.bal').rsplit('-',1)[1]==kind]
    kc=Counter(c['Status'] for c in group)
    kind_table.append(f'| {label} | {len(group)} | {kc["excluded-java"]} | {kc["pass"]} | {kc["fail"]} | {kc["timeout"]} |')
kind_table.append('')

classes=Counter(assessments[c['Id']]['classification'] if c['Id'] in assessments else 'pending' for c in failures)
lines=['# Migrated jBallerina tests: Nutcracker triage','',
       f'Generated {datetime.now(ZoneInfo("Asia/Colombo")).isoformat(timespec="seconds")}.', '',
       '**Status: '+('COMPLETE' if not pending else f'IN PROGRESS — {len(pending)} independent reviews still pending')+'**','',
       '[Grouped failures and feature gaps](GROUPED_FAILURES.md) — every failure grouped by observed signature or validation area, with primary counts and related symptoms.', '',
       '## Scope and revisions','',
       f'- Migrated suite: `origin/feat/test-migration-results` in `/workspace/ballerina-lang`, commit `{REV}`, directory `migration-results/` (487 candidate groups).',
       f'- Nutcracker: `/workspace/ballerina-lang-go`, commit `{ENGINE}`; Go 1.27.1, Linux.',
       f'- Supported-language reference: [subset10.md](../../doc/lang/subset10.md), SHA-256 `{hashlib.sha256((ROOT/"doc/lang/subset10.md").read_bytes()).hexdigest()}`.',
       '- The migration branch was read using `git archive`; neither checkout was switched to another branch. No GitHub issues were created. Implementation and migrated test sources were not changed.', '',
       '## Results','',
       '| Outcome | Cases |','| --- | ---: |',
       f'| Discovered | {len(RESULTS)} |',
       f'| Excluded: explicit Java dependency | {counts["excluded-java"]} |',
       f'| Executed | {len(RESULTS)-counts["excluded-java"]} |',
       f'| Passed | {counts["pass"]} |',
       f'| Failed marker, compilation, runtime, or process checks | {counts["fail"]} |',
       f'| Timed out | {counts["timeout"]} |',
       f'| Failure assessed as bug | {classes["bug"]} |',
       f'| Failure assessed as feature | {classes["feature"]} |',
       f'| Failure with ambiguous spec coverage | {classes["uncertain"]} |',
       f'| Failure awaiting independent review | {classes["pending"]} |','',
       'Counts are test cases, not unique implementation defects or proposed issues. One project is one case; its module files are not counted separately. A case may contain several independent assertions and may mix supported and unsupported constructs.', '',
       *kind_table, '## Method','',
       '1. Enumerated every migrated standalone suffix-marked `.bal` file and package project. Workspace manifests and peer package sources were retained as dependency context. An enclosing package manifest governs a standalone-looking file where present.',
       '2. Excluded a case if its source or governing/dependency manifest explicitly imports Java interop, uses `@java:`, or declares a Java platform/JAR dependency. Exclusion evidence is listed below. The fact that a test originated in a Java test harness is not an exclusion.',
       '3. Used the existing `test_util/testharness` compilation/runtime pipeline and its original marker validators, without the corpus skip lists. Added an explicit guard rejecting fatal compiler diagnostics for every case and compile-error diagnostics for valid/panic cases; an expected runtime panic requires successful compilation. `@output` is exact and ordered; `@error` requires bidirectional inclusive-span diagnostic coverage; `@panic` uses the existing corpus panic validator. Suffix invariants are checked. No Nutcracker-generated output was accepted as a new expected golden.',
       '4. Each case ran in its own subprocess, with six subprocesses in parallel, a 25-second Go test timeout, and a 30-second process-group watchdog. Full stdout, stderr, structured diagnostics, exit status, validation log and timings were saved.',
       '5. For every failure or timeout, a new subagent was launched with no inherited conversation. It received only that individual test case (including supplied project/dependency sources), the subset10 specification, and the assessment/output instructions. It did not receive execution logs, compiler implementation, migration review notes, or another reviewer’s assessment. Coordinators only assigned cases and collected results.',
       '6. Spec-supported cases are labelled **bug**; cases requiring unsupported/not-covered functionality are labelled **feature**. **Uncertain** preserves ambiguity instead of inventing a spec commitment. These are whole-case scope assessments; the observed evidence alongside each assessment identifies what actually failed.', '',
       '## Test-design and execution limits','',
       '- Corpus output instrumentation (`ballerina/io` and `io:println`) is treated as the test harness. Its absence from the restricted `ballerina/lang.*` library enumeration is not, by itself, a language feature gap.',
       '- The migrated fixtures already use the corpus `-v`/`-e`/`-p` suffixes and `@output`/`@error`/`@panic` design. Sources and markers were preserved. Existing migration `review.md` files were not used as expectations or reviewer context.',
       f'- {sum(c["Args"] is not None and c["Status"]!="excluded-java" for c in RESULTS)} eligible cases also use `@args`, an extension absent from the corpus marker contract. Those cases were attempted through the Nutcracker CLI with their exact argument arrays, and checked with the corpus output/panic validators. CLI entrypoint/argument support and the test design must be considered before treating those failures as language defects.',
       '- As in corpus, `@error` comment text is explanatory; diagnostic message wording is not compared. A pass confirms the marker contract, not equivalence of every diagnostic message. Source-integrity verification checked all 2,216 original migration files and found no changes.',
       '- The corpus harness includes its normal graceful-stop behavior for listeners. Its panic validator permits a runtime `error:` without a Ballerina stack frame; passing panic cases therefore have exactly the same verification limit as the existing corpus.',
       '- Eight workspace package cases use isolated copies in the corpus workspace layout: the intended target is listed first and only its workspace dependency closure is included. All package source files remain identical; only the copied workspace entrypoint manifest is adapted to the corpus first-package convention. No dependency was fetched, prepublished, or replaced with a stub. Independent reviewers receive only that target and its dependency sources.',
       '- The original harness gap is in [suffix checks](../../test_util/testharness/test_harness.go:692) and the [panic validator](../../test_util/testharness/test_harness.go:1030): fatal diagnostics were not consistently rejected before marker validation. The temporary adapter rejects them without changing compiler/runtime implementation.',
       '- A validation audit found 107 cases that the original marker checks allowed to pass despite fatal compilation diagnostics. The explicit compilation guard corrected those false passes. Correct workspace selection resolved two earlier failures. The final execution results below come from the complete corrected rerun; original evidence is preserved for comparison.',
       '- Timeout is a bounded-run observation, not proof that a program can never terminate. Go process crashes and recovered Go panics are distinguished from expected Ballerina runtime panics.',
       '- A feature label describes lack of whole-case coverage under subset10; it does not excuse an internal compiler error or prove that every construct in that file is unsupported. Mixed cases should be split before converting these findings into narrowly scoped bug reports.', '',
       '## Observed failure symptoms','', '| First applicable evidence category | Cases | Bug | Feature | Uncertain | Pending |','| --- | ---: | ---: | ---: | ---: | ---: |']
for kind,n in Counter(evidence_kind(c) for c in failures).most_common():
    scope=Counter(assessments[c['Id']]['classification'] if c['Id'] in assessments else 'pending' for c in failures if evidence_kind(c)==kind)
    lines.append(f'| {kind} | {n} | {scope["bug"]} | {scope["feature"]} | {scope["uncertain"]} | {scope["pending"]} |')
lines += ['', 'These symptom groups are independent of the clean reviewers’ bug/feature decisions.', '',
          '## Migration test-design blockers', '',
          'The fixtures were executed as supplied. The following cases produced `Unknown import: ballerina/test`; their assertions still depend on a test library beyond the corpus marker design. The clean spec assessment is preserved, but this evidence alone does not establish a defect in the language operations under test. These fixtures need corpus-style assertion migration before a language failure can be isolated. Other failures may mask the same import in additional cases.', '',
          '| Case | Clean spec assessment | Test |', '| --- | --- | --- |']
for c in failures:
    if 'Unknown import: ballerina/test' in load_capture(c).get('Stderr',''):
        a=assessments.get(c['Id'])
        lines.append(f'| [{c["Id"]}](#{c["Id"].lower()}) | {a["classification"] if a else "pending"} | `{cell(c["Name"])}` |')
lines += ['',
          '## Failure inventory','', '| Case | Spec assessment | Test | Observed symptom |','| --- | --- | --- | --- |']
for c in sorted(failures, key=lambda c: ({'bug':0,'feature':1,'uncertain':2,'pending':3}[assessments[c['Id']]['classification'] if c['Id'] in assessments else 'pending'], c['Id'])):
    a=assessments.get(c['Id']);classification=a['classification'] if a else 'pending'
    lines.append(f'| [{c["Id"]}](#{c["Id"].lower()}) | {classification} | `{cell(c["Name"])}` | {cell(observed(c))} |')
lines += ['', '## Individual failure assessments','']
for c in failures:
    a=assessments.get(c['Id']);cid=c['Id']
    lines += [f'### {cid}', '',f'**Test:** `{c["Name"]}`  ',f'**Execution:** {c["Status"]}; exit `{c["ExitCode"]}`; {c["Seconds"]} seconds.  ',f'**Spec assessment:** {a["classification"] if a else "pending"}.', '',f'**Observed:** {observed(c)}','',f'[Execution log](evidence/{cid}/run.log)']
    if (BASE/'evidence'/cid/'capture.json').exists():lines[-1]+=f' · [Captured stdout, stderr and diagnostics](evidence/{cid}/capture.json)'
    if a:
        lines[-1]+=f' · [Independent assessment JSON](assessments/{cid}.json)'
        lines += ['',text(a['reason'])]
        if a.get('spec_references'):
            lines += ['', '**Spec references:**','']
            for ref in a['spec_references']:
                if isinstance(ref,dict):lines.append(f'- {ref.get("heading", "Spec")}: {ref.get("quote",text(ref))}')
                else:lines.append(f'- {text(ref)}')
        for title,key in [('Unsupported constructs','unsupported_constructs'),('Test-design notes','test_design_notes')]:
            if a.get(key):
                lines += ['',f'**{title}:**','']
                vals=a[key] if isinstance(a[key],list) else [a[key]]
                lines += ['- '+text(item) for item in vals]
    rel=c['Name']
    lines += ['',f'[Extracted case](/workspace/scratch/jballerina-migration/migration-results/{rel}) · [Original migrated case](https://github.com/heshanpadmasiri/ballerina-lang/{"blob" if rel.endswith(".bal") else "tree"}/{REV}/migration-results/{rel})','']
lines += ['## Java exclusions','', '| Case | Test | Dependency evidence |','| --- | --- | --- |']
for c in RESULTS:
    if c['Status']=='excluded-java':
        lines.append(f'| {c["Id"]} | `{cell(c["Name"])}` | {cell(c["JavaEvidence"])} |')
lines += ['', '## Passed cases','', '| Case | Test | Seconds |','| --- | --- | ---: |']
for c in RESULTS:
    if c['Status']=='pass':lines.append(f'| {c["Id"]} | `{cell(c["Name"])}` | {c["Seconds"]} |')
lines += ['', '## Evidence and reproduction','',
          '- [Pinned subset10 reference](subset10.snapshot.md).',
          '- [Execution completeness and source audit](execution-audit.json).',
          '- [Validation audit](validation-audit.json), [initial run results](initial-results.jsonl), and original execution evidence under `initial-evidence/` document the fatal-diagnostic and workspace setup corrections.',
          '- [Individual reviewer provenance](reviewer-provenance.jsonl) and [initial coordinator provenance](initial-coordinator-provenance.json). Fresh QA re-reviews retain their original assessments outside the active assessment directory.',
          '- [Complete discovery inventory](inventory.json), [execution results](results.jsonl), [source-integrity check](source-integrity.json), and [runner](run_triage.py).',
          '- [Temporary corpus harness adapter](migration_triage_test.go.txt). This adapter calls the existing invariant and marker validators; it does not change compiler/runtime behavior.',
          '- Evidence directories are keyed by the stable `T####` IDs above. Each executed case has `result.json` and `run.log`; cases that returned normally also have `capture.json`.',
          '- The source extraction is `/workspace/scratch/jballerina-migration/migration-results`. The pinned source branch and commits above allow it to be recreated.', '',
          '```bash',
          'cd /workspace/ballerina-lang-go',
          'source /workspace/.onboarding/activate.sh',
          'cp triage/migrated-tests/migration_triage_test.go.txt test_util/testharness/migration_triage_test.go',
          'go test -c ./test_util/testharness -o /workspace/scratch/migration-triage.test',
          'python3 triage/migrated-tests/run_triage.py',
          'rm test_util/testharness/migration_triage_test.go',
          'python3 triage/migrated-tests/write_report.py --complete',
          'python3 triage/migrated-tests/group_failures.py',
          '```', '',
          'The clean independent assessments are preserved as JSON. Re-running the execution command does not regenerate those assessments; the report combines the existing assessments with the rerun evidence.','']
document='\n'.join(lines)
# A Ballerina function type can contain `](...)`; code spans are not links.
parts=re.split(r'(```[\s\S]*?```|`[^`\n]*`)', document)
def local_link(match):
    target=match.group(1)
    path=(BASE/re.sub(r':\d+$','',target)).resolve()
    if path.exists():
        return ']('+str((BASE/target).resolve())+')'
    # Unquoted function signatures in reviewer prose can resemble Markdown.
    return '\\]('+target+')'
for i in range(0,len(parts),2):
    parts[i]=re.sub(r'\]\((?!https?://|#|/)([^)\n]+)\)', local_link, parts[i])
document=''.join(parts)
(BASE/'REPORT.md').write_text(document)
print(json.dumps({'cases':len(RESULTS),'counts':dict(counts),'assessments':dict(classes),'pending':len(pending),'report_bytes':(BASE/'REPORT.md').stat().st_size}))
