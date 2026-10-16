# Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
"""Group preserved triage evidence without changing tests or spec assessments."""
from collections import Counter, defaultdict
from datetime import datetime
import html
import json
from pathlib import Path
import re
from urllib.parse import quote
from zoneinfo import ZoneInfo

BASE = Path(__file__).resolve().parent
RESULTS = list(map(json.loads, (BASE / 'results.jsonl').read_text().splitlines()))
FAILURES = [c for c in RESULTS if c['Status'] in ('fail', 'timeout')]
ASSESSMENTS = {c['Id']: json.loads((BASE / 'assessments' / (c['Id'] + '.json')).read_text()) for c in FAILURES}
GROUPS = {}


def group(key, title, explanation):
    GROUPS.setdefault(key, {'title': title, 'explanation': explanation})
    return key


def area(name):
    name = name.lower()
    for pattern, label in [
        (r'type-guard|type-test|type-narrow|sem-types|never-type', 'Type tests, narrowing and never'),
        (r'object|class|visibility|access-level|access-test|private|client|resource', 'Objects, classes and visibility'),
        (r'record|struct', 'Records and structural typing'),
        (r'array|tuple|list', 'Arrays, tuples and list values'),
        (r'xml', 'XML'), (r'query|group-by|order-by|join-clause|select-clause', 'Queries'), (r'table', 'Tables'),
        (r'stream', 'Streams'),
        (r'worker|fork|wait|future|async', 'Workers and asynchronous actions'),
        (r'function|lambda|arrow|closure|invocation|argument|param', 'Functions, parameters and closures'),
        (r'match', 'Match patterns'),
        (r'error|panic|check|trap|fail', 'Errors and control flow'),
        (r'decimal|float|integer|numeric|byte|bitwise|arithmetic|add-operation|binary|cast|conversion|division|multiply|subtract|greater|less-than|ref-equal|unary', 'Numbers, operators and conversions'),
        (r'map|mapping|json', 'Maps, JSON and mappings'),
        (r'constant|const-|enum|finite-type', 'Constants, enums and finite types'), (r'import|package|module|top-level', 'Modules and imports'),
        (r'type-definition|typedesc|readonly|freeze', 'Type descriptors and readonly values'),
        (r'isolat|lock', 'Isolation and locks'), (r'annotation', 'Annotations'),
        (r'string|unicode|identifier', 'Strings and identifiers'),
        (r'variable|assign|scope|reference|dataflow|data-flow|var-decl|var-ignore|lvalue|global|forward-referenc', 'Variables, assignment and scope'),
        (r'option-compilation', 'CLI option parameter checks'),
        (r'elvis|ternary|null-value|expression-stmt', 'Conditional, nil and expression statements'),
        (r'while|foreach|if-else|return|break|continue|reachab', 'Statement and reachability checks'),
    ]:
        if re.search(pattern, name):
            return label
    return 'Other language checks'


def slug(value):
    return re.sub(r'[^a-z0-9]+', '-', value.lower()).strip('-')


FATAL_RULES = [
    ('literal', r'transformBasicLiteral|BasicLiteralNode', 'Basic literal lowering'),
    ('binding-list', r'LIST_BINDING_PATTERN', 'List binding patterns'),
    ('binding-map', r'MAPPING_BINDING_PATTERN', 'Mapping binding patterns'),
    ('binding-error', r'ERROR_BINDING_PATTERN', 'Error binding patterns'),
    ('binding-assignment', r'binding-pattern assignment|AssignmentStatementNode', 'Destructuring assignments'),
    ('binding-other', r'binding pattern is not implemented', 'Binding declarations'),
    ('typeof', r'TypeofExpression', 'typeof expressions'),
    ('object-constructor', r'ObjectConstructorExpression', 'Anonymous object constructors'),
    ('table-type', r'TableTypeDescriptor', 'Table type descriptors'),
    ('table-constructor', r'TableConstructorExpression', 'Table constructors'),
    ('worker-declaration', r'named worker declarators', 'Named worker declarations'),
    ('worker-receive', r'ReceiveAction', 'Worker receive actions'),
    ('worker-send', r'AsyncSendAction', 'Worker send actions'),
    ('worker-fork', r'generateForkStatements', 'Fork statements'),
    ('worker-flush', r'FlushAction', 'Worker flush actions'),
    ('query-action', r'QueryAction', 'Query actions'),
    ('query-constructor', r'only map query construct', 'Query constructor kinds'),
    ('query-source', r'query from clause', 'Query collection sources'),
    ('let', r'LetExpression', 'Let expressions'),
    ('rest-argument', r'RestArgument', 'Spread/rest call arguments'),
    ('mapping-spread', r'mapping constructor spread', 'Mapping constructor spread fields'),
    ('start-annotation', r'annotations on start actions', 'Annotations on start actions'),
    ('listener-annotation', r'ListenerDeclaration', 'Annotations on listeners'),
    ('match-typed', r'TYPED_BINDING_PATTERN', 'Typed/capture match patterns'),
    ('match-list', r'LIST_MATCH_PATTERN', 'List match patterns'),
    ('match-map', r'MAPPING_MATCH_PATTERN', 'Mapping match patterns'),
    ('match-error', r'ERROR_MATCH_PATTERN', 'Error match patterns'),
    ('byte-literal', r'ByteArrayLiteral', 'Byte-array literals'),
    ('distinct', r'anonymous distinct', 'Anonymous distinct types'),
    ('configurable', r'configurable module variables', 'Configurable module variables'),
    ('fail', r'FailStatement', 'Fail statements'),
    ('do', r'DoStatement', 'Do statements'),
    ('retry', r'RetryStatement', 'Retry statements'),
    ('lock-on-fail', r'on-fail clause on lock', 'Lock on-fail clauses'),
    ('raw-template', r'raw templates|TemplateExpressionNode', 'Raw templates'),
    ('required-expression', r'RequiredExpression', 'Required expressions'),
    ('regexp', r'Regexp parser', 'Regular-expression parsing'),
    ('numeric-parse', r'failed to parse numeric literal', 'Numeric literal parsing crash'),
    ('object-member', r'unexpected member kind in object', 'Object member lowering crash'),
    ('resource-rest', r'resource access rest', 'Resource path rest segments'),
    ('qualified-field', r'FieldAccessExpression', 'Qualified field access lowering'),
    ('cast-node', r'type cast param node', 'Type-cast descriptor lowering'),
    ('action-expression', r'RemoteMethodCallAction|WaitAction|BracedExpression|StartAction', 'Actions used as expressions'),
    ('natural-expression', r'NaturalExpression', 'Natural-language expressions'),
]


def fatal_group(message):
    for key, pattern, title in FATAL_RULES:
        if re.search(pattern, message):
            return group('lowering-' + key, title,
                         'Shared compiler bailout or lowering signature. A derivative INTERNAL_ERROR can follow an explicit unimplemented diagnostic; those are kept together.')
    return group('fatal-' + slug(message), message, 'Same explicit fatal compiler signature; requires investigation before declaring a common implementation defect.')


def semantic_group(headers):
    meaningful = [h for h in headers if not re.search(r'unused variable|unused import prefix', h)]
    if meaningful:
        headers = meaningful
    joined = '\n'.join(headers)
    methods = re.findall(r'method not found: (\w+)', joined)
    if methods and 'Unknown import: ballerina/test' not in joined:
        return method_group(methods[0])
    for key, pattern, title in [
        ('test-library', r'Unknown import: ballerina/test', 'Migrated assertions still require ballerina/test'),
        ('unused', r'unused variable|unused import prefix', 'Unused-variable/import diagnostics'),
        ('filler', r'no filler value', 'Filler-value analysis and initialization'),
        ('infer-new', r'cannot infer type for implicit new', 'Implicit object constructor type inference'),
        ('container-index', r'unsupported container type for index', 'Indexing container type analysis'),
        ('operator', r'both operands|incompatible types for|binary operator', 'Operator operand compatibility'),
        ('method', r'method not found', 'Missing langlib or object methods'),
        ('name-resolution', r'Unknown symbol|Unknown type|Unknown import|Unknown annotation', 'Symbol, type, import and annotation resolution'),
        ('redeclaration', r'redeclared symbol|Variable already defined|duplicate key', 'Redeclaration and duplicate-member checks'),
        ('reachability', r'[Uu]nreachable|unmatchable', 'Reachability and match exhaustiveness'),
        ('empty-class', r'class definition .* is empty', 'Empty class definitions rejected'),
        ('inclusion', r'type inclusion must', 'Object/class type inclusion validation'),
        ('cyclic-type', r'invalid cycle detected', 'Recursive type cycle validation'),
        ('type-compatibility', r'incompatible type|must be subtype|must be a subtype|included type|constraint', 'Type compatibility and assignability'),
        ('syntax', r'SYNTAX_ERROR', 'Parser acceptance and recovery'),
        ('xml', r'xml|XML', 'XML namespace and attribute analysis'),
    ]:
        if re.search(pattern, joined):
            return group('diagnostics-' + key, title, 'Same diagnostic family blocks the supplied case. The per-case log distinguishes a rejected valid program from extra diagnostics in a negative test.')
    return group('diagnostics-other', 'Other semantic diagnostic mismatches', 'Diagnostic mismatch without a more specific shared signature; retain as separate investigation candidates.')


def method_group(name):
    return group('method-' + slug(name), 'Unresolved method: ' + name,
                 'The captured diagnostics report this specific method as missing. The receiver and supported-library restrictions remain in the independent case assessment; the same spelling can occur on different receiver types.')


def negative_group(c):
    topic = area(c['Name'])
    return group('negative-' + slug(topic), 'Missing expected diagnostics: ' + topic,
                 'One or more @error locations are not covered, or an -e case produced no diagnostics. Cases share a validation area, not a proven single root cause. Any extra diagnostics are listed separately.')


def plain(value):
    if isinstance(value, dict):
        return ' '.join(plain(x) for x in value.values())
    if isinstance(value, list):
        return ' '.join(plain(x) for x in value)
    return str(value)


TOPIC_RULES = [
    ('workers', r'worker|fork|synchronous send|asynchronous send|receive action|flush action', 'Workers, messaging and fork'),
    ('cli', r'command.line|CLI|argument (?:parsing|conversion|binding)|entry.?point|main\s*\(', 'CLI entrypoints and argument conversion'),
    ('langlib', r'lang\.|langlib|lang librar|cloneReadOnly|cloneWithType|error[:.]detail|MAX_VALUE|MIN_VALUE|toCodePoint|codePoint|concat|toJsonString|startsWith|endsWith|removeIfHasKey|removeAll|forEach|\b(?:array|map).{0,20}(?:filter|some)|\b(?:int|array|map|string):|\.(?:detail|sort|entries|forEach|filter|reduce|some|pop|map|sum)\(', 'Langlib functions, methods, types and constants'),
    ('query', r'query|join clause|collect clause|order by|group by', 'Query clauses, sources and constructors'),
    ('table', r'\btable\b|key constraint', 'Tables and key constraints'),
    ('stream', r'\bstream\b', 'Streams'),
    ('match', r'match pattern|match clause|match statement|capture pattern|capture.binding', 'Match patterns'),
    ('typeof', r'typeof', 'typeof expressions'),
    ('object-constructor', r'anonymous object|object.constructor expression', 'Anonymous object constructors'),
    ('binding', r'binding pattern|destructur', 'Binding patterns and destructuring'),
    ('let', r'\blet\b', 'Let expressions'),
    ('fail', r'\bfail\b|on.fail|retry|rollback|transaction|\bdo statement|\bdo\s*\{', 'Fail, on-fail, do, retry and transactions'),
    ('regexp', r'regexp|regular.expression', 'Regular expressions'),
    ('bytes', r'base16|base64|byte.array literal', 'Base16/base64 byte-array literals'),
    ('templates', r'raw template|backtick|raw-template|RawTemplate', 'Raw templates'),
    ('distinct', r'distinct', 'Distinct types and qualifiers'),
    ('configurable', r'configurable|required expression|\? expression', 'Configurable declarations and required expressions'),
    ('annotations', r'annotation|annotat:|deprecat|documentation', 'Annotations, documentation and metadata'),
    ('resources', r'resource|remote|client action', 'Resource and remote access'),
    ('xml', r'\bxml\b|namespace', 'XML forms and qualified names'),
    ('spread', r'spread|rest.argument|rest argument', 'Spread fields and call arguments'),
    ('future', r'future|cancel|wait action', 'Futures, cancellation and wait'),
    ('isolation', r'isolation|isolated|lock statement|locking', 'Isolation and locking'),
    ('qualifiers', r'readonly|public|private|visibility|class.qualif', 'Object/class qualifiers and visibility'),
    ('natural', r'natural.language|natural expression', 'Natural-language expressions'),
    ('elvis', r'Elvis|\?:', 'Elvis expressions'),
    ('recursion', r'stack.exhaustion|stack.overflow|deep recursion', 'Runtime recursion limits and stack exhaustion'),
    ('checking-statement', r'checking statement', 'Standalone checking statements'),
    ('type-descriptor', r'fixed.length array|array type descriptor|tuple.*rest|rest.*tuple', 'Array sizes and tuple rest descriptors'),
    ('record-default', r'record.*default', 'Record field defaults'),
    ('inclusion', r'[Oo]bject.type inclusion', 'Object type inclusion'),
    ('assignment-pattern', r'assignment patterns', 'Destructuring assignment patterns'),
]


def feature_topics(assessment):
    if assessment['classification'] != 'feature':
        return []
    # Use the explicitly unsupported list, not supported constructs in the reason.
    source = plain(assessment.get('unsupported_constructs') or assessment['reason'])
    matches = [key for key, pattern, _ in TOPIC_RULES if re.search(pattern, source, re.I)]
    return matches or ['other']


rows = []
for c in FAILURES:
    cid = c['Id']
    assessment = ASSESSMENTS[cid]
    path = BASE / 'evidence' / cid / 'capture.json'
    capture = json.loads(path.read_text()) if path.exists() else {}
    stderr = capture.get('Stderr', '')
    stdout = capture.get('Stdout', '')
    log = (BASE / 'evidence' / cid / 'run.log').read_text()
    all_headers = re.findall(r'(?m)^(?:fatal|error)\[[^\]]+\]:[^\n]*', stderr)
    headers = list(dict.fromkeys(all_headers))
    fatals = [h.split(': ', 1)[1] for h in headers if h.startswith('fatal[')]
    # Suppress the generic binding bailout when its concrete subtype is present.
    if any(re.search(r'(LIST|MAPPING|ERROR)_BINDING_PATTERN', x) for x in fatals):
        fatals = [x for x in fatals if x != 'binding pattern is not implemented']
    related = list(dict.fromkeys(fatal_group(x) for x in fatals))
    missing = len(re.findall(r'@error annotation at .* is not covered by any diagnostic', log))
    extra = len(re.findall(r'diagnostic at .* not covered by any @error annotation', log))
    extra_locations = set(re.findall(r'diagnostic at (.*?):(\d+)-(\d+) not covered by any @error annotation', log))
    resolved = capture.get('Diagnostics') or []
    unexpected_headers = []
    if extra_locations:
        assert len(resolved) == len(all_headers), (cid, 'cannot map extra diagnostics to messages')
        unexpected_headers = [h for h, d in zip(all_headers, resolved)
                              if (d['File'], str(d['StartLine']), str(d['EndLine'])) in extra_locations]
        assert unexpected_headers, (cid, 'extra diagnostic location not found')
    kind = c['Name'].removesuffix('.bal').rsplit('-', 1)[1]
    mode = ''
    if c['Status'] == 'timeout':
        topic = 'regexp' if 'regexp' in c['Name'].lower() else 'table'
        primary = group('timeout-' + topic, 'Timeouts: ' + topic, 'Same test area exceeded the execution bound. No common infinite loop or root cause has been established.')
        mode = 'process timeout'
    elif not capture:
        primary = group('crash-stack-overflow', 'Go stack overflow during type resolution', 'Both subprocesses exhausted the Go stack; recurrence in type resolution is visible in their logs. This is a crash family, not a confirmed duplicate defect.')
        mode = 'Go process crash'
    elif fatals:
        # Prefer the explicit unsupported operation to its derivative INTERNAL_ERROR.
        explicit = [x.split(': ', 1)[1] for x in headers if x.startswith('fatal[UNIMPLEMENTED_ERROR]')]
        explicit = [x for x in explicit if x in fatals]
        primary = fatal_group((explicit or fatals)[0])
        mode = 'fatal compiler diagnostic'
    elif c['Args'] is not None:
        if headers:
            primary = semantic_group(headers)
            mode = 'CLI compilation failure'
        else:
            primary = group('cli-arguments', 'CLI argument acceptance, rejection and conversion', 'Cases invoked through @args fail the CLI exit/output/panic contract; argument conversion and the corpus extension need investigation.')
            mode = 'CLI behavior mismatch'
    elif kind == 'e' and (missing or 'has empty expected stderr' in log):
        primary = negative_group(c)
        mode = 'missing expected diagnostics' + (' and extra diagnostics' if extra else '')
    elif kind == 'e' and extra:
        primary = semantic_group(unexpected_headers)
        mode = 'extra diagnostics at unmarked locations'
    elif headers:
        primary = semantic_group(headers)
        mode = 'compilation diagnostics'
    elif 'panic:' in stdout:
        primary = group('crash-go-panic', 'Recovered Go panic during compilation or execution', 'A Go panic escaped an internal operation and was recovered by the corpus harness. The group is split by language area in the case table; it does not imply one nil dereference site.')
        mode = 'Go panic'
    elif 'no filler value' in stderr:
        primary = group('runtime-filler', 'Unexpected no-filler-value runtime failures', 'Initialization failed before the intended output or panic site. Stack-location mismatches in these cases may be a consequence of that earlier runtime failure.')
        mode = 'unexpected runtime filler failure'
    elif 'top stack frame mismatch for @panic' in log:
        primary = group('panic-location', 'Panic source-location or top-frame mismatch', 'The panic validator reports a different source location from @panic. Some fixtures may have stale locations; shared symptom alone does not prove a runtime attribution bug.')
        mode = 'panic frame/location mismatch'
    elif 'expected non-zero exit code, got 0' in log or (kind == 'p' and 'has empty expected stderr' in log):
        primary = group('panic-missing', 'Expected runtime panic did not occur', 'A -p case completed without the required runtime failure.')
        mode = 'missing expected panic'
    elif stderr.startswith('error:'):
        title = 'Runtime assertion failures' if re.search(r'expected .*found|assert', stderr, re.I) else 'Other unexpected runtime errors'
        primary = group('runtime-' + slug(title), title, 'Execution raised a Ballerina runtime error; compare the original assertion and captured stack before combining implementation fixes.')
        mode = 'runtime error or panic mismatch'
    elif '@output annotation mismatch' in log:
        topic = area(c['Name'])
        primary = group('output-' + slug(topic), 'Output mismatch: ' + topic, 'Ordered @output markers differ from execution. These are behavior/formatting candidates grouped by test area; comparison of values is needed to establish duplicates.')
        mode = 'output mismatch'
    else:
        primary = group('validation-other', 'Other marker or process-contract failures', 'Remaining validator/process failure; see the exact validation log before combining.')
        mode = 'other validation failure'
    if primary not in related:
        related.insert(0, primary)
    if headers and any('unused variable' in h for h in headers):
        related.append(group('diagnostics-unused', 'Unused-variable/import diagnostics', 'Same diagnostic family blocks the supplied case. The per-case log distinguishes a rejected valid program from extra diagnostics in a negative test.'))
    if 'Unknown import: ballerina/test' in stderr:
        related.append(semantic_group(['error[SEMANTIC_ERROR]: Unknown import: ballerina/test']))
    related += [method_group(name) for name in dict.fromkeys(re.findall(r'method not found: (\w+)', stderr))]
    rows.append({'id': cid, 'name': c['Name'], 'classification': assessment['classification'],
                 'primary_group': primary, 'related_groups': list(dict.fromkeys(related)),
                 'feature_topics': feature_topics(assessment), 'area': area(c['Name']),
                 'validation_mode': mode, 'missing_error_markers': missing,
                 'extra_diagnostics': extra, 'diagnostic_signatures': headers,
                 'unexpected_diagnostic_signatures': unexpected_headers,
                 'unsupported_constructs': assessment.get('unsupported_constructs', [])})


def counts(items):
    return Counter(r['classification'] for r in items)


primary = defaultdict(list)
related = defaultdict(list)
topics = defaultdict(list)
for row in rows:
    primary[row['primary_group']].append(row)
    for key in row['related_groups']:
        related[key].append(row)
    for key in row['feature_topics']:
        topics[key].append(row)
assert len(rows) == 1178 and len({r['id'] for r in rows}) == 1178
assert counts(rows) == {'bug': 404, 'feature': 767, 'uncertain': 7}
assert sum(map(len, primary.values())) == 1178
assert {r['id'] for members in topics.values() for r in members} == {r['id'] for r in rows if r['classification'] == 'feature'}


def cell(value):
    return html.escape(plain(value), quote=False).replace('|', '\\|').replace('\n', ' ')


def case_link(row):
    return f'[{row["id"]}](#{row["id"].lower()})'


def source_links(row):
    name = row.get('name', row.get('Name'))
    kind = 'blob' if name.endswith('.bal') else 'tree'
    remote = ('https://github.com/heshanpadmasiri/ballerina-lang/' + kind +
              '/87c44c943d23ad9e99d61957d1e950610dd40f61/migration-results/' + quote(name, safe='/'))
    return f'[{name}]({remote}) · [local source](/workspace/scratch/jballerina-migration/migration-results/{name})'


# Embed the full findings so readers never need a second report for a case.
report = (BASE / 'REPORT.md').read_text()
scope = '## Scope and revisions\n' + report.split('## Scope and revisions\n', 1)[1].split('## Failure inventory\n', 1)[0]
individual = report.split('## Individual failure assessments\n', 1)[1].split('## Java exclusions\n', 1)[0]
parts = re.split(r'(?m)^### (T\d{4})\n', individual)
case_details = {parts[i]: parts[i + 1].strip() for i in range(1, len(parts), 2)}
case_details = {cid: body.replace('](#expressions)', f']({BASE.parents[1]}/doc/lang/subset10.md#expressions)')
                for cid, body in case_details.items()}
assert set(case_details) == {r['id'] for r in rows}


ordered = sorted(primary, key=lambda k: (-counts(primary[k])['bug'], -len(primary[k]), GROUPS[k]['title']))
group_link = lambda k: f'[{cell(GROUPS[k]["title"])}](#{k})'
lines = ['# Grouped migrated-test failures', '',
         f'Generated {datetime.now(ZoneInfo("Asia/Colombo")).isoformat(timespec="seconds")}.', '',
         'This single report contains the grouping, full per-case findings, spec references, and direct links to each original migrated test. Case IDs jump to findings in this document; test names link to the pinned source revision.', '',
         '**Coverage: all 1,178 failures/timeouts — 404 bug, 767 feature, 7 uncertain.** The 382 passes and 78 Java exclusions also have test references below.', '',
         scope,
         '## How to use these groups', '',
         '- Every failure has exactly one primary group below. Primary counts add up to 1,178; bug/feature/uncertain labels are unchanged from the clean reviews.',
         '- Groups indicate similar observed failures or a common investigation area. They are candidates for shared fixes, not a claim that each group is one confirmed root cause or one issue.',
         '- Explicit compiler bailouts take precedence over derivative INTERNAL_ERROR messages. Otherwise grouping uses the failing marker/process check. For negative tests, expected diagnostics are not treated as implementation failures merely because they appear in stderr.',
         '- Missing expected diagnostics are grouped by language area. Extra diagnostics and their counts remain visible in each row. A case may have both, so a missing-diagnostic case is not automatically classified as an unused-variable defect.',
         '- With multiple compiler bailouts, the first explicit unimplemented operation is the primary group; all observed fatal families appear in the related-signature index. Fixing the primary failure may expose another blocker.',
         '- The feature-gap index is separate: it groups reviewer-listed unsupported constructs. It can differ from the observed failure family. A feature classification does not make an internal crash acceptable.',
         '- No tests were rerun or modified and no GitHub issues were created for this grouping.', '',
         '## Primary group summary', '',
         '| Failure group | Total | Bug | Feature | Uncertain |', '| --- | ---: | ---: | ---: | ---: |']
for key in ordered:
    cc = counts(primary[key])
    lines.append(f'| {group_link(key)} | {len(primary[key])} | {cc["bug"]} | {cc["feature"]} | {cc["uncertain"]} |')
lines += ['| **Total** | **1178** | **404** | **767** | **7** |', '',
          '## Primary groups and every affected case', '']
for key in ordered:
    members = primary[key]
    cc = counts(members)
    lines += [f'<a id="{key}"></a>', '', f'### {GROUPS[key]["title"]}', '',
              f'**{len(members)} cases: {cc["bug"]} bug, {cc["feature"]} feature, {cc["uncertain"]} uncertain.**', '',
              GROUPS[key]['explanation'], '',
              '| Case | Assessment | Test | Failing check | Missing @error locations | Extra diagnostics | Other observed groups |',
              '| --- | --- | --- | --- | ---: | ---: | --- |']
    for row in sorted(members, key=lambda r: ({'bug': 0, 'feature': 1, 'uncertain': 2}[r['classification']], r['id'])):
        other = '; '.join(group_link(k) if k in primary else f'[{cell(GROUPS[k]["title"])}](#related-signatures)' for k in row['related_groups'] if k != key)
        lines.append(f'| {case_link(row)} | {row["classification"]} | `{row["name"]}` | {cell(row["validation_mode"])} | {row["missing_error_markers"]} | {row["extra_diagnostics"]} | {other} |')
        lines[-1] = lines[-1].replace('`' + row['name'] + '`', source_links(row))
    for row in sorted(members, key=lambda r: ({'bug': 0, 'feature': 1, 'uncertain': 2}[r['classification']], r['id'])):
        lines += ['', f'#### {row["id"]}', '', f'[Back to {cell(GROUPS[key]["title"])}](#{key})', '',
                  case_details[row['id']], '']
lines += ['', '## Related signatures', '',
          'This index includes secondary compiler bailouts and unused-variable diagnostics present in the captured output. Counts overlap and are not additional failures. A diagnostic may be expected in an -e test; consult its failing check in the primary group before treating that diagnostic as a defect.', '',
          '| Signature family | Primary cases | All cases showing family | Cases |', '| --- | ---: | ---: | --- |']
for key in sorted(related, key=lambda k: (-len(related[k]), GROUPS[k]['title'])):
    if len(related[key]) == len(primary.get(key, [])):
        continue
    lines.append(f'| {cell(GROUPS[key]["title"])} | {len(primary.get(key, []))} | {len(related[key])} | '+', '.join(case_link(r) for r in related[key])+' |')
lines += ['', '## Feature-gap index', '',
          'These topics come from the clean reviewers’ unsupported-construct lists (their reason when no list was supplied). Topic matching is an index, not a new spec judgment. Cases can appear under several topics; all 767 feature cases appear at least once. The full reviewer reasoning remains linked through each case.', '',
          '| Feature topic | Cases |', '| --- | ---: |']
topic_titles = {key: title for key, _, title in TOPIC_RULES}
topic_titles['other'] = 'Other explicitly cited spec gaps'
for key in sorted(topics, key=lambda k: -len(topics[k])):
    lines.append(f'| [{topic_titles[key]}](#feature-{key}) | {len(topics[key])} |')
for key in sorted(topics, key=lambda k: -len(topics[k])):
    lines += ['', f'<a id="feature-{key}"></a>', '', '### ' + topic_titles[key], '',
              ', '.join(case_link(r) for r in topics[key]) + '.']
    if key == 'other':
        lines += ['', '| Case | Reviewer-listed unsupported constructs |', '| --- | --- |']
        for row in topics[key]:
            lines.append(f'| {case_link(row)} | {cell(row["unsupported_constructs"] or ASSESSMENTS[row["id"]]["reason"])} |')
lines += ['', '## Spec ambiguities', '',
          '| Shared ambiguity | Cases |', '| --- | --- |',
          '| readonly class: supported qualifier list conflicts with object/class restrictions | ' + ', '.join(case_link(r) for r in rows if r['classification'] == 'uncertain' and r['id'] != 'T1619') + ' |',
          '| XML-qualified attribute field names: field-access support versus qualified-name restriction | ' + ', '.join(case_link(r) for r in rows if r['id'] == 'T1619') + ' |', '',
          '## Java exclusions', '',
          '| Case | Test references | Dependency evidence |', '| --- | --- | --- |']
for case in RESULTS:
    if case['Status'] == 'excluded-java':
        lines.append(f'| {case["Id"]} | {source_links(case)} | {cell(case["JavaEvidence"])} |')
lines += ['', '## Passed cases', '', '| Case | Test references | Seconds |', '| --- | --- | ---: |']
for case in RESULTS:
    if case['Status'] == 'pass':
        lines.append(f'| {case["Id"]} | {source_links(case)} | {case["Seconds"]} |')
lines += ['', '## Machine-readable grouping and reproduction', '',
          f'- [Case-to-group mapping]({BASE}/failure-groups.json) includes every primary/related group, raw diagnostic signature, feature topic, and marker mismatch count.',
          f'- [Grouping script]({BASE}/group_failures.py) reads the preserved execution evidence and assessment JSON; run `python3 triage/migrated-tests/group_failures.py` from the Nutcracker checkout to regenerate this document.', '']
(BASE / 'GROUPED_FAILURES.md').write_text('\n'.join(lines))
(BASE / 'failure-groups.json').write_text(json.dumps({'method': 'Observed signature/validator groups; shared root causes unconfirmed', 'groups': GROUPS, 'feature_topics': topic_titles, 'cases': rows}, indent=2) + '\n')
print(json.dumps({'cases': len(rows), 'primary_groups': len(primary), 'feature_topics': len(topics), 'classifications': dict(counts(rows)), 'unmatched_feature_topics': len(topics['other']), 'primary_group_counts': {k: len(primary[k]) for k in ordered}}, indent=2))
